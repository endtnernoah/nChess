package movegenerator

import (
	"endtner.dev/nChess/board"
	"endtner.dev/nChess/board/boardhelper"
	"endtner.dev/nChess/board/move"
	"endtner.dev/nChess/board/piece"
	"math/bits"
)

var DirectionalOffsets = []int{8, -8, -1, 1, 7, -7, 9, -9}
var KnightOffsets = []int{-6, 6, -10, 10, -15, 15, -17, 17}

var DistanceToEdge = func() [][]int {
	distances := make([][]int, 64)

	for file := 0; file < 8; file++ {
		for rank := 0; rank < 8; rank++ {
			numNorth := 7 - rank
			numSouth := rank
			numWest := file
			numEast := 7 - file

			squareIndex := rank*8 + file

			distances[squareIndex] = []int{
				numNorth,
				numSouth,
				numWest,
				numEast,
				Min(numNorth, numWest),
				Min(numSouth, numEast),
				Min(numNorth, numEast),
				Min(numSouth, numWest),
			}
		}
	}

	return distances
}()

var ComputedKnightMoves = func() []uint64 {
	moves := make([]uint64, 64)

	for startIndex := range 64 {
		for _, offset := range KnightOffsets {
			targetIndex := startIndex + offset
			if boardhelper.IsValidKnightMove(startIndex, targetIndex) {
				moves[startIndex] |= 1 << targetIndex
			}
		}
	}

	return moves
}()

var ComputedKingMoves = func() []uint64 {
	moves := make([]uint64, 64)

	for startIndex := range 64 {
		for i, offset := range DirectionalOffsets {
			if DistanceToEdge[startIndex][i] != 0 {
				moves[startIndex] |= 1 << (startIndex + offset)
			}
		}
	}

	return moves
}()

var ComputedPawnMoves = func() [][]uint64 {
	moves := make([][]uint64, 2)
	moves[0] = make([]uint64, 64)
	moves[1] = make([]uint64, 64)

	for startIndex := range 64 {
		targetIndexWhite := startIndex + 8
		targetIndexBlack := startIndex - 8

		if boardhelper.IsValidStraightMove(startIndex, targetIndexWhite) {
			moves[0][startIndex] |= 1 << targetIndexWhite
		}
		if boardhelper.IsValidStraightMove(startIndex, targetIndexBlack) {
			moves[1][startIndex] |= 1 << targetIndexBlack
		}
	}

	return moves
}()

var ComputedPawnAttacks = func() [][]uint64 {
	moves := make([][]uint64, 2)
	moves[0] = make([]uint64, 64)
	moves[1] = make([]uint64, 64)

	for startIndex := range 64 {
		targetIndexWhite := startIndex + 8
		targetIndexBlack := startIndex - 8

		if boardhelper.IsValidDiagonalMove(startIndex, targetIndexWhite+1) {
			moves[0][startIndex] |= 1 << (targetIndexWhite + 1)
		}
		if boardhelper.IsValidDiagonalMove(startIndex, targetIndexWhite-1) {
			moves[0][startIndex] |= 1 << (targetIndexWhite - 1)
		}

		if boardhelper.IsValidDiagonalMove(startIndex, targetIndexBlack+1) {
			moves[1][startIndex] |= 1 << (targetIndexBlack + 1)
		}
		if boardhelper.IsValidDiagonalMove(startIndex, targetIndexBlack-1) {
			moves[1][startIndex] |= 1 << (targetIndexBlack - 1)
		}
	}

	return moves
}()

/*
	Generators
*/

func PawnMoves(b *board.Board, moves *[]move.Move, index *int) {
	opponentPiecesMask := Occupancy[b.OpponentIndex]
	allPiecesMask := Occupancy[2]

	pieces := b.Bitboards[b.FriendlyColor|piece.Pawn]
	for pieces != 0 {
		// Get index of LSB
		pieceIndex := bits.TrailingZeros64(pieces)
		targetIndex := pieceIndex + b.PawnOffset

		// Continue if target index is out of bounds, just go to the next iteration
		if targetIndex < 0 || targetIndex > 63 {
			pieces &= pieces - 1
			continue
		}

		promotionFlag := targetIndex/8 == b.PromotionRank

		// Generate all valid attacks first, better for alpha-beta-pruning
		attackMask := ComputedPawnAttacks[b.FriendlyIndex][pieceIndex]
		validAttacks := attackMask & opponentPiecesMask

		// Generate en passant attacks
		if b.EnPassantTargetSquare != -1 {
			validAttacks |= attackMask & (1 << b.EnPassantTargetSquare)
		}

		// Handle pins
		if (Pins[b.FriendlyIndex] & (1 << pieceIndex)) != 0 {
			validAttacks &= calculatePinRay(b, pieceIndex)
		}

		for validAttacks != 0 {
			attackTargetIndex := bits.TrailingZeros64(validAttacks)

			if promotionFlag {
				(*moves)[*index] = move.New(pieceIndex, attackTargetIndex, move.WithPromotion(b.FriendlyColor|piece.Knight))
				*index++
				(*moves)[*index] = move.New(pieceIndex, attackTargetIndex, move.WithPromotion(b.FriendlyColor|piece.Bishop))
				*index++
				(*moves)[*index] = move.New(pieceIndex, attackTargetIndex, move.WithPromotion(b.FriendlyColor|piece.Queen))
				*index++
				(*moves)[*index] = move.New(pieceIndex, attackTargetIndex, move.WithPromotion(b.FriendlyColor|piece.Rook))
				*index++
			} else {
				if attackTargetIndex == b.EnPassantTargetSquare {
					(*moves)[*index] = move.New(pieceIndex, attackTargetIndex, move.WithEnPassantCaptureSquare(attackTargetIndex-b.PawnOffset))
					*index++
				} else {
					(*moves)[*index] = move.New(pieceIndex, attackTargetIndex)
					*index++
				}
			}

			validAttacks &= validAttacks - 1
		}

		// Add move if target square is empty
		if !boardhelper.IsIndexBitSet(targetIndex, allPiecesMask) {
			// Add all promotion moves
			if promotionFlag {
				(*moves)[*index] = move.New(pieceIndex, targetIndex, move.WithPromotion(b.FriendlyColor|piece.Knight))
				*index++
				(*moves)[*index] = move.New(pieceIndex, targetIndex, move.WithPromotion(b.FriendlyColor|piece.Bishop))
				*index++
				(*moves)[*index] = move.New(pieceIndex, targetIndex, move.WithPromotion(b.FriendlyColor|piece.Queen))
				*index++
				(*moves)[*index] = move.New(pieceIndex, targetIndex, move.WithPromotion(b.FriendlyColor|piece.Rook))
				*index++
			} else {
				(*moves)[*index] = move.New(pieceIndex, targetIndex)
				*index++
			}
		}

		// Check if the pawn is on starting square
		isStartingSquare := pieceIndex >= 8 && pieceIndex < 16
		if !b.WhiteToMove {
			isStartingSquare = pieceIndex >= 48 && pieceIndex < 56
		}

		// Can move 2 rows from starting square
		if isStartingSquare {
			epTargetIndex := targetIndex + b.PawnOffset
			if !boardhelper.IsIndexBitSet(targetIndex, allPiecesMask) &&
				!boardhelper.IsIndexBitSet(epTargetIndex, allPiecesMask) {
				(*moves)[*index] = move.New(pieceIndex, epTargetIndex, move.WithEnPassantPassedSquare(targetIndex))
				*index++
			}
		}

		// Remove LSB of bitboard
		pieces &= pieces - 1
	}
}

func SlidingMoves(b *board.Board, moves *[]move.Move, index *int) {
	friendlyPiecesMask := Occupancy[b.FriendlyIndex]
	allPiecesMask := Occupancy[2]

	friendlyOrthogonalSliders := b.Bitboards[b.FriendlyColor|piece.Rook] | b.Bitboards[b.FriendlyColor|piece.Queen]
	friendlyDiagonalSliders := b.Bitboards[b.FriendlyColor|piece.Bishop] | b.Bitboards[b.FriendlyColor|piece.Queen]

	for friendlyOrthogonalSliders|friendlyDiagonalSliders != 0 {
		pieceIndex := bits.TrailingZeros64(friendlyOrthogonalSliders | friendlyDiagonalSliders)

		if friendlyOrthogonalSliders&(1<<pieceIndex) != 0 {
			entry := RookMagics[pieceIndex]
			moveIndex := ((allPiecesMask & entry.Mask) * entry.Magic) >> (64 - entry.Shift)

			validMoveMask := RookMoveTable[entry.Offset+int(moveIndex)] & ^friendlyPiecesMask
			if (Pins[b.FriendlyIndex] & (1 << pieceIndex)) != 0 {
				validMoveMask &= calculatePinRay(b, pieceIndex)
			}

			for validMoveMask != 0 {
				targetIndex := bits.TrailingZeros64(validMoveMask)

				(*moves)[*index] = move.New(pieceIndex, targetIndex)
				*index++

				validMoveMask &= validMoveMask - 1
			}

			friendlyOrthogonalSliders &= friendlyOrthogonalSliders - 1
		}
		if friendlyDiagonalSliders&(1<<pieceIndex) != 0 {
			entry := BishopMagics[pieceIndex]
			moveIndex := ((allPiecesMask & entry.Mask) * entry.Magic) >> (64 - entry.Shift)

			validMoveMask := BishopMoveTable[entry.Offset+int(moveIndex)] & ^friendlyPiecesMask
			if (Pins[b.FriendlyIndex] & 1 << pieceIndex) != 0 {
				validMoveMask &= calculatePinRay(b, pieceIndex)
			}

			for validMoveMask != 0 {
				targetIndex := bits.TrailingZeros64(validMoveMask)

				(*moves)[*index] = move.New(pieceIndex, targetIndex)
				*index++

				validMoveMask &= validMoveMask - 1
			}

			friendlyDiagonalSliders &= friendlyDiagonalSliders - 1
		}
	}
}

func KnightMoves(b *board.Board, moves *[]move.Move, index *int) {
	friendlyPiecesMask := Occupancy[b.FriendlyIndex]

	knights := b.Bitboards[b.FriendlyColor|piece.Knight] & ^Pins[b.FriendlyIndex] // Knights cannot move if pinned
	for knights != 0 {
		pieceIndex := bits.TrailingZeros64(knights)

		validMoveMask := ComputedKnightMoves[pieceIndex] & ^friendlyPiecesMask

		for validMoveMask != 0 {
			(*moves)[*index] = move.New(pieceIndex, bits.TrailingZeros64(validMoveMask))
			*index++

			validMoveMask &= validMoveMask - 1
		}

		knights &= knights - 1
	}
}

func KingMoves(b *board.Board, moves *[]move.Move, index *int) {
	friendlyKingIndex := b.FriendlyKingIndex
	friendlyRooks := b.Bitboards[b.FriendlyColor|piece.Rook]

	allPiecesMask := Occupancy[2]
	friendlyPiecesMask := Occupancy[b.FriendlyIndex]
	opponentAttackMask := Attacks[b.OpponentIndex]

	kingMoveMask := ComputedKingMoves[friendlyKingIndex] & ^friendlyPiecesMask & ^opponentAttackMask
	for kingMoveMask != 0 {
		(*moves)[*index] = move.New(friendlyKingIndex, bits.TrailingZeros64(kingMoveMask))
		*index++

		kingMoveMask &= kingMoveMask - 1
	}

	initialKingIndex := 4
	if !b.WhiteToMove {
		initialKingIndex = 60
	}

	// King is not on its original square, will not be allowed to castle
	if friendlyKingIndex != initialKingIndex {
		return
	}

	kingSideAllowed := b.CastlingAvailability&0b1000 != 0
	queenSideAllowed := b.CastlingAvailability&0b0100 != 0

	var kingSideEmptyMask uint64 = 0b1100000
	var queenSideEmptyMask uint64 = 0b1110

	kingSideRookIndex := 7
	queenSideRookIndex := 0

	var kingSideAttackMask uint64 = 1<<friendlyKingIndex | 1<<(friendlyKingIndex+1) | 1<<(friendlyKingIndex+2)
	var queenSideAttackMask uint64 = 1<<friendlyKingIndex | 1<<(friendlyKingIndex-1) | 1<<(friendlyKingIndex-2)

	if !b.WhiteToMove {
		kingSideAllowed = b.CastlingAvailability&0b0010 != 0
		queenSideAllowed = b.CastlingAvailability&0b0001 != 0

		kingSideEmptyMask <<= 56
		queenSideEmptyMask <<= 56

		kingSideRookIndex += 56
		queenSideRookIndex += 56
	}

	if kingSideAllowed &&
		(kingSideAttackMask&opponentAttackMask) == 0 && // King does not start or pass through attacked field
		(kingSideEmptyMask&allPiecesMask) == 0 && // All fields are empty
		boardhelper.IsIndexBitSet(kingSideRookIndex, friendlyRooks) { // There is a rook on its field

		(*moves)[*index] = move.New(friendlyKingIndex, friendlyKingIndex+2, move.WithRookStartingSquare(kingSideRookIndex))
		*index++
	}

	if queenSideAllowed &&
		(queenSideAttackMask&opponentAttackMask) == 0 &&
		(queenSideEmptyMask&allPiecesMask) == 0 &&
		boardhelper.IsIndexBitSet(queenSideRookIndex, friendlyRooks) {

		(*moves)[*index] = move.New(friendlyKingIndex, friendlyKingIndex-2, move.WithRookStartingSquare(queenSideRookIndex))
		*index++
	}
}

/*
	Utility
*/

func Min(x int, y int) int {
	if x < y {
		return x
	}
	return y
}
