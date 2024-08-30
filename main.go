package main

import (
	"endtner.dev/nChess/board"
	"endtner.dev/nChess/formatter"
	"endtner.dev/nChess/movegenerator"
	"fmt"
	"time"
)

var startPosition = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

func main() {
	b := board.New(startPosition)

	formatter.DisplayPretty(b)
	movegenerator.ComputeAll(b)
	fmt.Println(b.ToFEN())

	ply := 6
	startTotal := time.Now()
	fmt.Printf("Perft(%d)=%d, took %s\n", ply, movegenerator.Perft(b, ply), time.Since(startTotal))
}
