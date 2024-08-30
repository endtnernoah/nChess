package movegenerator

import (
	"endtner.dev/nChess/board"
	"endtner.dev/nChess/board/boardhelper"
	"endtner.dev/nChess/board/piece"
	"math/bits"
)

var Pins = make([]uint64, 2)

func ComputePins(b *board.Board, color uint8) {
	straightPinnedPieces := StraightPins(b, color)
	diagonalPinnedPieces := DiagonalPins(b, color)

	Pins[(color>>3)-1] = straightPinnedPieces | diagonalPinnedPieces
}

/*
	Generators
*/

func StraightPins(b *board.Board, friendlyColor uint8) uint64 {
	var pinMask uint64

	friendlyKingIndex := bits.TrailingZeros64(b.Bitboards[friendlyColor|piece.King])

	friendlyPiecesMask := Occupancy[(friendlyColor>>3)-1]
	opponentPiecesMask := Occupancy[1-((friendlyColor>>3)-1)]

	opponentColor := piece.White
	if friendlyColor == piece.White {
		opponentColor = piece.Black
	}

	opponentAttackers := b.Bitboards[opponentColor|piece.Rook] | b.Bitboards[opponentColor|piece.Queen]
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

func DiagonalPins(b *board.Board, friendlyColor uint8) uint64 {
	var pinMask uint64

	friendlyKingIndex := bits.TrailingZeros64(b.Bitboards[friendlyColor|piece.King])
	friendlyPiecesMask := Occupancy[(friendlyColor>>3)-1]
	opponentPiecesMask := Occupancy[1-((friendlyColor>>3)-1)]

	opponentColor := piece.White
	if friendlyColor == piece.White {
		opponentColor = piece.Black
	}

	opponentAttackers := b.Bitboards[opponentColor|piece.Bishop] | b.Bitboards[opponentColor|piece.Queen]
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
