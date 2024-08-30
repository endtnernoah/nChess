package movegenerator

import (
	"endtner.dev/nChess/board"
	"endtner.dev/nChess/board/boardhelper"
	"endtner.dev/nChess/board/piece"
)

var Pins = make([]uint64, 2)

func ComputePins(b *board.Board) {
	straightPinnedPieces := OrthogonalPins(b)
	diagonalPinnedPieces := DiagonalPins(b)

	Pins[b.FriendlyIndex] = straightPinnedPieces | diagonalPinnedPieces
}

/*
	Generators
*/

func OrthogonalPins(b *board.Board) uint64 {
	var pinMask uint64

	friendlyKingIndex := b.FriendlyKingIndex

	friendlyPiecesMask := Occupancy[b.FriendlyIndex]
	opponentPiecesMask := Occupancy[b.OpponentIndex]

	opponentAttackers := b.Bitboards[b.OpponentColor|piece.Rook] | b.Bitboards[b.OpponentColor|piece.Queen]
	otherOpponentPiecesMask := opponentPiecesMask & ^opponentAttackers

	for i, offset := range DirectionalOffsets[:4] {
		step := friendlyKingIndex + offset
		ownPiece := -1

		if step < 0 || step > 63 {
			continue
		}

		depth := 1
		for depth <= DistanceToEdge[friendlyKingIndex][i] {

			// Hit our own piece
			if boardhelper.IsIndexBitSet(step, friendlyPiecesMask) {

				// Hit own piece 2 times in a row
				if ownPiece != -1 {
					break
				}

				// Hit own piece first time
				ownPiece = step
			}

			// Hit enemy attacking piece
			if boardhelper.IsIndexBitSet(step, opponentAttackers) {

				// If we have hit our own piece before, that piece is pinned
				if ownPiece != -1 {
					pinMask |= 1 << ownPiece
					break
				}

				// Fun fact: we are in check if the code comes to this comment
			}

			// Hit any other enemy piece
			if boardhelper.IsIndexBitSet(step, otherOpponentPiecesMask) {
				break // Can just exit, no matter if we hit our own piece first, there is an enemy piece in the way of possible attackers
			}

			step += offset
			depth++
		}
	}

	return pinMask
}

func DiagonalPins(b *board.Board) uint64 {
	var pinMask uint64

	friendlyKingIndex := b.FriendlyKingIndex
	friendlyPiecesMask := Occupancy[b.FriendlyIndex]
	opponentPiecesMask := Occupancy[b.OpponentIndex]

	opponentAttackers := b.Bitboards[b.OpponentColor|piece.Bishop] | b.Bitboards[b.OpponentColor|piece.Queen]
	otherOpponentPiecesMask := opponentPiecesMask & ^opponentAttackers

	for i, offset := range DirectionalOffsets[4:] {
		step := friendlyKingIndex + offset
		ownPiece := -1

		if step < 0 || step > 63 {
			continue
		}

		// Go as long as the ray moves to valid fields
		depth := 1
		for depth <= DistanceToEdge[friendlyKingIndex][i+4] {
			// Hit our own piece
			if boardhelper.IsIndexBitSet(step, friendlyPiecesMask) {

				// Hit own piece 2 times in a row
				if ownPiece != -1 {
					break
				}

				// Hit own piece first time
				ownPiece = step
			}

			// Hit enemy attacking piece
			if boardhelper.IsIndexBitSet(step, opponentAttackers) {

				// If we have hit our own piece before, that piece is pinned
				if ownPiece != -1 {
					pinMask |= 1 << ownPiece
					break
				}

				// Fun fact: we are in check if the code comes to this comment
			}

			// Hit any other enemy piece
			if boardhelper.IsIndexBitSet(step, otherOpponentPiecesMask) {
				break // Can just exit, no matter if we hit our own piece first, there is an enemy piece in the way of possible attackers
			}

			step += offset
			depth++
		}
	}

	return pinMask
}
