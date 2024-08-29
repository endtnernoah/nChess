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

	startTotal := time.Now()
	g.Perft(6)
	fmt.Printf("Perft(6) took %s\n", time.Since(startTotal))
}
