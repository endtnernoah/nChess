package main

import (
	"endtner.dev/nChess/game"
	"endtner.dev/nChess/game/board"
	"fmt"
)

var defaultFen string = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
var testFen string = "8/8/8/4p1K1/2k1P3/8/8/8 b - - 0 1"

func main() {
	g := game.New(testFen)
	g.DisplayBoard()
	fmt.Println(board.IndexToSquare(63))
}
