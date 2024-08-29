package main

import (
	"endtner.dev/nChess/game"
	"fmt"
	"time"
)

var startPosition = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

func main() {
	g := game.New(startPosition)

	g.DisplayBoardPretty()
	fmt.Println(g.ToFEN())

	ply := 6
	startTotal := time.Now()
	fmt.Printf("Perft(%d)=%d, took %s\n", ply, g.Perft(ply), time.Since(startTotal))
}
