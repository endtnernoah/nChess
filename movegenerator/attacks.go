package movegenerator

import (
	"endtner.dev/nChess/board"
	"endtner.dev/nChess/board/boardhelper"
	"endtner.dev/nChess/board/piece"
	"math/bits"
)

var Attacks = make([]uint64, 2)

func ComputeAttacks(b *board.Board, color uint8) {
	pawnAttacks := PawnAttacks(b, color)
	straightSlidingAttacks := StraightSlidingAttacks(b, color)
	diagonalSlidingAttacks := DiagonalSlidingAttacks(b, color)
	knightAttacks := KnightAttacks(b, color)

	Attacks[(color>>3)-1] = pawnAttacks |
		straightSlidingAttacks |
		diagonalSlidingAttacks |
		knightAttacks
}

/*
	Generators
*/

func PawnAttacks(b *board.Board, friendlyColor uint8) uint64 {
	var attacks uint64

	offset := 8
	if friendlyColor != piece.White {
		offset = -8
	}

	pieces := b.Bitboards[friendlyColor|piece.Pawn]
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

func StraightSlidingAttacks(b *board.Board, friendlyColor uint8) uint64 {
	var attacks uint64

	friendlyPiecesMask := Occupancy[(friendlyColor>>3)-1]
	opponentPiecesMask := Occupancy[1-((friendlyColor>>3)-1)]

	opponentColor := piece.Black
	if friendlyColor != piece.White {
		opponentColor = piece.White
	}
	opponentKingIndex := bits.TrailingZeros64(b.Bitboards[opponentColor|piece.King])

	pieces := b.Bitboards[friendlyColor|piece.Rook] | b.Bitboards[friendlyColor|piece.Queen] | b.Bitboards[friendlyColor|piece.King]
	for pieces != 0 {
		pieceIndex := bits.TrailingZeros64(pieces)

		for i, offset := range DirectionalOffsets[:4] {
			targetIndex := pieceIndex + offset

			depth := 1
			for depth <= DistanceToEdge[pieceIndex][i] {

				attacks |= 1 << targetIndex

				// If we hit our own piece, we break
				if boardhelper.IsIndexBitSet(targetIndex, friendlyPiecesMask) {
					break
				}

				// If we hit an enemy piece except the enemy king, we break
				if boardhelper.IsIndexBitSet(targetIndex, opponentPiecesMask) && (targetIndex != opponentKingIndex) {
					break
				}

				// If we are a king, we break
				if (b.Pieces[pieceIndex] & 0b00111) == piece.King {
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

func DiagonalSlidingAttacks(b *board.Board, friendlyColor uint8) uint64 {
	var attacks uint64

	friendlyPiecesMask := Occupancy[(friendlyColor>>3)-1]
	opponentPiecesMask := Occupancy[1-((friendlyColor>>3)-1)]

	opponentColor := piece.Black
	if friendlyColor != piece.White {
		opponentColor = piece.White
	}
	enemyKingIndex := bits.TrailingZeros64(b.Bitboards[opponentColor|piece.King])

	pieces := b.Bitboards[friendlyColor|piece.Bishop] | b.Bitboards[friendlyColor|piece.Queen] | b.Bitboards[friendlyColor|piece.King]
	for pieces != 0 {
		pieceIndex := bits.TrailingZeros64(pieces)

		for i, offset := range DirectionalOffsets[4:] {
			targetIndex := pieceIndex + offset

			depth := 1
			for depth <= DistanceToEdge[pieceIndex][i+4] {

				attacks |= 1 << targetIndex

				// If we hit our own piece, we break
				if boardhelper.IsIndexBitSet(targetIndex, friendlyPiecesMask) {
					break
				}

				// If we hit an enemy piece except the enemy king, we break
				if boardhelper.IsIndexBitSet(targetIndex, opponentPiecesMask) && (targetIndex != enemyKingIndex) {
					break
				}

				// If we are a king, we break
				if (b.Pieces[pieceIndex] & 0b00111) == piece.King {
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

func KnightAttacks(b *board.Board, friendlyColor uint8) uint64 {
	var attacks uint64

	knights := b.Bitboards[friendlyColor|piece.Knight]
	for knights != 0 {
		attacks |= ComputedKnightMoves[bits.TrailingZeros64(knights)]
		knights &= knights - 1
	}

	return attacks
}
