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
		OrthogonalSlidingAttacks(b) |
		DiagonalSlidingAttacks(b) |
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

func OrthogonalSlidingAttacks(b *board.Board) uint64 {
	var attacks uint64

	blockers := Occupancy[2] & ^(1 << b.FriendlyKingIndex)

	pieces := b.Bitboards[b.OpponentColor|piece.Rook] | b.Bitboards[b.OpponentColor|piece.Queen]
	for pieces != 0 {
		pieceIndex := bits.TrailingZeros64(pieces)

		entry := RookMagics[pieceIndex]
		index := int(((blockers & entry.Mask) * entry.Magic) >> (64 - entry.Shift))

		attacks |= RookMoveTable[entry.Offset+index]

		pieces &= pieces - 1
	}

	return attacks
}

func DiagonalSlidingAttacks(b *board.Board) uint64 {
	var attacks uint64

	blockers := Occupancy[2] & ^(1 << b.FriendlyKingIndex)

	pieces := b.Bitboards[b.OpponentColor|piece.Bishop] | b.Bitboards[b.OpponentColor|piece.Queen]
	for pieces != 0 {
		pieceIndex := bits.TrailingZeros64(pieces)

		entry := BishopMagics[pieceIndex]
		index := int(((blockers & entry.Mask) * entry.Magic) >> (64 - entry.Shift))

		attacks |= BishopMoveTable[entry.Offset+index]

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
