package main

import (
	"endtner.dev/nChess/board"
	"endtner.dev/nChess/formatter"
	"endtner.dev/nChess/movegenerator"
	"fmt"
	"time"
)

var startPosition = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

/*
	1. Refactor all the board.go into a single old.go
	2. Move the generation completely to the move generator.
	3. Do the pinned ray checks in the generator function
	4. Better move generation order
*/

func main() {
	b := board.New(startPosition)

	formatter.DisplayPretty(b)
	fmt.Println(b.ToFEN())

	ply := 6
	startTotal := time.Now()
	fmt.Printf("Perft(%d)=%d, took %s\n", ply, movegenerator.Perft(b, ply), time.Since(startTotal))
}
