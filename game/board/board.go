package board

import (
	"endtner.dev/nChess/game/boardhelper"
	"endtner.dev/nChess/game/piece"
	"fmt"
	"math/bits"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

type Board struct {
	// Bitboards are stored from bottom left to top right, meaning A1 to H8
	// Bitboards are stored at the index of the piece they have
	bitboards []uint64

	WhiteAttackFields uint64
	BlackAttackFields uint64

	whitePinnedPieces uint64
	blackPinnedPieces uint64

	whitePieces uint64
	blackPieces uint64
	occupied    uint64
}

func New(fenString string) *Board {
	// Since fen starts at a8, we want to split it at the /, reverse the list and join it back with /
	fenRows := strings.Split(fenString, "/")
	slices.Reverse(fenRows)
	fenString = strings.Join(fenRows, "/")

	b := Board{}
	b.bitboards = make([]uint64, 0b11111)

	// Setting up pieces
	boardPosition := 0
	for i := 0; i < len(fenString); i++ {
		currentChar := fenString[i]

		if unicode.IsNumber(rune(currentChar)) {
			data, err := strconv.Atoi(string(currentChar))
			if err != nil {
				fmt.Printf("Error parsing character '%q'", currentChar)
				panic(err)
			}

			// Skip n - 1 fields
			boardPosition += data

			continue
		}

		// We set that to a 0 so
		if rune(currentChar) == '/' {
			continue
		}

		// Matching character, setting bits
		b.bitboards[piece.Value(rune(currentChar))] |= 1 << boardPosition

		boardPosition++
	}
	// Setting the occupied fields
	for p, bitboard := range b.bitboards {
		if uint(p)&piece.ColorWhite == piece.ColorWhite {
			b.whitePieces |= bitboard
		}
		if uint(p)&piece.ColorBlack == piece.ColorBlack {
			b.blackPieces |= bitboard
		}
	}

	b.occupied = b.whitePieces | b.blackPieces

	b.UpdateWhiteAttackBitboard()
	b.UpdateBlackAttackBitboard()

	return &b
}

func (b *Board) PieceBitboard(piece uint) uint64 {
	return b.bitboards[piece]
}

func (b *Board) WhitePieces() uint64 {
	return b.whitePieces
}

func (b *Board) BlackPieces() uint64 {
	return b.blackPieces
}

func (b *Board) Occupied() uint64 {
	return b.occupied
}

func (b *Board) UpdateWhiteAttackBitboard() {
	whitePawnAttacks := pawnAttackBitboard(b.PieceBitboard(piece.ColorWhite|piece.TypePawn), b.whitePieces, 8)
	whiteStraightSlidingAttacks := straightSlidingAttackBitboard(
		b.PieceBitboard(piece.ColorWhite|piece.TypeRook)|b.PieceBitboard(piece.ColorWhite|piece.TypeQueen),
		b.whitePieces,
		b.blackPieces,
		8)
	whiteStraightKingAttacks := straightSlidingAttackBitboard(
		b.PieceBitboard(piece.ColorWhite|piece.TypeKing),
		b.whitePieces,
		b.blackPieces,
		1)
	whiteDiagonalSlidingAttacks := diagonalSlidingAttackBitboard(
		b.PieceBitboard(piece.ColorWhite|piece.TypeBishop)|b.PieceBitboard(piece.ColorWhite|piece.TypeQueen),
		b.whitePieces,
		b.blackPieces,
		8)
	whiteDiagonalKingAttacks := diagonalSlidingAttackBitboard(
		b.PieceBitboard(piece.ColorWhite|piece.TypeKing),
		b.whitePieces,
		b.blackPieces,
		1)
	whiteKnightAttacks := knightAttackBitboard(b.PieceBitboard(piece.ColorWhite|piece.TypeKnight), b.whitePieces)

	b.WhiteAttackFields = whitePawnAttacks |
		whiteStraightSlidingAttacks |
		whiteStraightKingAttacks |
		whiteDiagonalSlidingAttacks |
		whiteDiagonalKingAttacks |
		whiteKnightAttacks
}

func (b *Board) UpdateBlackAttackBitboard() {
	blackPawnAttacks := pawnAttackBitboard(b.PieceBitboard(piece.ColorBlack|piece.TypePawn), b.blackPieces, -8)
	blackStraightSlidingAttacks := straightSlidingAttackBitboard(
		b.PieceBitboard(piece.ColorBlack|piece.TypeRook)|b.PieceBitboard(piece.ColorWhite|piece.TypeQueen),
		b.blackPieces,
		b.whitePieces,
		8)
	blackStraightKingAttacks := straightSlidingAttackBitboard(
		b.PieceBitboard(piece.ColorBlack|piece.TypeKing),
		b.blackPieces,
		b.whitePieces,
		1)
	blackDiagonalSlidingAttacks := diagonalSlidingAttackBitboard(
		b.PieceBitboard(piece.ColorBlack|piece.TypeBishop)|b.PieceBitboard(piece.ColorWhite|piece.TypeQueen),
		b.blackPieces,
		b.whitePieces,
		8)
	blackDiagonalKingAttacks := diagonalSlidingAttackBitboard(
		b.PieceBitboard(piece.ColorBlack|piece.TypeKing),
		b.blackPieces,
		b.whitePieces,
		1)
	blackKnightAttacks := knightAttackBitboard(b.PieceBitboard(piece.ColorBlack|piece.TypeKnight), b.blackPieces)

	b.BlackAttackFields = blackPawnAttacks |
		blackStraightSlidingAttacks |
		blackStraightKingAttacks |
		blackDiagonalSlidingAttacks |
		blackDiagonalKingAttacks |
		blackKnightAttacks
}

func pawnAttackBitboard(pieceBitboard uint64, ownPiecesBitboard uint64, pawnIndexOffset int) uint64 {
	var pawnAttacks uint64

	for pieceBitboard != 0 {
		startIndex := bits.TrailingZeros64(pieceBitboard)

		side1 := startIndex + pawnIndexOffset - 1
		side2 := startIndex + pawnIndexOffset + 1

		if (side1 < 0 || side1 > 63) || (side2 < 0 || side2 > 63) {
			continue
		}

		// Add move if it is in the same target row and targets are not empty
		if side1/8 == (startIndex+pawnIndexOffset)/8 && !boardhelper.IsIndexBitSet(side1, ownPiecesBitboard) {
			fmt.Printf("Pawn attack on %s\n", boardhelper.IndexToSquare(side1))
			pawnAttacks |= 1 << side1
		}
		if side2/8 == (startIndex+pawnIndexOffset)/8 && !boardhelper.IsIndexBitSet(side2, ownPiecesBitboard) {
			fmt.Printf("Pawn attack on %s\n", boardhelper.IndexToSquare(side2))
			pawnAttacks |= 1 << side2
		}

		pieceBitboard &= pieceBitboard - 1
	}

	return pawnAttacks
}

func straightSlidingAttackBitboard(pieceBitboard uint64, ownPiecesBitboard uint64, enemyPiecesBitboard uint64, maxLength int) uint64 {
	straightIndexOffsets := []int{1, -1, 8, -8}
	var straightSlidingAttacks uint64

	for pieceBitboard != 0 {
		startIndex := bits.TrailingZeros64(pieceBitboard)

		for _, offset := range straightIndexOffsets {
			targetIndex := startIndex + offset

			length := 0
			// Go deep until we hit borders or our own pieces or maxLength
			for length < maxLength && boardhelper.IsValidStraightMove(startIndex, targetIndex) && !boardhelper.IsIndexBitSet(targetIndex, ownPiecesBitboard) {

				// We can go as deep as a capture
				if boardhelper.IsIndexBitSet(targetIndex, enemyPiecesBitboard) {
					straightSlidingAttacks |= 1 << targetIndex
					break
				}

				straightSlidingAttacks |= 1 << targetIndex
				targetIndex += offset
				length++
			}

		}

		pieceBitboard &= pieceBitboard - 1
	}

	return straightSlidingAttacks
}

func diagonalSlidingAttackBitboard(pieceBitboard uint64, ownPiecesBitboard uint64, enemyPiecesBitboard uint64, maxLength int) uint64 {
	diagonalIndexOffsets := []int{-7, 7, -9, 9}
	var diagonalSlidingAttacks uint64

	for pieceBitboard != 0 {
		startIndex := bits.TrailingZeros64(pieceBitboard)

		for _, offset := range diagonalIndexOffsets {
			targetIndex := startIndex + offset

			length := 0
			// Go deep until we hit borders or our own pieces or maxLength
			for length < maxLength && boardhelper.IsValidDiagonalMove(startIndex, targetIndex) && !boardhelper.IsIndexBitSet(targetIndex, ownPiecesBitboard) {

				// We can go as deep as a capture
				if boardhelper.IsIndexBitSet(targetIndex, enemyPiecesBitboard) {
					diagonalSlidingAttacks |= 1 << targetIndex
					break
				}

				diagonalSlidingAttacks |= 1 << targetIndex
				targetIndex += offset
				length++
			}

		}

		pieceBitboard &= pieceBitboard - 1
	}

	return diagonalSlidingAttacks
}

func knightAttackBitboard(pieceBitboard uint64, ownPiecesBitboard uint64) uint64 {
	knightIndexOffsets := []int{-6, 6, -10, 10, -15, 15, -17, 17}
	var knightAttacks uint64

	for pieceBitboard != 0 {
		startIndex := bits.TrailingZeros64(pieceBitboard)

		for _, offset := range knightIndexOffsets {
			if boardhelper.IsValidKnightMove(startIndex, startIndex+offset) && !boardhelper.IsIndexBitSet(startIndex+offset, ownPiecesBitboard) {
				knightAttacks |= 1 << (startIndex + offset)
			}
		}

		pieceBitboard &= pieceBitboard - 1
	}

	return knightAttacks
}
