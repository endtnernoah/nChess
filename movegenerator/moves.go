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

/*
	Generators
*/

func PawnMoves(b *board.Board, moves *[]move.Move) {
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

		// Add move if target square is empty
		if !boardhelper.IsIndexBitSet(targetIndex, allPiecesMask) {
			// Add all promotion moves
			if promotionFlag {
				*moves = append(*moves, move.New(pieceIndex, targetIndex, move.WithPromotion(b.FriendlyColor|piece.Knight)))
				*moves = append(*moves, move.New(pieceIndex, targetIndex, move.WithPromotion(b.FriendlyColor|piece.Bishop)))
				*moves = append(*moves, move.New(pieceIndex, targetIndex, move.WithPromotion(b.FriendlyColor|piece.Queen)))
				*moves = append(*moves, move.New(pieceIndex, targetIndex, move.WithPromotion(b.FriendlyColor|piece.Rook)))
			} else {
				*moves = append(*moves, move.New(pieceIndex, targetIndex))
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
				*moves = append(*moves, move.New(pieceIndex, epTargetIndex, move.WithEnPassantPassedSquare(targetIndex)))
			}
		}

		// Check for possible captures
		side1 := targetIndex - 1
		side2 := targetIndex + 1

		// Add move if it is in the same target row and targets are not empty
		if boardhelper.IsValidDiagonalMove(pieceIndex, side1) && boardhelper.IsIndexBitSet(side1, opponentPiecesMask) {
			if promotionFlag {
				*moves = append(*moves, move.New(pieceIndex, side1, move.WithPromotion(b.FriendlyColor|piece.Knight)))
				*moves = append(*moves, move.New(pieceIndex, side1, move.WithPromotion(b.FriendlyColor|piece.Bishop)))
				*moves = append(*moves, move.New(pieceIndex, side1, move.WithPromotion(b.FriendlyColor|piece.Queen)))
				*moves = append(*moves, move.New(pieceIndex, side1, move.WithPromotion(b.FriendlyColor|piece.Rook)))
			} else {
				*moves = append(*moves, move.New(pieceIndex, side1))
			}
		}
		if boardhelper.IsValidDiagonalMove(pieceIndex, side2) && boardhelper.IsIndexBitSet(side2, opponentPiecesMask) {
			if promotionFlag {
				*moves = append(*moves, move.New(pieceIndex, side2, move.WithPromotion(b.FriendlyColor|piece.Knight)))
				*moves = append(*moves, move.New(pieceIndex, side2, move.WithPromotion(b.FriendlyColor|piece.Bishop)))
				*moves = append(*moves, move.New(pieceIndex, side2, move.WithPromotion(b.FriendlyColor|piece.Queen)))
				*moves = append(*moves, move.New(pieceIndex, side2, move.WithPromotion(b.FriendlyColor|piece.Rook)))
			} else {
				*moves = append(*moves, move.New(pieceIndex, side2))
			}
		}

		// Add move if either side can capture en passant
		if boardhelper.IsValidDiagonalMove(pieceIndex, side1) && side1 == b.EnPassantTargetSquare {
			*moves = append(*moves, move.New(pieceIndex, side1, move.WithEnPassantCaptureSquare(side1-b.PawnOffset)))
		}
		if boardhelper.IsValidDiagonalMove(pieceIndex, side2) && side2 == b.EnPassantTargetSquare {
			*moves = append(*moves, move.New(pieceIndex, side2, move.WithEnPassantCaptureSquare(side2-b.PawnOffset)))
		}

		// Remove LSB of bitboard
		pieces &= pieces - 1
	}
}

func SlidingMoves(b *board.Board, moves *[]move.Move) {
	friendlyPiecesMask := Occupancy[b.FriendlyIndex]
	opponentPiecesMask := Occupancy[b.OpponentIndex]

	pieces := b.Bitboards[b.FriendlyColor|piece.Rook] | b.Bitboards[b.FriendlyColor|piece.Bishop] | b.Bitboards[b.FriendlyColor|piece.Queen]
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

				*moves = append(*moves, move.New(pieceIndex, targetIndex))

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
}

func KnightMoves(b *board.Board, moves *[]move.Move) {
	friendlyPiecesMask := Occupancy[b.FriendlyIndex]

	knights := b.Bitboards[b.FriendlyColor|piece.Knight] & ^Pins[b.FriendlyIndex] // Knights cannot move if pinned
	for knights != 0 {
		pieceIndex := bits.TrailingZeros64(knights)

		validMoveMask := ComputedKnightMoves[pieceIndex] & ^friendlyPiecesMask

		for validMoveMask != 0 {
			*moves = append(*moves, move.New(pieceIndex, bits.TrailingZeros64(validMoveMask)))
			validMoveMask &= validMoveMask - 1
		}

		knights &= knights - 1
	}
}

func KingMoves(b *board.Board, moves *[]move.Move) {
	friendlyKingIndex := b.FriendlyKingIndex
	friendlyRooks := b.Bitboards[b.FriendlyColor|piece.Rook]

	allPiecesMask := Occupancy[2]
	friendlyPiecesMask := Occupancy[b.FriendlyIndex]
	opponentAttackMask := Attacks[b.OpponentIndex]

	kingMoveMask := ComputedKingMoves[friendlyKingIndex] & ^friendlyPiecesMask & ^opponentAttackMask
	for kingMoveMask != 0 {
		*moves = append(*moves, move.New(friendlyKingIndex, bits.TrailingZeros64(kingMoveMask)))
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
		*moves = append(*moves, move.New(friendlyKingIndex, friendlyKingIndex+2, move.WithRookStartingSquare(kingSideRookIndex)))
	}

	if queenSideAllowed &&
		(queenSideAttackMask&opponentAttackMask) == 0 &&
		(queenSideEmptyMask&allPiecesMask) == 0 &&
		boardhelper.IsIndexBitSet(queenSideRookIndex, friendlyRooks) {
		*moves = append(*moves, move.New(friendlyKingIndex, friendlyKingIndex-2, move.WithRookStartingSquare(queenSideRookIndex)))
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
