package main

import (
	"endtner.dev/nChess/board"
)

var defaultFen string = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
var testFen string = "8/8/8/2k5/4K3/8/8/8 w - - 0 1"

func main() {
	b := board.New(testFen)
	b.ToUnicode()
}
