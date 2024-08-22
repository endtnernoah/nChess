package game

import (
	"endtner.dev/nChess/game/board"
	"endtner.dev/nChess/game/formatter"
	"fmt"
	"strconv"
	"strings"
)

type Game struct {
	b *board.Board

	// From fen
	whiteToMove           bool
	castlingAvailability  uint
	enPassantTargetSquare uint64
	halfMoves             uint
	moveCount             uint
}

func New(fenString string) *Game {
	g := Game{}

	// Setting up game from fen
	fenFields := strings.Split(fenString, " ")

	// Setting up board
	g.b = board.New(fenFields[0])

	// Checking who is to move
	g.whiteToMove = fenFields[1] == "w"

	// Castling availability
	castlingAvailabilityFlags := fenFields[2]
	if strings.Contains(castlingAvailabilityFlags, "K") {
		g.castlingAvailability |= 0b1000
	}
	if strings.Contains(castlingAvailabilityFlags, "Q") {
		g.castlingAvailability |= 0b0100
	}
	if strings.Contains(castlingAvailabilityFlags, "k") {
		g.castlingAvailability |= 0b0010
	}
	if strings.Contains(castlingAvailabilityFlags, "q") {
		g.castlingAvailability |= 0b0001
	}

	// EP Target Square
	if fenFields[3] != "-" {
		g.enPassantTargetSquare = 1 << board.SquareToIndex(fenFields[3])
	}

	// Half move count
	data, err := strconv.ParseUint(fenFields[4], 10, 64)
	if err != nil {
		fmt.Println("Failed parsing halfMove number")
		panic(err)
	}
	g.halfMoves = uint(data)

	// Move count
	data, err = strconv.ParseUint(fenFields[5], 10, 64)
	if err != nil {
		fmt.Println("Failed parsing moveCount number")
		panic(err)
	}
	g.moveCount = uint(data)

	return &g
}

func (g *Game) DisplayBoard() {
	unicodeBoard := formatter.ToUnicodeBoard(formatter.BitboardMappingAll(g.b))
	fmt.Println(formatter.FormatUnicodeBoardWithBorders(unicodeBoard))
}
