package main

import (
	"endtner.dev/nChess/board"
	"endtner.dev/nChess/evaluator"
	"endtner.dev/nChess/formatter"
	"endtner.dev/nChess/movegenerator"
	"fmt"
	"time"
)

var startPosition = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
var testPosition = "2B5/2P2n2/1KR5/p6k/1P1P1P2/2N3pp/1P2p3/n7 w - - 0 1"

func main() {
	b := board.New(testPosition)

	formatter.DisplayPretty(b)
	fmt.Println(b.ToFEN())
	fmt.Printf("score: %d\n", evaluator.Evaluate(b))

	return

	ply := 6
	startTotal := time.Now()
	fmt.Printf("Perft(%d)=%d, took %s\n", ply, movegenerator.Perft(b, ply, -1), time.Since(startTotal))

	fmt.Println("")
	fmt.Printf("Precomputation: %s\n", movegenerator.TotalTimePrecompute)
	fmt.Printf("King Generation: %s\n", movegenerator.TotalTimeKingGeneration)
	fmt.Printf("Pawn Generation: %s\n", movegenerator.TotalTimePawnGeneration)
	fmt.Printf("Sliding Generation: %s\n", movegenerator.TotalTimeSlidingGeneration)
	fmt.Printf("Knight Generation: %s\n", movegenerator.TotalTimeKnightGeneration)
}
