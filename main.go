package main

import (
	"endtner.dev/nChess/game"
	"fmt"
	"time"
)

var defaultFen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
var perftPosition6 = "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10 "

func main() {
	g := game.New(perftPosition6)

	g.DisplayBoardPretty()
	fmt.Println(g.ToFEN())

	startTotal := time.Now()
	for _, m := range g.GenerateLegalMoves() {
		g.MakeMove(m)
		g.UnmakeMove()
	}
	fmt.Printf("Generating legal moves took %s\n", time.Since(startTotal))
}
