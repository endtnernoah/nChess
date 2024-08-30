package formatter

import (
	"endtner.dev/nChess/board"
	"endtner.dev/nChess/board/piece"
	"fmt"
	"strings"
)

func BitboardMappingWhite(b *board.Board) map[uint64]string {
	return map[uint64]string{
		b.Bitboards[piece.White|piece.Pawn]:   "♙",
		b.Bitboards[piece.White|piece.Rook]:   "♖",
		b.Bitboards[piece.White|piece.Knight]: "♘",
		b.Bitboards[piece.White|piece.Bishop]: "♗",
		b.Bitboards[piece.White|piece.Queen]:  "♕",
		b.Bitboards[piece.White|piece.King]:   "♔",
	}
}

func BitboardMappingBlack(b *board.Board) map[uint64]string {
	return map[uint64]string{
		b.Bitboards[piece.Black|piece.Pawn]:   "♟",
		b.Bitboards[piece.Black|piece.Rook]:   "♜",
		b.Bitboards[piece.Black|piece.Knight]: "♞",
		b.Bitboards[piece.Black|piece.Bishop]: "♝",
		b.Bitboards[piece.Black|piece.Queen]:  "♛",
		b.Bitboards[piece.Black|piece.King]:   "♚",
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
	for bitboard, p := range bitboardMapping {
		for i := 0; i < 64; i++ {
			if bitboard&(1<<i) != 0 {
				unicodeBoard[i] = p
			}
		}
	}
	return unicodeBoard
}

func UnicodeBoard(board []string) string {
	/*
		Please just do not use this function anywhere
	*/
	var result strings.Builder

	for i := 63; i >= 0; i -= 8 {
		for j := 7; j >= 0; j-- {
			result.WriteString(board[i-j])
			result.WriteString(" ")
		}
		result.WriteString("\n")
	}
	return result.String()
}

func UnicodeBoardWithBorders(board []string) string {
	/*
		claude.ai is responsible for this satanic child of a function, but it does work like a charm
	*/
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
			index := (rank-1)*8 + file
			result.WriteString(fmt.Sprintf(" %s ", board[index]))
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

func Display(b *board.Board) {
	unicodeBoard := ToUnicodeBoard(BitboardMappingAll(b))
	fmt.Println(UnicodeBoard(unicodeBoard))
}

func DisplayPretty(b *board.Board) {
	unicodeBoard := ToUnicodeBoard(BitboardMappingAll(b))
	fmt.Println(UnicodeBoardWithBorders(unicodeBoard))
}
