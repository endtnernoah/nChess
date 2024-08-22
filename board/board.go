package board

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type Board struct {
	whiteToMove           bool
	castlingAvailability  uint
	enPassantTargetSquare uint64
	halfMoves             uint
	moveCount             uint

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

	occupied      uint64
	occupiedWhite uint64
	occupiedBlack uint64
}

func New(fen string) *Board {
	b := Board{}

	fenFields := strings.Split(fen, " ")

	// Setting up pieces
	boardPosition := 0
	for i := 0; i < len(fenFields[0]); i++ {
		currentChar := fenFields[0][i]

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
	b.occupiedWhite = b.whiteRooks | b.whiteKnights | b.whiteBishops | b.whiteQueens | b.whiteKing | b.whitePawns
	b.occupiedBlack = b.blackRooks | b.blackKnights | b.blackBishops | b.blackQueens | b.blackKing | b.blackPawns
	b.occupied = b.occupiedWhite | b.occupiedBlack

	// Checking who is to move
	b.whiteToMove = fenFields[1] == "w"

	// Castling availability
	castlingAvailabilityFlags := fenFields[2]
	if strings.Contains(castlingAvailabilityFlags, "K") {
		b.castlingAvailability |= 0b1000
	}
	if strings.Contains(castlingAvailabilityFlags, "Q") {
		b.castlingAvailability |= 0b0100
	}
	if strings.Contains(castlingAvailabilityFlags, "k") {
		b.castlingAvailability |= 0b0010
	}
	if strings.Contains(castlingAvailabilityFlags, "q") {
		b.castlingAvailability |= 0b0001
	}

	// EP Target Square
	if fenFields[3] != "-" {
		b.enPassantTargetSquare = 1 << SquareToIndex(fenFields[3])
	}

	// Move numbers
	data, err := strconv.ParseUint(fenFields[4], 10, 64)
	if err != nil {
		fmt.Println("Failed parsing halfMove number")
		panic(err)
	}
	b.halfMoves = uint(data)

	data, err = strconv.ParseUint(fenFields[5], 10, 64)
	if err != nil {
		fmt.Println("Failed parsing moveCount number")
		panic(err)
	}
	b.moveCount = uint(data)

	return &b
}

func (b *Board) ToUnicode() {
	board := make([]string, 64)
	for i := range board {
		board[i] = " " // Initialize with empty squares
	}

	// Function to set pieces on the board
	setPieces := func(bitboard uint64, piece string) {
		for i := 0; i < 64; i++ {
			if bitboard&(1<<i) != 0 {
				board[i] = piece
			}
		}
	}

	// Set pieces for each bitboard
	setPieces(b.whitePawns, "♙")
	setPieces(b.whiteRooks, "♖")
	setPieces(b.whiteKnights, "♘")
	setPieces(b.whiteBishops, "♗")
	setPieces(b.whiteQueens, "♕")
	setPieces(b.whiteKing, "♔")
	setPieces(b.blackPawns, "♟")
	setPieces(b.blackRooks, "♜")
	setPieces(b.blackKnights, "♞")
	setPieces(b.blackBishops, "♝")
	setPieces(b.blackQueens, "♛")
	setPieces(b.blackKing, "♚")

	fmt.Println(formatUnicodeBoardWithBorders(board))
}

func (b *Board) PrintOccupied() {
	fmt.Println(formatString(bitboardToString(b.occupied)))
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

func bitboardToString(bitboard uint64) string {
	return reverseString(fmt.Sprintf("%064b", bitboard))
}

func formatUnicodeBoard(board []string) string {
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

func formatUnicodeBoardWithBorders(board []string) string {
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
		result.WriteString(fmt.Sprintf("  %c ", file))
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
		result.WriteString(fmt.Sprintf("  %c ", file))
	}
	result.WriteString("\n")

	return result.String()
}

func formatString(s string) string {
	fmt.Println(len(s))

	// Build string
	var result strings.Builder
	for i, char := range s {
		if i > 0 && i%8 == 0 {
			result.WriteRune('\n')
		}
		result.WriteRune(char)
		result.WriteRune(' ')
	}

	return result.String()
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]

	}
	return string(runes)
}
