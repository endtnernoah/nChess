package board

import (
	"fmt"
	"strconv"
	"unicode"
)

type Board struct {
	whitePawns   uint64
	whiteRooks   uint64
	whiteKnights uint64
	whiteBishops uint64
	whiteQueens  uint64
	whiteKing    uint64

	blackPawns   uint64
	blackRooks   uint64
	blackKnights uint64
	blackBishops uint64
	blackQueens  uint64
	blackKing    uint64

	whitePieces uint64
	blackPieces uint64
	occupied    uint64
}

func New(fenString string) *Board {
	b := Board{}

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
		switch rune(currentChar) {
		case 'r':
			b.blackRooks |= 1 << boardPosition
		case 'n':
			b.blackKnights |= 1 << boardPosition
		case 'b':
			b.blackBishops |= 1 << boardPosition
		case 'q':
			b.blackQueens |= 1 << boardPosition
		case 'k':
			b.blackKing |= 1 << boardPosition
		case 'p':
			b.blackPawns |= 1 << boardPosition

		case 'R':
			b.whiteRooks |= 1 << boardPosition
		case 'N':
			b.whiteKnights |= 1 << boardPosition
		case 'B':
			b.whiteBishops |= 1 << boardPosition
		case 'Q':
			b.whiteQueens |= 1 << boardPosition
		case 'K':
			b.whiteKing |= 1 << boardPosition
		case 'P':
			b.whitePawns |= 1 << boardPosition
		default:
			panic("Invalid char in fen string")
		}

		boardPosition++
	}
	// Setting the occupied fields
	b.whitePieces = b.whiteRooks | b.whiteKnights | b.whiteBishops | b.whiteQueens | b.whiteKing | b.whitePawns
	b.blackPieces = b.blackRooks | b.blackKnights | b.blackBishops | b.blackQueens | b.blackKing | b.blackPawns
	b.occupied = b.whitePieces | b.blackPieces

	return &b
}

func SquareToIndex(square string) uint {
	if len(square) != 2 {
		return 0
	}

	file := int(unicode.ToLower(rune(square[0]))) - int('a')
	rank := 7 - (int(square[1]) - int('1'))

	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return 0
	}

	return uint(rank*8 + file)
}

func IndexToSquare(index uint) string {
	if index < 0 || index > 63 {
		return ""
	}

	file := index % 8
	rank := 7 - (index / 8)

	return string(rune('a'+file)) + string(rune('1'+rank))
}

func (b *Board) WhitePawns() uint64 {
	return b.whitePawns
}

func (b *Board) WhiteRooks() uint64 {
	return b.whiteRooks
}

func (b *Board) WhiteKnights() uint64 {
	return b.whiteKnights
}

func (b *Board) WhiteBishops() uint64 {
	return b.whiteBishops
}

func (b *Board) WhiteQueens() uint64 {
	return b.whiteQueens
}

func (b *Board) WhiteKing() uint64 {
	return b.whiteKing
}

func (b *Board) BlackPawns() uint64 {
	return b.blackPawns
}

func (b *Board) BlackRooks() uint64 {
	return b.blackRooks
}

func (b *Board) BlackKnights() uint64 {
	return b.blackKnights
}

func (b *Board) BlackBishops() uint64 {
	return b.blackBishops
}

func (b *Board) BlackQueens() uint64 {
	return b.blackQueens
}

func (b *Board) BlackKing() uint64 {
	return b.blackKing
}

func (b *Board) OccupiedWhite() uint64 {
	return b.whitePieces
}

func (b *Board) OccupiedBlack() uint64 {
	return b.blackPieces
}

func (b *Board) Occupied() uint64 {
	return b.occupied
}
