package movegenerator

import (
	"endtner.dev/nChess/board"
	"endtner.dev/nChess/board/boardhelper"
	"endtner.dev/nChess/board/move"
	"endtner.dev/nChess/board/piece"
	"math/bits"
	"sync"
)

/*
	MoveGenerator is now responsible for generating all Bitboards. THIS CODE DOES NOT RUN IN PARALLEL!

	When parallelized, other generators will override these variables for a given generator, so everything crashes.
	I tried instancing the generator per class, that created so much overhead that the performance increase was mitigated.
*/

var DirectionalOffsets = []int{8, -8, -1, 1, 7, -7, 9, -9}
var KnightOffsets = []int{-6, 6, -10, 10, -15, 15, -17, 17}

var DistanceToEdge = computeDistanceToEdges()

func computeDistanceToEdges() [][]int {
	precomputedDistances := make([][]int, 64)

	for file := 0; file < 8; file++ {
		for rank := 0; rank < 8; rank++ {
			numNorth := 7 - rank
			numSouth := rank
			numWest := file
			numEast := 7 - file

			squareIndex := rank*8 + file

			precomputedDistances[squareIndex] = []int{
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

	return precomputedDistances
}

var ComputedKnightMoves = computeKnightMoves()

func computeKnightMoves() []uint64 {
	var knightMoves = make([]uint64, 64)

	for startIndex := range 64 {
		var validMoves uint64

		for _, offset := range KnightOffsets {
			targetIndex := startIndex + offset

			if boardhelper.IsValidKnightMove(startIndex, targetIndex) {
				validMoves |= 1 << targetIndex
			}
		}

		knightMoves[startIndex] = validMoves
	}

	return knightMoves
}

// ComputedOccupancy & ComputedAttacks & ComputedPins can be accessed at index (3 >> color) - 1 & 1 - accessIndex for enemy color
var ComputedOccupancy = make([]uint64, 3)
var ComputedAttacks = make([]uint64, 2)
var ComputedPins = make([]uint64, 2)

func Min(x int, y int) int {
	if x < y {
		return x
	}
	return y
}

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

	ComputedOccupancy[0] = whitePieces
	ComputedOccupancy[1] = blackPieces
	ComputedOccupancy[2] = whitePieces | blackPieces

	ComputeAttacks(b, piece.White)
	ComputeAttacks(b, piece.Black)
	ComputePins(b, piece.White)
	ComputePins(b, piece.Black)
}

/*
	Generators
*/

func PseudoLegalMoves(b *board.Board) []move.Move {
	/*
		Creating pseudo-legal moves in an iterative approach. This function is used in the actual implementation.
	*/

	pseudoLegalMoves := make([]move.Move, 0, 218) // Maximum possible moves in a chess position is 218

	colorToMove := piece.White

	if !b.WhiteToMove {
		colorToMove = piece.Black
	}

	pseudoLegalMoves = append(pseudoLegalMoves, PawnMoves(b, colorToMove, b.EnPassantTargetSquare)...)
	pseudoLegalMoves = append(pseudoLegalMoves, SlidingMoves(b, colorToMove)...)
	pseudoLegalMoves = append(pseudoLegalMoves, KnightMoves(b, colorToMove)...)
	pseudoLegalMoves = append(pseudoLegalMoves, KingMoves(b, colorToMove, b.CastlingAvailability)...)

	return pseudoLegalMoves
}

func PseudoLegalMovesParallel(b *board.Board) []move.Move {
	/*
		Generating all pseudo-legal moves in parallel. Somehow, this is slower than generating them in a normal way.
		I have not figured out why, so I am leaving this function in here. One possible reason might be the overhead the parallelization creates.
	*/
	pseudoLegalMovesChan := make(chan []move.Move)

	var waitGroup sync.WaitGroup
	waitGroup.Add(5)

	colorToMove := piece.White
	if !b.WhiteToMove {
		colorToMove = piece.Black
	}

	// Pawn moves
	go func(b *board.Board, colorToMove uint8, enPassantTargetSquare int) {
		defer waitGroup.Done()
		pseudoLegalMovesChan <- PawnMoves(b, colorToMove, enPassantTargetSquare)
	}(b, colorToMove, b.EnPassantTargetSquare)

	// Straight sliding moves
	go func(b *board.Board, colorToMove uint8) {
		defer waitGroup.Done()
		pseudoLegalMovesChan <- StraightSlidingMoves(b, colorToMove)
	}(b, colorToMove)

	// Diagonal sliding moves
	go func(b *board.Board, colorToMove uint8) {
		defer waitGroup.Done()
		pseudoLegalMovesChan <- DiagonalSlidingMoves(b, colorToMove)
	}(b, colorToMove)

	// Knight moves
	go func(b *board.Board, colorToMove uint8) {
		defer waitGroup.Done()
		pseudoLegalMovesChan <- KnightMoves(b, colorToMove)
	}(b, colorToMove)

	// King moves
	go func(b *board.Board, colorToMove uint8, castlingAvailability uint8) {
		defer waitGroup.Done()
		pseudoLegalMovesChan <- KingMoves(b, colorToMove, castlingAvailability)
	}(b, colorToMove, b.CastlingAvailability)

	// Wait for all generators to finish
	go func() {
		waitGroup.Wait()
		close(pseudoLegalMovesChan)
	}()

	// Joining to a list, returning
	var pseudoLegalMoves []move.Move
	for m := range pseudoLegalMovesChan {
		pseudoLegalMoves = append(pseudoLegalMoves, m...)
	}

	return pseudoLegalMoves
}

func LegalMoves(b *board.Board) []move.Move {
	/*
		Filtering out all illegal moves
	*/

	// Precompute attacks, pins
	ComputeAll(b)

	legalMoves := make([]move.Move, 0, 218) // Maximum possible moves in a chess position is 218

	pseudoLegalMoves := PseudoLegalMoves(b)

	colorToMove := piece.White

	if !b.WhiteToMove {
		colorToMove = piece.Black
	}

	ownKingBitboard := b.Bitboards[colorToMove|piece.King]
	ownPinnedPieces := ComputedPins[(colorToMove>>3)-1]

	ownKingIndex := bits.TrailingZeros64(ownKingBitboard)

	enemyAttackFields := ComputedAttacks[1-((colorToMove>>3)-1)]

	checkCount := 0
	possibleProtectMoves := ^uint64(0)

	if boardhelper.IsIndexBitSet(ownKingIndex, enemyAttackFields) {
		checkCount, possibleProtectMoves = CheckMask(b, colorToMove)
	}

	//fmt.Println(formatter.FormatUnicodeBoardWithBorders(formatter.ToUnicodeBoard(map[uint64]string{enemyAttackFields: "A"})))

	for _, m := range pseudoLegalMoves {

		// Only move along pin ray if piece is pinned
		if boardhelper.IsIndexBitSet(m.StartIndex, ownPinnedPieces) && !IsPinnedMoveAlongRay(b, colorToMove, m) {
			continue
		}

		// If the king is in check
		if checkCount > 0 {

			// King is in single check
			if checkCount == 1 {

				// If a move is not to any of the protect square OR not a king move
				if !boardhelper.IsIndexBitSet(m.TargetIndex, possibleProtectMoves) && (ownKingIndex != m.StartIndex) {

					// Move can not enPassantCapture the checking pawn, not allowed
					if m.EnPassantCaptureSquare == -1 {
						continue
					}

					// Move CAN capture enPassant, but not the attacking pawn, not allowed
					if m.EnPassantCaptureSquare != -1 && !boardhelper.IsIndexBitSet(m.EnPassantCaptureSquare, possibleProtectMoves) {
						continue
					}
				}
			}

			// If the king is in double (or higher) check, only allow king moves
			if checkCount > 1 && ownKingIndex != m.StartIndex {
				continue
			}
		}

		// If we do an enPassant Capture, make sure it does not leave our own king in check
		if m.EnPassantCaptureSquare != -1 && IsEnPassantMovePinned(b, colorToMove, m) {
			continue
		}

		legalMoves = append(legalMoves, m)
	}

	return legalMoves
}

func PawnMoves(b *board.Board, pieceColor uint8, enPassantTargetSquare int) []move.Move {
	var pawnMoves []move.Move

	// Default offset for white
	pawnIndexOffset := 8
	promotionRank := 7

	// Change offset for black
	if pieceColor != piece.White {
		pawnIndexOffset = -8
		promotionRank = 0
	}

	enemyOccupiedFields := ComputedOccupancy[1-((pieceColor>>3)-1)]
	occupiedFields := ComputedOccupancy[2]

	pawnBitboard := b.Bitboards[pieceColor|piece.Pawn]
	for pawnBitboard != 0 {
		// Get index of LSB
		startIndex := bits.TrailingZeros64(pawnBitboard)
		targetIndex := startIndex + pawnIndexOffset

		// Continue if target index is out of bounds, just go to the next iteration
		if targetIndex < 0 || targetIndex > 63 {
			pawnBitboard &= pawnBitboard - 1
			continue
		}

		promotionFlag := targetIndex/8 == promotionRank

		// Add move if target square is empty
		if !boardhelper.IsIndexBitSet(targetIndex, occupiedFields) {
			// Add all promotion moves
			if promotionFlag {
				pawnMoves = append(pawnMoves, move.New(startIndex, targetIndex, move.WithPromotion(pieceColor|piece.Knight)))
				pawnMoves = append(pawnMoves, move.New(startIndex, targetIndex, move.WithPromotion(pieceColor|piece.Bishop)))
				pawnMoves = append(pawnMoves, move.New(startIndex, targetIndex, move.WithPromotion(pieceColor|piece.Queen)))
				pawnMoves = append(pawnMoves, move.New(startIndex, targetIndex, move.WithPromotion(pieceColor|piece.Rook)))
			} else {
				pawnMoves = append(pawnMoves, move.New(startIndex, targetIndex))
			}
		}

		// Check if the pawn is on starting square
		isStartingSquare := startIndex >= 8 && startIndex < 16
		if pieceColor != piece.White {
			isStartingSquare = startIndex >= 48 && startIndex < 56
		}

		// Can move 2 rows from starting square
		if isStartingSquare {
			if !boardhelper.IsIndexBitSet(targetIndex, occupiedFields) &&
				!boardhelper.IsIndexBitSet(targetIndex+pawnIndexOffset, occupiedFields) {
				pawnMoves = append(pawnMoves, move.New(startIndex, targetIndex+pawnIndexOffset, move.WithEnPassantPassedSquare(targetIndex)))
			}
		}

		// Check for possible captures
		side1 := targetIndex - 1
		side2 := targetIndex + 1

		// Add move if it is in the same target row and targets are not empty
		if boardhelper.IsValidDiagonalMove(startIndex, side1) && boardhelper.IsIndexBitSet(side1, enemyOccupiedFields) {
			if promotionFlag {
				pawnMoves = append(pawnMoves, move.New(startIndex, side1, move.WithPromotion(pieceColor|piece.Knight)))
				pawnMoves = append(pawnMoves, move.New(startIndex, side1, move.WithPromotion(pieceColor|piece.Bishop)))
				pawnMoves = append(pawnMoves, move.New(startIndex, side1, move.WithPromotion(pieceColor|piece.Queen)))
				pawnMoves = append(pawnMoves, move.New(startIndex, side1, move.WithPromotion(pieceColor|piece.Rook)))
			} else {
				pawnMoves = append(pawnMoves, move.New(startIndex, side1))
			}
		}
		if boardhelper.IsValidDiagonalMove(startIndex, side2) && boardhelper.IsIndexBitSet(side2, enemyOccupiedFields) {
			if promotionFlag {
				pawnMoves = append(pawnMoves, move.New(startIndex, side2, move.WithPromotion(pieceColor|piece.Knight)))
				pawnMoves = append(pawnMoves, move.New(startIndex, side2, move.WithPromotion(pieceColor|piece.Bishop)))
				pawnMoves = append(pawnMoves, move.New(startIndex, side2, move.WithPromotion(pieceColor|piece.Queen)))
				pawnMoves = append(pawnMoves, move.New(startIndex, side2, move.WithPromotion(pieceColor|piece.Rook)))
			} else {
				pawnMoves = append(pawnMoves, move.New(startIndex, side2))
			}
		}

		// Add move if either side can capture en passant
		if boardhelper.IsValidDiagonalMove(startIndex, side1) && side1 == enPassantTargetSquare {
			pawnMoves = append(pawnMoves, move.New(startIndex, side1, move.WithEnPassantCaptureSquare(side1-pawnIndexOffset)))
		}
		if boardhelper.IsValidDiagonalMove(startIndex, side2) && side2 == enPassantTargetSquare {
			pawnMoves = append(pawnMoves, move.New(startIndex, side2, move.WithEnPassantCaptureSquare(side2-pawnIndexOffset)))
		}

		// Remove LSB of bitboard
		pawnBitboard &= pawnBitboard - 1
	}

	return pawnMoves
}

func SlidingMoves(b *board.Board, pieceColor uint8) []move.Move {
	var slidingMoves []move.Move

	ownOccupiedFields := ComputedOccupancy[(pieceColor>>3)-1]
	enemyOccupiedFields := ComputedOccupancy[1-((pieceColor>>3)-1)]

	piecesBitboard := b.Bitboards[pieceColor|piece.Rook] | b.Bitboards[pieceColor|piece.Bishop] | b.Bitboards[pieceColor|piece.Queen]
	for piecesBitboard != 0 {
		startIndex := bits.TrailingZeros64(piecesBitboard)
		pieceType := b.Pieces[startIndex] & 0b00111

		offsetIndexStart := 0
		offsetIndexEnd := 8

		// Manipulating indices based on piece type
		if pieceType == piece.Rook {
			offsetIndexEnd = 4
		}
		if pieceType == piece.Bishop {
			offsetIndexStart = 4
		}

		for i, offset := range DirectionalOffsets[offsetIndexStart:offsetIndexEnd] {
			targetIndex := startIndex + offset

			depth := 1
			for depth <= DistanceToEdge[startIndex][i+offsetIndexStart] {
				if boardhelper.IsIndexBitSet(targetIndex, ownOccupiedFields) {
					break
				}

				slidingMoves = append(slidingMoves, move.New(startIndex, targetIndex))

				// Break if we captured an enemy
				if boardhelper.IsIndexBitSet(targetIndex, enemyOccupiedFields) {
					break
				}

				targetIndex += offset
				depth++
			}
		}

		piecesBitboard &= piecesBitboard - 1
	}

	return slidingMoves
}

func StraightSlidingMoves(b *board.Board, pieceColor uint8) []move.Move {
	var straightSlidingMoves []move.Move

	ownOccupiedFields := ComputedOccupancy[(pieceColor>>3)-1]
	enemyOccupiedFields := ComputedOccupancy[1-((pieceColor>>3)-1)]

	piecesBitboard := b.Bitboards[pieceColor|piece.Rook] | b.Bitboards[pieceColor|piece.Queen]
	for piecesBitboard != 0 {
		startIndex := bits.TrailingZeros64(piecesBitboard)

		for i, offset := range DirectionalOffsets[:4] {
			targetIndex := startIndex + offset

			depth := 1
			for depth <= DistanceToEdge[startIndex][i] {

				if boardhelper.IsIndexBitSet(targetIndex, ownOccupiedFields) {
					break
				}

				straightSlidingMoves = append(straightSlidingMoves, move.New(startIndex, targetIndex))

				// Break if we captured an enemy
				if boardhelper.IsIndexBitSet(targetIndex, enemyOccupiedFields) {
					break
				}

				targetIndex += offset
				depth++
			}
		}

		piecesBitboard &= piecesBitboard - 1
	}

	return straightSlidingMoves
}

func DiagonalSlidingMoves(b *board.Board, pieceColor uint8) []move.Move {
	var diagonalSlidingMoves []move.Move

	ownOccupiedFields := ComputedOccupancy[(pieceColor>>3)-1]
	enemyOccupiedFields := ComputedOccupancy[1-((pieceColor>>3)-1)]

	piecesBitboard := b.Bitboards[pieceColor|piece.Bishop] | b.Bitboards[pieceColor|piece.Queen]
	for piecesBitboard != 0 {
		startIndex := bits.TrailingZeros64(piecesBitboard)

		// Go as deep as possible
		for i, offset := range DirectionalOffsets[4:] {
			targetIndex := startIndex + offset

			depth := 1
			for depth <= DistanceToEdge[startIndex][i+4] {

				if boardhelper.IsIndexBitSet(targetIndex, ownOccupiedFields) {
					break
				}

				diagonalSlidingMoves = append(diagonalSlidingMoves, move.New(startIndex, targetIndex))

				// Break if we captured an enemy
				if boardhelper.IsIndexBitSet(targetIndex, enemyOccupiedFields) {
					break
				}

				targetIndex += offset
				depth++
			}
		}

		// Remove LSB
		piecesBitboard &= piecesBitboard - 1
	}

	return diagonalSlidingMoves
}

func KnightMoves(b *board.Board, pieceColor uint8) []move.Move {
	var validMoves []move.Move

	ownOccupiedFields := ComputedOccupancy[(pieceColor>>3)-1]

	knights := b.Bitboards[pieceColor|piece.Knight]
	for knights != 0 {
		startIndex := bits.TrailingZeros64(knights)
		validMoveMask := ComputedKnightMoves[startIndex] & ^ownOccupiedFields

		for validMoveMask != 0 {
			validMoves = append(validMoves, move.New(startIndex, bits.TrailingZeros64(validMoveMask)))
			validMoveMask &= validMoveMask - 1
		}

		knights &= knights - 1
	}

	return validMoves
}

func KingMoves(b *board.Board, pieceColor uint8, castlingAvailability uint8) []move.Move {
	var kingMoves []move.Move

	kingBitboard := b.Bitboards[pieceColor|piece.King]
	rookBitboard := b.Bitboards[pieceColor|piece.Rook]
	startIndex := bits.TrailingZeros64(kingBitboard)

	allOccupiedFields := ComputedOccupancy[2]
	ownOccupiedFields := ComputedOccupancy[(pieceColor>>3)-1]
	enemyAttackFields := ComputedAttacks[1-((pieceColor>>3)-1)]

	for i, offset := range DirectionalOffsets {
		targetIndex := startIndex + offset

		if targetIndex < 0 || targetIndex > 63 {
			continue
		}

		if DistanceToEdge[startIndex][i] == 0 ||
			boardhelper.IsIndexBitSet(targetIndex, ownOccupiedFields) ||
			boardhelper.IsIndexBitSet(targetIndex, enemyAttackFields) {
			continue
		}

		kingMoves = append(kingMoves, move.New(startIndex, targetIndex))
	}

	initialKingSquare := 4
	if pieceColor != piece.White {
		initialKingSquare = 60
	}

	// King is not on its original square, will not be allowed to castle
	if startIndex != initialKingSquare {
		return kingMoves
	}

	kingSideAllowed := castlingAvailability&0b1000 != 0
	queenSideAllowed := castlingAvailability&0b0100 != 0

	var kingSideEmptyMask uint64 = 0b1100000
	var queenSideEmptyMask uint64 = 0b1110

	kingSideRookIndex := 7
	queenSideRookIndex := 0

	var kingSideAttackMask uint64 = 1<<startIndex | 1<<(startIndex+1) | 1<<(startIndex+2)
	var queenSideAttackMask uint64 = 1<<startIndex | 1<<(startIndex-1) | 1<<(startIndex-2)

	if pieceColor != piece.White {
		kingSideAllowed = castlingAvailability&0b0010 != 0
		queenSideAllowed = castlingAvailability&0b0001 != 0

		kingSideEmptyMask <<= 56
		queenSideEmptyMask <<= 56

		kingSideRookIndex += 56
		queenSideRookIndex += 56
	}

	if kingSideAllowed &&
		(kingSideAttackMask&enemyAttackFields) == 0 && // King does not start or pass through attacked field
		(kingSideEmptyMask&allOccupiedFields) == 0 && // All fields are empty
		boardhelper.IsIndexBitSet(kingSideRookIndex, rookBitboard) { // There is a rook on its field
		kingMoves = append(kingMoves, move.New(startIndex, startIndex+2, move.WithRookStartingSquare(kingSideRookIndex)))
	}

	if queenSideAllowed &&
		(queenSideAttackMask&enemyAttackFields) == 0 &&
		(queenSideEmptyMask&allOccupiedFields) == 0 &&
		boardhelper.IsIndexBitSet(queenSideRookIndex, rookBitboard) {
		kingMoves = append(kingMoves, move.New(startIndex, startIndex-2, move.WithRookStartingSquare(queenSideRookIndex)))
	}

	return kingMoves
}

/*
	Attack pattern generators
*/

func ComputeAttacks(b *board.Board, color uint8) {
	pawnAttacks := PawnAttacks(b, color)
	straightSlidingAttacks := StraightSlidingAttacks(b, color)
	diagonalSlidingAttacks := DiagonalSlidingAttacks(b, color)
	knightAttacks := KnightAttacks(b, color)

	ComputedAttacks[(color>>3)-1] = pawnAttacks |
		straightSlidingAttacks |
		diagonalSlidingAttacks |
		knightAttacks
}

func PawnAttacks(b *board.Board, colorToMove uint8) uint64 {
	var pawnAttacks uint64

	pawnIndexOffset := 8

	if colorToMove != piece.White {
		pawnIndexOffset = -8
	}

	pieceBitboard := b.Bitboards[colorToMove|piece.Pawn]

	for pieceBitboard != 0 {
		startIndex := bits.TrailingZeros64(pieceBitboard)

		side1 := startIndex + pawnIndexOffset - 1
		side2 := startIndex + pawnIndexOffset + 1

		// Add move if it is in the same target row and targets are not empty
		if boardhelper.IsValidDiagonalMove(startIndex, side1) {
			pawnAttacks |= 1 << side1
		}
		if boardhelper.IsValidDiagonalMove(startIndex, side2) {
			pawnAttacks |= 1 << side2
		}

		pieceBitboard &= pieceBitboard - 1
	}

	return pawnAttacks
}

func StraightSlidingAttacks(b *board.Board, colorToMove uint8) uint64 {
	straightIndexOffsets := []int{1, -1, 8, -8}
	var straightSlidingAttacks uint64

	pieceBitboard := b.Bitboards[colorToMove|piece.Rook] | b.Bitboards[colorToMove|piece.Queen] | b.Bitboards[colorToMove|piece.King]
	ownPiecesBitboard := ComputedOccupancy[(colorToMove>>3)-1]
	enemyPiecesBitboard := ComputedOccupancy[1-((colorToMove>>3)-1)]

	enemyColor := piece.Black
	if colorToMove != piece.White {
		enemyColor = piece.White
	}
	enemyKingIndex := bits.TrailingZeros64(b.Bitboards[enemyColor|piece.King])

	for pieceBitboard != 0 {
		startIndex := bits.TrailingZeros64(pieceBitboard)

		maxLength := 8
		if (b.Pieces[startIndex] & 0b00111) == piece.King {
			maxLength = 1
		}

		for _, offset := range straightIndexOffsets {
			targetIndex := startIndex + offset

			length := 0
			// Go deep until we hit maxLength or crossed borders
			for length < maxLength && boardhelper.IsValidStraightMove(startIndex, targetIndex) {

				straightSlidingAttacks |= 1 << targetIndex

				// If we hit our own piece, we break
				if boardhelper.IsIndexBitSet(targetIndex, ownPiecesBitboard) {
					break
				}

				// If we hit an enemy piece except the enemy king, we break
				if boardhelper.IsIndexBitSet(targetIndex, enemyPiecesBitboard) && (targetIndex != enemyKingIndex) {
					break
				}

				targetIndex += offset
				length++
			}

		}

		pieceBitboard &= pieceBitboard - 1
	}

	return straightSlidingAttacks
}

func DiagonalSlidingAttacks(b *board.Board, colorToMove uint8) uint64 {
	diagonalIndexOffsets := []int{-7, 7, -9, 9}
	var diagonalSlidingAttacks uint64

	pieceBitboard := b.Bitboards[colorToMove|piece.Bishop] | b.Bitboards[colorToMove|piece.Queen] | b.Bitboards[colorToMove|piece.King]
	ownPiecesBitboard := ComputedOccupancy[(colorToMove>>3)-1]
	enemyPiecesBitboard := ComputedOccupancy[1-((colorToMove>>3)-1)]

	enemyColor := piece.Black
	if colorToMove != piece.White {
		enemyColor = piece.White
	}
	enemyKingIndex := bits.TrailingZeros64(b.Bitboards[enemyColor|piece.King])

	for pieceBitboard != 0 {
		startIndex := bits.TrailingZeros64(pieceBitboard)

		maxLength := 8
		if (b.Pieces[startIndex] & 0b00111) == piece.King {
			maxLength = 1
		}

		for _, offset := range diagonalIndexOffsets {
			targetIndex := startIndex + offset

			length := 0
			// Go deep until we hit maxlength or crossed border
			for length < maxLength && boardhelper.IsValidDiagonalMove(startIndex, targetIndex) {

				diagonalSlidingAttacks |= 1 << targetIndex

				// If we hit our own piece, we break
				if boardhelper.IsIndexBitSet(targetIndex, ownPiecesBitboard) {
					break
				}

				// If we hit an enemy piece except the enemy king, we break
				if boardhelper.IsIndexBitSet(targetIndex, enemyPiecesBitboard) && (targetIndex != enemyKingIndex) {
					break
				}

				targetIndex += offset
				length++
			}

		}

		pieceBitboard &= pieceBitboard - 1
	}

	return diagonalSlidingAttacks
}

func KnightAttacks(b *board.Board, colorToMove uint8) uint64 {
	var attacks uint64

	knights := b.Bitboards[colorToMove|piece.Knight]
	for knights != 0 {
		attacks |= ComputedKnightMoves[bits.TrailingZeros64(knights)]
		knights &= knights - 1
	}

	return attacks
}

/*
	Calculating pins
*/

func ComputePins(b *board.Board, color uint8) {
	straightPinnedPieces := StraightPins(b, color)
	diagonalPinnedPieces := DiagonalPins(b, color)

	ComputedPins[(color>>3)-1] = straightPinnedPieces | diagonalPinnedPieces
}

func StraightPins(b *board.Board, colorToMove uint8) uint64 {
	var pinnedPieces uint64

	kingIndex := bits.TrailingZeros64(b.Bitboards[colorToMove|piece.King])

	ownPiecesBitboard := ComputedOccupancy[(colorToMove>>3)-1]
	enemyPiecesBitboard := ComputedOccupancy[1-((colorToMove>>3)-1)]

	enemyColor := piece.White
	if colorToMove == piece.White {
		enemyColor = piece.Black
	}

	attackingEnemyPiecesBitboard := b.Bitboards[enemyColor|piece.Rook] | b.Bitboards[enemyColor|piece.Queen]
	otherEnemyPiecesBitboard := enemyPiecesBitboard & ^attackingEnemyPiecesBitboard

	for i, offset := range DirectionalOffsets[:4] {
		rayIndex := kingIndex + offset
		ownPieceIndex := -1

		if rayIndex < 0 || rayIndex > 63 {
			continue
		}

		depth := 1
		for depth <= DistanceToEdge[kingIndex][i] {

			// Hit our own piece
			if boardhelper.IsIndexBitSet(rayIndex, ownPiecesBitboard) {

				// Hit own piece 2 times in a row
				if ownPieceIndex != -1 {
					break
				}

				// Hit own piece first time
				ownPieceIndex = rayIndex
			}

			// Hit enemy attacking piece
			if boardhelper.IsIndexBitSet(rayIndex, attackingEnemyPiecesBitboard) {

				// If we have hit our own piece before, that piece is pinned
				if ownPieceIndex != -1 {
					pinnedPieces |= 1 << ownPieceIndex
					break
				}

				// Fun fact: we are in check if the code comes to this comment
			}

			// Hit any other enemy piece
			if boardhelper.IsIndexBitSet(rayIndex, otherEnemyPiecesBitboard) {
				break // Can just exit, no matter if we hit our own piece first, there is an enemy piece in the way of possible attackers
			}

			rayIndex += offset
			depth++
		}
	}

	return pinnedPieces
}

func DiagonalPins(b *board.Board, colorToMove uint8) uint64 {
	var pinnedPieces uint64

	kingIndex := bits.TrailingZeros64(b.Bitboards[colorToMove|piece.King])
	ownPiecesBitboard := ComputedOccupancy[(colorToMove>>3)-1]
	enemyPiecesBitboard := ComputedOccupancy[1-((colorToMove>>3)-1)]

	enemyColor := piece.White
	if colorToMove == piece.White {
		enemyColor = piece.Black
	}

	attackingEnemyPiecesBitboard := b.Bitboards[enemyColor|piece.Bishop] | b.Bitboards[enemyColor|piece.Queen]
	otherEnemyPiecesBitboard := enemyPiecesBitboard & ^attackingEnemyPiecesBitboard

	for i, offset := range DirectionalOffsets[4:] {
		rayIndex := kingIndex + offset
		ownPieceIndex := -1

		if rayIndex < 0 || rayIndex > 63 {
			continue
		}

		// Go as long as the ray moves to valid fields
		depth := 1
		for depth <= DistanceToEdge[kingIndex][i+4] {
			// Hit our own piece
			if boardhelper.IsIndexBitSet(rayIndex, ownPiecesBitboard) {

				// Hit own piece 2 times in a row
				if ownPieceIndex != -1 {
					break
				}

				// Hit own piece first time
				ownPieceIndex = rayIndex
			}

			// Hit enemy attacking piece
			if boardhelper.IsIndexBitSet(rayIndex, attackingEnemyPiecesBitboard) {

				// If we have hit our own piece before, that piece is pinned
				if ownPieceIndex != -1 {
					pinnedPieces |= 1 << ownPieceIndex
					break
				}

				// Fun fact: we are in check if the code comes to this comment
			}

			// Hit any other enemy piece
			if boardhelper.IsIndexBitSet(rayIndex, otherEnemyPiecesBitboard) {
				break // Can just exit, no matter if we hit our own piece first, there is an enemy piece in the way of possible attackers
			}

			rayIndex += offset
			depth++
		}
	}

	return pinnedPieces
}

/*
	Valid move squares
*/

func CheckMask(b *board.Board, colorToMove uint8) (int, uint64) {
	checkCnt := 0
	var validMoveMask uint64

	enemyColor := piece.Black
	if colorToMove == piece.Black {
		enemyColor = piece.White
	}

	indexOffsetsPawns := []int{7, 9}

	if colorToMove == piece.Black {
		indexOffsetsPawns = []int{-7, -9}
	}

	kingIndex := bits.TrailingZeros64(b.Bitboards[colorToMove|piece.King])

	enemyStraightAttackers := b.Bitboards[enemyColor|piece.Rook] | b.Bitboards[enemyColor|piece.Queen]
	enemyDiagonalAttackers := b.Bitboards[enemyColor|piece.Bishop] | b.Bitboards[enemyColor|piece.Queen]
	enemyPawnsBitboard := b.Bitboards[enemyColor|piece.Pawn]

	otherPiecesStraight := ComputedOccupancy[2] & ^enemyStraightAttackers
	otherPiecesDiagonal := ComputedOccupancy[2] & ^enemyDiagonalAttackers

	knightAttackMask := ComputedKnightMoves[kingIndex] & b.Bitboards[enemyColor|piece.Knight]

	for i, offset := range DirectionalOffsets[:4] {
		var currentOffsetBitboard uint64
		rayIndex := kingIndex + offset

		depth := 1
		for depth <= DistanceToEdge[kingIndex][i] {
			currentOffsetBitboard |= 1 << rayIndex

			if boardhelper.IsIndexBitSet(rayIndex, otherPiecesStraight) {
				break
			}

			if boardhelper.IsIndexBitSet(rayIndex, enemyStraightAttackers) {
				validMoveMask |= currentOffsetBitboard
				checkCnt++
				break
			}

			rayIndex += offset
			depth++
		}
	}

	for i, offset := range DirectionalOffsets[4:] {
		var currentOffsetBitboard uint64
		rayIndex := kingIndex + offset

		depth := 1
		for depth <= DistanceToEdge[kingIndex][i+4] {
			currentOffsetBitboard |= 1 << rayIndex

			if boardhelper.IsIndexBitSet(rayIndex, otherPiecesDiagonal) {
				break
			}

			if boardhelper.IsIndexBitSet(rayIndex, enemyDiagonalAttackers) {
				validMoveMask |= currentOffsetBitboard
				checkCnt++
				break
			}

			rayIndex += offset
			depth++
		}
	}

	if knightAttackMask != 0 {
		validMoveMask |= knightAttackMask
		checkCnt += bits.OnesCount64(knightAttackMask)
	}

	for _, offset := range indexOffsetsPawns {
		rayIndex := kingIndex + offset

		if !boardhelper.IsValidDiagonalMove(kingIndex, rayIndex) {
			continue
		}

		if !boardhelper.IsIndexBitSet(rayIndex, enemyPawnsBitboard) {
			continue
		}

		validMoveMask |= 1 << rayIndex
		checkCnt++
	}

	return checkCnt, validMoveMask
}

func IsPinnedMoveAlongRay(b *board.Board, colorToMove uint8, m move.Move) bool {
	var rayBitboard uint64

	ownKingIndex := bits.TrailingZeros64(b.Bitboards[colorToMove|piece.King])
	enemyPieces := ComputedOccupancy[1-((colorToMove>>3)-1)]

	rayOffset := boardhelper.CalculateRayOffset(ownKingIndex, m.StartIndex)

	rayIndex := ownKingIndex + rayOffset

	for boardhelper.IsValidStraightMove(ownKingIndex, rayIndex) || boardhelper.IsValidDiagonalMove(ownKingIndex, rayIndex) {
		// Cannot move to our own square
		if rayIndex != m.StartIndex {
			rayBitboard |= 1 << rayIndex
		}

		// Since we know we are pinned, the first enemy piece has to be our attacker
		if boardhelper.IsIndexBitSet(rayIndex, enemyPieces) {
			break
		}

		rayIndex += rayOffset
	}

	return boardhelper.IsIndexBitSet(m.TargetIndex, rayBitboard)
}

func IsEnPassantMovePinned(b *board.Board, colorToMove uint8, m move.Move) bool {

	enemyColor := piece.Black
	if colorToMove != piece.White {
		enemyColor = piece.White
	}

	ownKingIndex := bits.TrailingZeros64(b.Bitboards[colorToMove|piece.King])

	// Can instantly return if there is no direct ray between ownKingIndex & enPassantCaptureSquare
	offset := boardhelper.CalculateRayOffset(ownKingIndex, m.EnPassantCaptureSquare)
	if offset == 0 {
		return false
	}

	enemyAttackers := b.Bitboards[enemyColor|piece.Queen]
	isValidMoveFunction := boardhelper.IsValidStraightMove
	switch offset {
	case -1, 1, -8, 8:
		enemyAttackers |= b.Bitboards[enemyColor|piece.Rook]
	case -7, 7, -9, 9:
		enemyAttackers |= b.Bitboards[enemyColor|piece.Bishop]
		isValidMoveFunction = boardhelper.IsValidDiagonalMove
	default:
		return false
	}

	otherPieces := ComputedOccupancy[2] & ^(enemyAttackers)

	rayIndex := ownKingIndex + offset

	for isValidMoveFunction(ownKingIndex, rayIndex) {

		// Return if we hit the moves target index (new blocking piece)
		if rayIndex == m.TargetIndex {
			return false
		}

		// Return if we hit an enemy attacker
		if boardhelper.IsIndexBitSet(rayIndex, enemyAttackers) {
			return true
		}

		// Return if we hit any piece that is not on either enPassantCaptureSquare or m.StartIndex
		if rayIndex != m.EnPassantCaptureSquare &&
			rayIndex != m.StartIndex &&
			boardhelper.IsIndexBitSet(rayIndex, otherPieces) {
			return false
		}

		rayIndex += offset
	}

	// Ray was cast until the edge, we can return false
	return false
}

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
