package utils

import (
	"endtner.dev/nChess/internal/board"
	"fmt"
	"strings"
)

func BitboardMappingWhite(p *board.Position) map[uint64]string {
	return map[uint64]string{
		p.Bitboards[board.White|board.Pawn]:   "♙",
		p.Bitboards[board.White|board.Rook]:   "♖",
		p.Bitboards[board.White|board.Knight]: "♘",
		p.Bitboards[board.White|board.Bishop]: "♗",
		p.Bitboards[board.White|board.Queen]:  "♕",
		p.Bitboards[board.White|board.King]:   "♔",
	}
}

func BitboardMappingBlack(p *board.Position) map[uint64]string {
	return map[uint64]string{
		p.Bitboards[board.Black|board.Pawn]:   "♟",
		p.Bitboards[board.Black|board.Rook]:   "♜",
		p.Bitboards[board.Black|board.Knight]: "♞",
		p.Bitboards[board.Black|board.Bishop]: "♝",
		p.Bitboards[board.Black|board.Queen]:  "♛",
		p.Bitboards[board.Black|board.King]:   "♚",
	}
}

func BitboardMappingAll(p *board.Position) map[uint64]string {
	white := BitboardMappingWhite(p)
	black := BitboardMappingBlack(p)

	for k, v := range black {
		white[k] = v
	}
	return white
}

func FromMapping(bitboardMapping map[uint64]string) []string {
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

func ToString(board []string) string {
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

func Display(p *board.Position) {
	unicodeBoard := FromMapping(BitboardMappingAll(p))
	fmt.Println(ToString(unicodeBoard))
}
