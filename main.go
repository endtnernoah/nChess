package main

import (
	"endtner.dev/nChess/internal/engine"
	"endtner.dev/nChess/internal/utils"
	"fmt"
	"time"
)

var startPosition = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

func main() {
	p := utils.FromFen(startPosition)

	utils.Display(p)
	fmt.Println(utils.ToFEN(p))

	ply := 6
	startTotal := time.Now()
	fmt.Printf("Perft(%d)=%d, took %s\n", ply, engine.Perft(p, ply, ply), time.Since(startTotal))

	fmt.Println("")
	fmt.Printf("Precomputation: %s\n", engine.TotalTimePrecompute)
	fmt.Printf("King Generation: %s\n", engine.TotalTimeKingGeneration)
	fmt.Printf("Pawn Generation: %s\n", engine.TotalTimePawnGeneration)
	fmt.Printf("Sliding Generation: %s\n", engine.TotalTimeSlidingGeneration)
	fmt.Printf("Knight Generation: %s\n", engine.TotalTimeKnightGeneration)
}
