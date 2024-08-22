package formatter

import (
	"endtner.dev/nChess/game/board"
	"fmt"
	"strings"
)

func BitboardMappingWhite(b *board.Board) map[uint64]string {
	return map[uint64]string{
		b.WhitePawns():   "♙",
		b.WhiteRooks():   "♖",
		b.WhiteKnights(): "♘",
		b.WhiteBishops(): "♗",
		b.WhiteQueens():  "♕",
		b.WhiteKing():    "♔",
	}
}

func BitboardMappingBlack(b *board.Board) map[uint64]string {
	return map[uint64]string{
		b.BlackPawns():   "♟",
		b.BlackRooks():   "♜",
		b.BlackKnights(): "♞",
		b.BlackBishops(): "♝",
		b.BlackQueens():  "♛",
		b.BlackKing():    "♚",
	}
}

func BitboardMappingAll(b *board.Board) map[uint64]string {
	white := BitboardMappingWhite(b)
	black := BitboardMappingBlack(b)

	for k, v := range black {
		white[k] = v
	}
	return white
}

func ToUnicodeBoard(bitboardMapping map[uint64]string) []string {
	unicodeBoard := make([]string, 64)
	for i := range unicodeBoard {
		unicodeBoard[i] = " " // Initialize with empty squares
	}

	// Setting pieces
	for bitboard, piece := range bitboardMapping {
		for i := 0; i < 64; i++ {
			if bitboard&(1<<i) != 0 {
				unicodeBoard[i] = piece
			}
		}
	}
	return unicodeBoard
}

func FormatUnicodeBoard(board []string) string {
	var result strings.Builder

	for i := 63; i >= 0; i -= 8 {
		for j := 0; j < 8; j++ {
			result.WriteString(board[i-j])
			result.WriteString(" ")
		}
		result.WriteString("\n")
	}
	return result.String()
}

func FormatUnicodeBoardWithBorders(board []string) string {
	var result strings.Builder

	// Unicode box-drawing characters
	const (
		topLeft     = "┌"
		topRight    = "┐"
		bottomLeft  = "└"
		bottomRight = "┘"
		horizontal  = "─"
		vertical    = "│"
		cross       = "┼"
		topT        = "┬"
		bottomT     = "┴"
		leftT       = "├"
		rightT      = "┤"
	)

	// Write top border with file letters
	result.WriteString("  ")
	for file := 'a'; file <= 'h'; file++ {
		result.WriteString(fmt.Sprintf(" %c  ", file))
	}
	result.WriteString("\n")

	result.WriteString(" " + topLeft + horizontal + strings.Repeat(horizontal+horizontal+topT+horizontal, 7) + horizontal + horizontal + topRight + "\n")

	for rank := 8; rank >= 1; rank-- {
		result.WriteString(fmt.Sprintf("%d", rank)) // Rank number
		result.WriteString(vertical)
		for file := 0; file < 8; file++ {
			index := (rank-1)*8 + (7 - file)
			result.WriteString(fmt.Sprintf(" %s ", board[63-index]))
			result.WriteString(vertical)
		}
		result.WriteString(fmt.Sprintf("%d\n", rank)) // Rank number

		if rank > 1 {
			result.WriteString(" " + leftT + horizontal + strings.Repeat(horizontal+horizontal+cross+horizontal, 7) + horizontal + horizontal + rightT + "\n")
		}
	}

	result.WriteString(" " + bottomLeft + horizontal + strings.Repeat(horizontal+horizontal+bottomT+horizontal, 7) + horizontal + horizontal + bottomRight + "\n")

	// Write bottom border with file letters
	result.WriteString("  ")
	for file := 'a'; file <= 'h'; file++ {
		result.WriteString(fmt.Sprintf(" %c  ", file))
	}
	result.WriteString("\n")

	return result.String()
}
