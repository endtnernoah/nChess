package main

import (
	"endtner.dev/nChess/board"
	"endtner.dev/nChess/board/move"
	"endtner.dev/nChess/formatter"
	"endtner.dev/nChess/movegenerator"
	"fmt"
	"time"
)

var startPosition = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
var testPosition = "8/2p5/3p4/KP5r/1R2Pp1k/8/6P1/8 b - e3 0 1"

func main() {
	b := board.New("8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1")

	b.MakeMove(move.New(25, 1))
	b.MakeMove(move.New(29, 21))
	b.MakeMove(move.New(32, 25))
	b.MakeMove(move.New(50, 34, move.WithEnPassantPassedSquare(42)))

	formatter.DisplayPretty(b)
	fmt.Println(b.ToFEN())

	ply := 2
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
