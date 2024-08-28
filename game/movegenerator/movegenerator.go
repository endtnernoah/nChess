package movegenerator

import (
	"endtner.dev/nChess/game/board"
	"endtner.dev/nChess/game/boardhelper"
	"endtner.dev/nChess/game/move"
	"endtner.dev/nChess/game/piece"
	"math/bits"
)

func GeneratePawnMoves(b *board.Board, pieceColor uint, enPassantTargetSquare int) []move.Move {
	var pawnMoves []move.Move

	// Default offset for white
	pawnIndexOffset := 8
	promotionRank := 7

	// Change offset for black
	if pieceColor != piece.ColorWhite {
		pawnIndexOffset = -8
		promotionRank = 0
	}

	enemyOccupiedFields := b.Pieces[1-((pieceColor>>3)-1)]
	occupiedFields := b.Pieces[2]

	pawnBitboard := b.PieceBitboard(pieceColor | piece.TypePawn)
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

func GenerateStraightSlidingMoves(b *board.Board, pieceColor uint) []move.Move {
	var straightSlidingMoves []move.Move

	indexOffsets := []int{-1, 1, -8, 8}

	ownOccupiedFields := b.Pieces[(pieceColor>>3)-1]
	enemyOccupiedFields := b.Pieces[1-((pieceColor>>3)-1)]

	piecesBitboard := b.PieceBitboard(pieceColor|piece.TypeRook) | b.PieceBitboard(pieceColor|piece.TypeQueen)
	for piecesBitboard != 0 {
		startIndex := bits.TrailingZeros64(piecesBitboard)

		// Go as deep as possible
		for _, offset := range indexOffsets {
			targetIndex := startIndex + offset

			depth := 1
			for depth <= 7 && boardhelper.IsValidStraightMove(startIndex, targetIndex) && !boardhelper.IsIndexBitSet(targetIndex, ownOccupiedFields) {

				straightSlidingMoves = append(straightSlidingMoves, move.New(startIndex, targetIndex))

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

	return straightSlidingMoves
}

func GenerateDiagonalSlidingMoves(b *board.Board, pieceColor uint) []move.Move {
	var diagonalSlidingMoves []move.Move

	indexOffsets := []int{-7, 7, -9, 9}

	ownOccupiedFields := b.Pieces[(pieceColor>>3)-1]
	enemyOccupiedFields := b.Pieces[1-((pieceColor>>3)-1)]

	piecesBitboard := b.PieceBitboard(pieceColor|piece.TypeBishop) | b.PieceBitboard(pieceColor|piece.TypeQueen)
	for piecesBitboard != 0 {
		startIndex := bits.TrailingZeros64(piecesBitboard)

		// Go as deep as possible
		for _, offset := range indexOffsets {
			targetIndex := startIndex + offset

			depth := 1
			for depth <= 7 && boardhelper.IsValidDiagonalMove(startIndex, targetIndex) && !boardhelper.IsIndexBitSet(targetIndex, ownOccupiedFields) {

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

func GenerateKnightMoves(b *board.Board, pieceColor uint) []move.Move {
	var knightMoves []move.Move

	indexOffsets := []int{-6, 6, -10, 10, -15, 15, -17, 17}

	ownOccupiedFields := b.Pieces[(pieceColor>>3)-1]

	piecesBitboard := b.PieceBitboard(pieceColor | piece.TypeKnight)
	for piecesBitboard != 0 {
		// Get index of LSB
		startIndex := bits.TrailingZeros64(piecesBitboard)

		for _, offset := range indexOffsets {
			targetIndex := startIndex + offset

			if !boardhelper.IsValidKnightMove(startIndex, targetIndex) || boardhelper.IsIndexBitSet(targetIndex, ownOccupiedFields) {
				continue
			}

			knightMoves = append(knightMoves, move.New(startIndex, targetIndex))
		}

		piecesBitboard &= piecesBitboard - 1
	}

	return knightMoves
}

func GenerateKingMoves(b *board.Board, pieceColor uint, castlingAvailability uint) []move.Move {
	var kingMoves []move.Move

	indexOffsetsStraight := []int{-1, 1, -8, 8}
	indexOffsetsDiagonal := []int{-7, 7, -9, 9}

	kingBitboard := b.PieceBitboard(pieceColor | piece.TypeKing)
	rookBitboard := b.PieceBitboard(pieceColor | piece.TypeRook)
	startIndex := bits.TrailingZeros64(kingBitboard)

	allOccupiedFields := b.Pieces[2]
	ownOccupiedFields := b.Pieces[(pieceColor>>3)-1]
	enemyAttackFields := b.AttackFields[1-((pieceColor>>3)-1)]

	for _, offset := range indexOffsetsStraight {
		targetIndex := startIndex + offset

		if !boardhelper.IsValidStraightMove(startIndex, targetIndex) ||
			boardhelper.IsIndexBitSet(targetIndex, ownOccupiedFields) ||
			boardhelper.IsIndexBitSet(targetIndex, enemyAttackFields) {
			continue
		}

		kingMoves = append(kingMoves, move.New(startIndex, targetIndex))
	}

	for _, offset := range indexOffsetsDiagonal {
		targetIndex := startIndex + offset

		if !boardhelper.IsValidDiagonalMove(startIndex, targetIndex) ||
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
