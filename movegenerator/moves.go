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

/*
	Generators
*/

func PawnMoves(b *board.Board, friendlyColor uint8, enPassantTargetSquare int) []move.Move {
	var moves []move.Move

	// Default offset for white
	offset := 8
	promotionRank := 7

	// Change offset for black
	if friendlyColor != piece.White {
		offset = -8
		promotionRank = 0
	}

	opponentPiecesMask := Occupancy[1-((friendlyColor>>3)-1)]
	allPiecesMask := Occupancy[2]

	pieces := b.Bitboards[friendlyColor|piece.Pawn]
	for pieces != 0 {
		// Get index of LSB
		pieceIndex := bits.TrailingZeros64(pieces)
		targetIndex := pieceIndex + offset

		// Continue if target index is out of bounds, just go to the next iteration
		if targetIndex < 0 || targetIndex > 63 {
			pieces &= pieces - 1
			continue
		}

		promotionFlag := targetIndex/8 == promotionRank

		// Add move if target square is empty
		if !boardhelper.IsIndexBitSet(targetIndex, allPiecesMask) {
			// Add all promotion moves
			if promotionFlag {
				moves = append(moves, move.New(pieceIndex, targetIndex, move.WithPromotion(friendlyColor|piece.Knight)))
				moves = append(moves, move.New(pieceIndex, targetIndex, move.WithPromotion(friendlyColor|piece.Bishop)))
				moves = append(moves, move.New(pieceIndex, targetIndex, move.WithPromotion(friendlyColor|piece.Queen)))
				moves = append(moves, move.New(pieceIndex, targetIndex, move.WithPromotion(friendlyColor|piece.Rook)))
			} else {
				moves = append(moves, move.New(pieceIndex, targetIndex))
			}
		}

		// Check if the pawn is on starting square
		isStartingSquare := pieceIndex >= 8 && pieceIndex < 16
		if friendlyColor != piece.White {
			isStartingSquare = pieceIndex >= 48 && pieceIndex < 56
		}

		// Can move 2 rows from starting square
		if isStartingSquare {
			if !boardhelper.IsIndexBitSet(targetIndex, allPiecesMask) &&
				!boardhelper.IsIndexBitSet(targetIndex+offset, allPiecesMask) {
				moves = append(moves, move.New(pieceIndex, targetIndex+offset, move.WithEnPassantPassedSquare(targetIndex)))
			}
		}

		// Check for possible captures
		side1 := targetIndex - 1
		side2 := targetIndex + 1

		// Add move if it is in the same target row and targets are not empty
		if boardhelper.IsValidDiagonalMove(pieceIndex, side1) && boardhelper.IsIndexBitSet(side1, opponentPiecesMask) {
			if promotionFlag {
				moves = append(moves, move.New(pieceIndex, side1, move.WithPromotion(friendlyColor|piece.Knight)))
				moves = append(moves, move.New(pieceIndex, side1, move.WithPromotion(friendlyColor|piece.Bishop)))
				moves = append(moves, move.New(pieceIndex, side1, move.WithPromotion(friendlyColor|piece.Queen)))
				moves = append(moves, move.New(pieceIndex, side1, move.WithPromotion(friendlyColor|piece.Rook)))
			} else {
				moves = append(moves, move.New(pieceIndex, side1))
			}
		}
		if boardhelper.IsValidDiagonalMove(pieceIndex, side2) && boardhelper.IsIndexBitSet(side2, opponentPiecesMask) {
			if promotionFlag {
				moves = append(moves, move.New(pieceIndex, side2, move.WithPromotion(friendlyColor|piece.Knight)))
				moves = append(moves, move.New(pieceIndex, side2, move.WithPromotion(friendlyColor|piece.Bishop)))
				moves = append(moves, move.New(pieceIndex, side2, move.WithPromotion(friendlyColor|piece.Queen)))
				moves = append(moves, move.New(pieceIndex, side2, move.WithPromotion(friendlyColor|piece.Rook)))
			} else {
				moves = append(moves, move.New(pieceIndex, side2))
			}
		}

		// Add move if either side can capture en passant
		if boardhelper.IsValidDiagonalMove(pieceIndex, side1) && side1 == enPassantTargetSquare {
			moves = append(moves, move.New(pieceIndex, side1, move.WithEnPassantCaptureSquare(side1-offset)))
		}
		if boardhelper.IsValidDiagonalMove(pieceIndex, side2) && side2 == enPassantTargetSquare {
			moves = append(moves, move.New(pieceIndex, side2, move.WithEnPassantCaptureSquare(side2-offset)))
		}

		// Remove LSB of bitboard
		pieces &= pieces - 1
	}

	return moves
}

func SlidingMoves(b *board.Board, friendlyColor uint8) []move.Move {
	var moves []move.Move

	friendlyPiecesMask := Occupancy[(friendlyColor>>3)-1]
	opponentPiecesMask := Occupancy[1-((friendlyColor>>3)-1)]

	pieces := b.Bitboards[friendlyColor|piece.Rook] | b.Bitboards[friendlyColor|piece.Bishop] | b.Bitboards[friendlyColor|piece.Queen]
	for pieces != 0 {
		pieceIndex := bits.TrailingZeros64(pieces)
		pieceType := b.Pieces[pieceIndex] & 0b00111

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
			targetIndex := pieceIndex + offset

			depth := 1
			for depth <= DistanceToEdge[pieceIndex][i+offsetIndexStart] {
				if boardhelper.IsIndexBitSet(targetIndex, friendlyPiecesMask) {
					break
				}

				moves = append(moves, move.New(pieceIndex, targetIndex))

				// Break if we captured an enemy
				if boardhelper.IsIndexBitSet(targetIndex, opponentPiecesMask) {
					break
				}

				targetIndex += offset
				depth++
			}
		}

		pieces &= pieces - 1
	}

	return moves
}

func KnightMoves(b *board.Board, friendlyColor uint8) []move.Move {
	var moves []move.Move

	friendlyPiecesMask := Occupancy[(friendlyColor>>3)-1]

	knights := b.Bitboards[friendlyColor|piece.Knight]
	for knights != 0 {
		startIndex := bits.TrailingZeros64(knights)
		validMoveMask := ComputedKnightMoves[startIndex] & ^friendlyPiecesMask

		for validMoveMask != 0 {
			moves = append(moves, move.New(startIndex, bits.TrailingZeros64(validMoveMask)))
			validMoveMask &= validMoveMask - 1
		}

		knights &= knights - 1
	}

	return moves
}

func KingMoves(b *board.Board, friendlyColor uint8, castlingAvailability uint8) []move.Move {
	var moves []move.Move

	friendlyKingIndex := bits.TrailingZeros64(b.Bitboards[friendlyColor|piece.King])
	friendlyRooks := b.Bitboards[friendlyColor|piece.Rook]

	allPiecesMask := Occupancy[2]
	friendlyPiecesMask := Occupancy[(friendlyColor>>3)-1]
	opponentAttackMask := Attacks[1-((friendlyColor>>3)-1)]

	for i, offset := range DirectionalOffsets {
		targetIndex := friendlyKingIndex + offset

		if targetIndex < 0 || targetIndex > 63 {
			continue
		}

		if DistanceToEdge[friendlyKingIndex][i] == 0 ||
			boardhelper.IsIndexBitSet(targetIndex, friendlyPiecesMask) ||
			boardhelper.IsIndexBitSet(targetIndex, opponentAttackMask) {
			continue
		}

		moves = append(moves, move.New(friendlyKingIndex, targetIndex))
	}

	initialKingIndex := 4
	if friendlyColor != piece.White {
		initialKingIndex = 60
	}

	// King is not on its original square, will not be allowed to castle
	if friendlyKingIndex != initialKingIndex {
		return moves
	}

	kingSideAllowed := castlingAvailability&0b1000 != 0
	queenSideAllowed := castlingAvailability&0b0100 != 0

	var kingSideEmptyMask uint64 = 0b1100000
	var queenSideEmptyMask uint64 = 0b1110

	kingSideRookIndex := 7
	queenSideRookIndex := 0

	var kingSideAttackMask uint64 = 1<<friendlyKingIndex | 1<<(friendlyKingIndex+1) | 1<<(friendlyKingIndex+2)
	var queenSideAttackMask uint64 = 1<<friendlyKingIndex | 1<<(friendlyKingIndex-1) | 1<<(friendlyKingIndex-2)

	if friendlyColor != piece.White {
		kingSideAllowed = castlingAvailability&0b0010 != 0
		queenSideAllowed = castlingAvailability&0b0001 != 0

		kingSideEmptyMask <<= 56
		queenSideEmptyMask <<= 56

		kingSideRookIndex += 56
		queenSideRookIndex += 56
	}

	if kingSideAllowed &&
		(kingSideAttackMask&opponentAttackMask) == 0 && // King does not start or pass through attacked field
		(kingSideEmptyMask&allPiecesMask) == 0 && // All fields are empty
		boardhelper.IsIndexBitSet(kingSideRookIndex, friendlyRooks) { // There is a rook on its field
		moves = append(moves, move.New(friendlyKingIndex, friendlyKingIndex+2, move.WithRookStartingSquare(kingSideRookIndex)))
	}

	if queenSideAllowed &&
		(queenSideAttackMask&opponentAttackMask) == 0 &&
		(queenSideEmptyMask&allPiecesMask) == 0 &&
		boardhelper.IsIndexBitSet(queenSideRookIndex, friendlyRooks) {
		moves = append(moves, move.New(friendlyKingIndex, friendlyKingIndex-2, move.WithRookStartingSquare(queenSideRookIndex)))
	}

	return moves
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
