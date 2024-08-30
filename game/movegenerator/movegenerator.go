package movegenerator

import (
	"endtner.dev/nChess/game/board"
	"endtner.dev/nChess/game/boardhelper"
	"endtner.dev/nChess/game/move"
	"endtner.dev/nChess/game/piece"
	"math/bits"
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

// ComputedAttacks & ComputedPins can be accessed at index (3 >> color) - 1 & 1 - accessIndex for enemy color
var ComputedAttacks = make([]uint64, 2)
var ComputedPins = make([]uint64, 2)

func Min(x int, y int) int {
	if x < y {
		return x
	}
	return y
}

func ComputeAll(b *board.Board) {
	ComputeAttacks(b, piece.ColorWhite)
	ComputeAttacks(b, piece.ColorBlack)
	ComputePins(b, piece.ColorWhite)
	ComputePins(b, piece.ColorBlack)
}

/*
	Generators
*/

func PawnMoves(b *board.Board, pieceColor uint, enPassantTargetSquare int) []move.Move {
	var pawnMoves []move.Move

	// Default offset for white
	pawnIndexOffset := 8
	promotionRank := 7

	// Change offset for black
	if pieceColor != piece.ColorWhite {
		pawnIndexOffset = -8
		promotionRank = 0
	}

	enemyOccupiedFields := b.OccupancyBitboards[1-((pieceColor>>3)-1)]
	occupiedFields := b.OccupancyBitboards[2]

	pawnBitboard := b.Bitboards[pieceColor|piece.TypePawn]
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
				pawnMoves = append(pawnMoves, move.New(startIndex, targetIndex, move.WithPromotion(pieceColor|piece.TypeKnight)))
				pawnMoves = append(pawnMoves, move.New(startIndex, targetIndex, move.WithPromotion(pieceColor|piece.TypeBishop)))
				pawnMoves = append(pawnMoves, move.New(startIndex, targetIndex, move.WithPromotion(pieceColor|piece.TypeQueen)))
				pawnMoves = append(pawnMoves, move.New(startIndex, targetIndex, move.WithPromotion(pieceColor|piece.TypeRook)))
			} else {
				pawnMoves = append(pawnMoves, move.New(startIndex, targetIndex))
			}
		}

		// Check if the pawn is on starting square
		isStartingSquare := startIndex >= 8 && startIndex < 16
		if pieceColor != piece.ColorWhite {
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
				pawnMoves = append(pawnMoves, move.New(startIndex, side1, move.WithPromotion(pieceColor|piece.TypeKnight)))
				pawnMoves = append(pawnMoves, move.New(startIndex, side1, move.WithPromotion(pieceColor|piece.TypeBishop)))
				pawnMoves = append(pawnMoves, move.New(startIndex, side1, move.WithPromotion(pieceColor|piece.TypeQueen)))
				pawnMoves = append(pawnMoves, move.New(startIndex, side1, move.WithPromotion(pieceColor|piece.TypeRook)))
			} else {
				pawnMoves = append(pawnMoves, move.New(startIndex, side1))
			}
		}
		if boardhelper.IsValidDiagonalMove(startIndex, side2) && boardhelper.IsIndexBitSet(side2, enemyOccupiedFields) {
			if promotionFlag {
				pawnMoves = append(pawnMoves, move.New(startIndex, side2, move.WithPromotion(pieceColor|piece.TypeKnight)))
				pawnMoves = append(pawnMoves, move.New(startIndex, side2, move.WithPromotion(pieceColor|piece.TypeBishop)))
				pawnMoves = append(pawnMoves, move.New(startIndex, side2, move.WithPromotion(pieceColor|piece.TypeQueen)))
				pawnMoves = append(pawnMoves, move.New(startIndex, side2, move.WithPromotion(pieceColor|piece.TypeRook)))
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

func SlidingMoves(b *board.Board, pieceColor uint) []move.Move {
	var slidingMoves []move.Move

	ownOccupiedFields := b.OccupancyBitboards[(pieceColor>>3)-1]
	enemyOccupiedFields := b.OccupancyBitboards[1-((pieceColor>>3)-1)]

	piecesBitboard := b.Bitboards[pieceColor|piece.TypeRook] | b.Bitboards[pieceColor|piece.TypeBishop] | b.Bitboards[pieceColor|piece.TypeQueen]
	for piecesBitboard != 0 {
		startIndex := bits.TrailingZeros64(piecesBitboard)
		pieceType := b.Pieces[startIndex] & 0b00111

		offsetIndexStart := 0
		offsetIndexEnd := 8

		// Manipulating indices based on piece type
		if pieceType == piece.TypeRook {
			offsetIndexEnd = 4
		}
		if pieceType == piece.TypeBishop {
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

func StraightSlidingMoves(b *board.Board, pieceColor uint) []move.Move {
	var straightSlidingMoves []move.Move

	ownOccupiedFields := b.OccupancyBitboards[(pieceColor>>3)-1]
	enemyOccupiedFields := b.OccupancyBitboards[1-((pieceColor>>3)-1)]

	piecesBitboard := b.Bitboards[pieceColor|piece.TypeRook] | b.Bitboards[pieceColor|piece.TypeQueen]
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

func DiagonalSlidingMoves(b *board.Board, pieceColor uint) []move.Move {
	var diagonalSlidingMoves []move.Move

	ownOccupiedFields := b.OccupancyBitboards[(pieceColor>>3)-1]
	enemyOccupiedFields := b.OccupancyBitboards[1-((pieceColor>>3)-1)]

	piecesBitboard := b.Bitboards[pieceColor|piece.TypeBishop] | b.Bitboards[pieceColor|piece.TypeQueen]
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

func KnightMoves(b *board.Board, pieceColor uint) []move.Move {
	var validMoves []move.Move

	ownOccupiedFields := b.OccupancyBitboards[(pieceColor>>3)-1]

	knights := b.Bitboards[pieceColor|piece.TypeKnight]
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

func KingMoves(b *board.Board, pieceColor uint, castlingAvailability uint) []move.Move {
	var kingMoves []move.Move

	kingBitboard := b.Bitboards[pieceColor|piece.TypeKing]
	rookBitboard := b.Bitboards[pieceColor|piece.TypeRook]
	startIndex := bits.TrailingZeros64(kingBitboard)

	allOccupiedFields := b.OccupancyBitboards[2]
	ownOccupiedFields := b.OccupancyBitboards[(pieceColor>>3)-1]
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
	if pieceColor != piece.ColorWhite {
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

	if pieceColor != piece.ColorWhite {
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

func ComputeAttacks(b *board.Board, color uint) {
	pawnAttacks := PawnAttacks(b, color)
	straightSlidingAttacks := StraightSlidingAttacks(b, color)
	diagonalSlidingAttacks := DiagonalSlidingAttacks(b, color)
	knightAttacks := KnightAttacks(b, color)

	ComputedAttacks[(color>>3)-1] = pawnAttacks |
		straightSlidingAttacks |
		diagonalSlidingAttacks |
		knightAttacks
}

func PawnAttacks(b *board.Board, colorToMove uint) uint64 {
	var pawnAttacks uint64

	pawnIndexOffset := 8

	if colorToMove != piece.ColorWhite {
		pawnIndexOffset = -8
	}

	pieceBitboard := b.Bitboards[colorToMove|piece.TypePawn]

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

func StraightSlidingAttacks(b *board.Board, colorToMove uint) uint64 {
	straightIndexOffsets := []int{1, -1, 8, -8}
	var straightSlidingAttacks uint64

	pieceBitboard := b.Bitboards[colorToMove|piece.TypeRook] | b.Bitboards[colorToMove|piece.TypeQueen] | b.Bitboards[colorToMove|piece.TypeKing]
	ownPiecesBitboard := b.OccupancyBitboards[(colorToMove>>3)-1]
	enemyPiecesBitboard := b.OccupancyBitboards[1-((colorToMove>>3)-1)]

	enemyColor := piece.ColorBlack
	if colorToMove != piece.ColorWhite {
		enemyColor = piece.ColorWhite
	}
	enemyKingIndex := bits.TrailingZeros64(b.Bitboards[enemyColor|piece.TypeKing])

	for pieceBitboard != 0 {
		startIndex := bits.TrailingZeros64(pieceBitboard)

		maxLength := 8
		if (b.Pieces[startIndex] & 0b00111) == piece.TypeKing {
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

func DiagonalSlidingAttacks(b *board.Board, colorToMove uint) uint64 {
	diagonalIndexOffsets := []int{-7, 7, -9, 9}
	var diagonalSlidingAttacks uint64

	pieceBitboard := b.Bitboards[colorToMove|piece.TypeBishop] | b.Bitboards[colorToMove|piece.TypeQueen] | b.Bitboards[colorToMove|piece.TypeKing]
	ownPiecesBitboard := b.OccupancyBitboards[(colorToMove>>3)-1]
	enemyPiecesBitboard := b.OccupancyBitboards[1-((colorToMove>>3)-1)]

	enemyColor := piece.ColorBlack
	if colorToMove != piece.ColorWhite {
		enemyColor = piece.ColorWhite
	}
	enemyKingIndex := bits.TrailingZeros64(b.Bitboards[enemyColor|piece.TypeKing])

	for pieceBitboard != 0 {
		startIndex := bits.TrailingZeros64(pieceBitboard)

		maxLength := 8
		if (b.Pieces[startIndex] & 0b00111) == piece.TypeKing {
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

func KnightAttacks(b *board.Board, colorToMove uint) uint64 {
	var attacks uint64

	knights := b.Bitboards[colorToMove|piece.TypeKnight]
	for knights != 0 {
		attacks |= ComputedKnightMoves[bits.TrailingZeros64(knights)]
		knights &= knights - 1
	}

	return attacks
}

/*
	Calculating pins
*/

func ComputePins(b *board.Board, color uint) {
	straightPinnedPieces := StraightPins(b, color)
	diagonalPinnedPieces := DiagonalPins(b, color)

	ComputedPins[(color>>3)-1] = straightPinnedPieces | diagonalPinnedPieces
}

func StraightPins(b *board.Board, colorToMove uint) uint64 {
	var pinnedPieces uint64

	kingIndex := bits.TrailingZeros64(b.Bitboards[colorToMove|piece.TypeKing])

	ownPiecesBitboard := b.OccupancyBitboards[(colorToMove>>3)-1]
	enemyPiecesBitboard := b.OccupancyBitboards[1-((colorToMove>>3)-1)]

	enemyColor := piece.ColorWhite
	if colorToMove == piece.ColorWhite {
		enemyColor = piece.ColorBlack
	}

	attackingEnemyPiecesBitboard := b.Bitboards[enemyColor|piece.TypeRook] | b.Bitboards[enemyColor|piece.TypeQueen]
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

func DiagonalPins(b *board.Board, colorToMove uint) uint64 {
	var pinnedPieces uint64

	kingIndex := bits.TrailingZeros64(b.Bitboards[colorToMove|piece.TypeKing])
	ownPiecesBitboard := b.OccupancyBitboards[(colorToMove>>3)-1]
	enemyPiecesBitboard := b.OccupancyBitboards[1-((colorToMove>>3)-1)]

	enemyColor := piece.ColorWhite
	if colorToMove == piece.ColorWhite {
		enemyColor = piece.ColorBlack
	}

	attackingEnemyPiecesBitboard := b.Bitboards[enemyColor|piece.TypeBishop] | b.Bitboards[enemyColor|piece.TypeQueen]
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

func CheckMask(b *board.Board, colorToMove uint) (int, uint64) {
	checkCnt := 0
	var validMoveMask uint64

	enemyColor := piece.ColorBlack
	if colorToMove == piece.ColorBlack {
		enemyColor = piece.ColorWhite
	}

	indexOffsetsPawns := []int{7, 9}

	if colorToMove == piece.ColorBlack {
		indexOffsetsPawns = []int{-7, -9}
	}

	kingIndex := bits.TrailingZeros64(b.Bitboards[colorToMove|piece.TypeKing])

	enemyStraightAttackers := b.Bitboards[enemyColor|piece.TypeRook] | b.Bitboards[enemyColor|piece.TypeQueen]
	enemyDiagonalAttackers := b.Bitboards[enemyColor|piece.TypeBishop] | b.Bitboards[enemyColor|piece.TypeQueen]
	enemyPawnsBitboard := b.Bitboards[enemyColor|piece.TypePawn]

	otherPiecesStraight := b.OccupancyBitboards[2] & ^enemyStraightAttackers
	otherPiecesDiagonal := b.OccupancyBitboards[2] & ^enemyDiagonalAttackers

	knightAttackMask := ComputedKnightMoves[kingIndex] & b.Bitboards[enemyColor|piece.TypeKnight]

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
