package main

import (
	"endtner.dev/nChess/board"
	"endtner.dev/nChess/formatter"
	"endtner.dev/nChess/movegenerator"
	"fmt"
	"time"
)

var startPosition = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
var testPosition = "8/2p5/3p4/KP5r/1R2Pp1k/8/6P1/8 b - e3 0 1"

func main() {
	b := board.New(startPosition)

	formatter.DisplayPretty(b)
	fmt.Println(b.ToFEN())

	ply := 6
	startTotal := time.Now()
	fmt.Printf("Perft(%d)=%d, took %s\n", ply, movegenerator.Perft(b, ply, ply), time.Since(startTotal))

	return

	fmt.Println("")
	fmt.Printf("Precomputation: %s\n", movegenerator.TotalTimePrecompute)
	fmt.Printf("King Generation: %s\n", movegenerator.TotalTimeKingGeneration)
	fmt.Printf("Pawn Generation: %s\n", movegenerator.TotalTimePawnGeneration)
	fmt.Printf("Sliding Generation: %s\n", movegenerator.TotalTimeSlidingGeneration)
	fmt.Printf("Knight Generation: %s\n", movegenerator.TotalTimeKnightGeneration)
}
