package main

import (
	"endtner.dev/nChess/game"
	"fmt"
)

var defaultFen string = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
var testFen string = "8/8/8/4p1K1/2k1P3/8/8/8 b - - 0 1"

func main() {
	g := game.New(defaultFen)
	fmt.Println("Game is set up")
	g.DisplayBoardPretty()

	/*
		b := g.Board()
		fmt.Println("-------")
		b.MakeMove(move.Move{8, 16, -1, -1, false})
		b.MakeMove(move.Move{48, 32, -1, 40, false})
		b.MakeMove(move.Move{1, 18, -1, -1, false})
		g.DisplayBoardPretty()
		fmt.Println("-------")
		b.UnmakeMove()
		g.DisplayBoardPretty()

	*/

	for i := range 10 {
		fmt.Printf("Perft(%d) = %d\n", i, g.Perft(i))
	}

	//fmt.Println(len(possibleMoves))
}
