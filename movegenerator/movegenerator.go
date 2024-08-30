package movegenerator

import (
	"endtner.dev/nChess/board"
	"endtner.dev/nChess/board/boardhelper"
	"endtner.dev/nChess/board/move"
	"endtner.dev/nChess/board/piece"
	"math/bits"
)

/*
	MoveGenerator is now responsible for generating all Bitboards. THIS CODE DOES NOT RUN IN PARALLEL!

	When parallelized, other generators will override these variables for a given generator, so everything crashes.
	I tried instancing the generator per class, that created so much overhead that the performance increase was mitigated.
*/

// Occupancy & Attacks & Pins can be accessed at index (3 >> color) - 1 & 1 - accessIndex for enemy color
var Occupancy = make([]uint64, 3)

func ComputeAll(b *board.Board) {
	var whitePieces uint64
	var blackPieces uint64

	for p, bitboard := range b.Bitboards {
		if uint8(p)&piece.White == piece.White {
			whitePieces |= bitboard
		}
		if uint8(p)&piece.Black == piece.Black {
			blackPieces |= bitboard
		}
	}

	Occupancy[0] = whitePieces
	Occupancy[1] = blackPieces
	Occupancy[2] = whitePieces | blackPieces

	ComputeAttacks(b)
	ComputePins(b)
}

/*
	Generators
*/

func LegalMoves(b *board.Board) []move.Move {
	/*
		Creating pseudo-legal moves in an iterative approach, then filtering out illegal moves
	*/

	// Precompute occupancy, enemy attacks & own pins
	ComputeAll(b)

	pseudoLegalMoves := make([]move.Move, 0, 218) // Maximum possible moves in a chess position is 218
	KingMoves(b, &pseudoLegalMoves)

	friendlyKingIndex := b.FriendlyKingIndex

	opponentAttackMask := Attacks[b.OpponentIndex]
	friendlyPinMask := Pins[b.FriendlyIndex]

	checkCount := 0
	protectMoveMask := ^uint64(0)

	if boardhelper.IsIndexBitSet(friendlyKingIndex, opponentAttackMask) {
		checkCount, protectMoveMask = CheckMask(b)
	}

	// Can directly return king moves if we are in multi check
	if checkCount > 1 {
		return pseudoLegalMoves
	}

	PawnMoves(b, &pseudoLegalMoves)
	SlidingMoves(b, &pseudoLegalMoves)
	KnightMoves(b, &pseudoLegalMoves)

	legalMoves := make([]move.Move, 0, 218) // Maximum possible moves in a chess position is 218

	for _, m := range pseudoLegalMoves {

		// Only move along pin ray if piece is pinned
		if boardhelper.IsIndexBitSet(m.StartIndex, friendlyPinMask) && !boardhelper.IsIndexBitSet(m.TargetIndex, calculatePinRay(b, m.StartIndex)) {
			continue
		}

		// King is in single check
		if checkCount == 1 {

			// If a move is not to any of the protect square OR not a king move
			if !boardhelper.IsIndexBitSet(m.TargetIndex, protectMoveMask) && (friendlyKingIndex != m.StartIndex) {

				// Move can not enPassantCapture the checking pawn, not allowed
				if m.EnPassantCaptureSquare == -1 {
					continue
				}

				// Move CAN capture enPassant, but not the attacking pawn, not allowed
				if m.EnPassantCaptureSquare != -1 && !boardhelper.IsIndexBitSet(m.EnPassantCaptureSquare, protectMoveMask) {
					continue
				}
			}
		}

		// If we do an enPassant Capture, make sure it does not leave our own king in check
		if m.EnPassantCaptureSquare != -1 && IsEnPassantMovePinned(b, m) {
			continue
		}

		legalMoves = append(legalMoves, m)
	}

	return legalMoves
}

/*
	Valid move squares
*/

func CheckMask(b *board.Board) (int, uint64) {
	checkCnt := 0
	var validMoveMask uint64

	indexOffsetsPawns := []int{b.PawnOffset - 1, b.PawnOffset + 1}

	friendlyKingIndex := b.FriendlyKingIndex

	opponentOrthogonalAttackers := b.Bitboards[b.OpponentColor|piece.Rook] | b.Bitboards[b.OpponentColor|piece.Queen]
	opponentDiagonalAttackers := b.Bitboards[b.OpponentColor|piece.Bishop] | b.Bitboards[b.OpponentColor|piece.Queen]
	enemyPawns := b.Bitboards[b.OpponentColor|piece.Pawn]

	orthogonalOtherPiecesMask := Occupancy[2] & ^opponentOrthogonalAttackers
	diagonalOtherPiecesMask := Occupancy[2] & ^opponentDiagonalAttackers

	knightAttackMask := ComputedKnightMoves[friendlyKingIndex] & b.Bitboards[b.OpponentColor|piece.Knight]

	for i, offset := range DirectionalOffsets[:4] {
		var mask uint64
		step := friendlyKingIndex + offset

		depth := 1
		for depth <= DistanceToEdge[friendlyKingIndex][i] {
			mask |= 1 << step

			if boardhelper.IsIndexBitSet(step, orthogonalOtherPiecesMask) {
				break
			}

			if boardhelper.IsIndexBitSet(step, opponentOrthogonalAttackers) {
				validMoveMask |= mask
				checkCnt++
				break
			}

			step += offset
			depth++
		}
	}

	for i, offset := range DirectionalOffsets[4:] {
		var mask uint64
		step := friendlyKingIndex + offset

		depth := 1
		for depth <= DistanceToEdge[friendlyKingIndex][i+4] {
			mask |= 1 << step

			if boardhelper.IsIndexBitSet(step, diagonalOtherPiecesMask) {
				break
			}

			if boardhelper.IsIndexBitSet(step, opponentDiagonalAttackers) {
				validMoveMask |= mask
				checkCnt++
				break
			}

			step += offset
			depth++
		}
	}

	if knightAttackMask != 0 {
		validMoveMask |= knightAttackMask
		checkCnt += bits.OnesCount64(knightAttackMask)
	}

	for _, offset := range indexOffsetsPawns {
		step := friendlyKingIndex + offset

		if !boardhelper.IsValidDiagonalMove(friendlyKingIndex, step) {
			continue
		}

		if !boardhelper.IsIndexBitSet(step, enemyPawns) {
			continue
		}

		validMoveMask |= 1 << step
		checkCnt++
	}

	return checkCnt, validMoveMask
}

func calculatePinRay(b *board.Board, pieceIndex int) uint64 {
	var pinRay uint64

	friendlyKingIndex := b.FriendlyKingIndex
	opponentPieceMask := Occupancy[b.OpponentIndex]

	offset := boardhelper.CalculateRayOffset(friendlyKingIndex, pieceIndex)
	step := friendlyKingIndex + offset

	for boardhelper.IsValidStraightMove(friendlyKingIndex, step) || boardhelper.IsValidDiagonalMove(friendlyKingIndex, step) {
		// Cannot move to our own square
		if step != pieceIndex {
			pinRay |= 1 << step
		}

		// Since we know we are pinned, the first enemy piece has to be our attacker
		if boardhelper.IsIndexBitSet(step, opponentPieceMask) {
			break
		}

		step += offset
	}

	return pinRay
}

func IsEnPassantMovePinned(b *board.Board, m move.Move) bool {

	friendlyKingIndex := b.FriendlyKingIndex

	// Can instantly return if there is no direct ray between friendlyKingIndex & enPassantCaptureSquare
	offset := boardhelper.CalculateRayOffset(friendlyKingIndex, m.EnPassantCaptureSquare)
	if offset == 0 {
		return false
	}

	enemyAttackers := b.Bitboards[b.OpponentColor|piece.Queen]
	isValidMoveFunction := boardhelper.IsValidStraightMove
	switch offset {
	case -1, 1, -8, 8:
		enemyAttackers |= b.Bitboards[b.OpponentColor|piece.Rook]
	case -7, 7, -9, 9:
		enemyAttackers |= b.Bitboards[b.OpponentColor|piece.Bishop]
		isValidMoveFunction = boardhelper.IsValidDiagonalMove
	default:
		return false
	}

	otherPieces := Occupancy[2] & ^(enemyAttackers)

	step := friendlyKingIndex + offset

	for isValidMoveFunction(friendlyKingIndex, step) {

		// Return if we hit the moves target index (new blocking piece)
		if step == m.TargetIndex {
			return false
		}

		// Return if we hit an enemy attacker
		if boardhelper.IsIndexBitSet(step, enemyAttackers) {
			return true
		}

		// Return if we hit any piece that is not on either enPassantCaptureSquare or m.StartIndex
		if step != m.EnPassantCaptureSquare &&
			step != m.StartIndex &&
			boardhelper.IsIndexBitSet(step, otherPieces) {
			return false
		}

		step += offset
	}

	// Ray was cast until the edge, we can return false
	return false
}

/*
	Utility
*/

func Perft(b *board.Board, ply int) int64 {
	/*
		Perft Testing Utility
	*/

	if ply == 0 {
		return 1
	}

	legalMoves := LegalMoves(b)
	var totalNodes int64 = 0

	// Not the official implementation, but works a lot faster
	if ply == 1 {
		return int64(len(legalMoves))
	}

	for _, m := range legalMoves {
		b.MakeMove(m)

		subNodes := Perft(b, ply-1)
		totalNodes += subNodes

		b.UnmakeMove()
	}

	return totalNodes
}
