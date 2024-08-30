package movegenerator

import (
	"endtner.dev/nChess/board"
	"endtner.dev/nChess/board/boardhelper"
	"endtner.dev/nChess/board/piece"
	"math/bits"
)

var Attacks = make([]uint64, 2)

func ComputeAttacks(b *board.Board) {
	Attacks[b.OpponentIndex] = PawnAttacks(b) |
		SlidingAttacks(b) |
		KnightAttacks(b) |
		ComputedKingMoves[b.OpponentKingIndex]
}

/*
	Generators
*/

func PawnAttacks(b *board.Board) uint64 {
	var attacks uint64

	offset := b.PawnOffset * -1

	pieces := b.Bitboards[b.OpponentColor|piece.Pawn]
	for pieces != 0 {
		pieceIndex := bits.TrailingZeros64(pieces)

		side1 := pieceIndex + offset - 1
		side2 := pieceIndex + offset + 1

		// Add move if it is in the same target row and targets are not empty
		if boardhelper.IsValidDiagonalMove(pieceIndex, side1) {
			attacks |= 1 << side1
		}
		if boardhelper.IsValidDiagonalMove(pieceIndex, side2) {
			attacks |= 1 << side2
		}

		pieces &= pieces - 1
	}

	return attacks
}

func SlidingAttacks(b *board.Board) uint64 {
	var attacks uint64

	friendlyPiecesMask := Occupancy[b.OpponentIndex]
	opponentPiecesMask := Occupancy[b.FriendlyIndex]

	opponentKingIndex := b.FriendlyKingIndex

	pieces := b.Bitboards[b.OpponentColor|piece.Rook] | b.Bitboards[b.OpponentColor|piece.Bishop] | b.Bitboards[b.OpponentColor|piece.Queen]
	for pieces != 0 {
		pieceIndex := bits.TrailingZeros64(pieces)
		pieceType := b.Pieces[pieceIndex] & 0b00111

		offsetIndexStart := 0
		offsetIndexEnd := 8

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

				attacks |= 1 << targetIndex

				// If we hit our own piece, we break
				if boardhelper.IsIndexBitSet(targetIndex, friendlyPiecesMask) {
					break
				}

				// If we hit an enemy piece except the enemy king, we break
				if boardhelper.IsIndexBitSet(targetIndex, opponentPiecesMask) && (targetIndex != opponentKingIndex) {
					break
				}

				targetIndex += offset
				depth++
			}

		}

		pieces &= pieces - 1
	}

	return attacks
}

func KnightAttacks(b *board.Board) uint64 {
	var attacks uint64

	knights := b.Bitboards[b.OpponentColor|piece.Knight]
	for knights != 0 {
		attacks |= ComputedKnightMoves[bits.TrailingZeros64(knights)]
		knights &= knights - 1
	}

	return attacks
}
