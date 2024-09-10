package main

import (
	"endtner.dev/nChess/internal/game"
	"endtner.dev/nChess/internal/game/players"
)

func main() {
	g := game.NewGame(players.HumanPlayer{}, players.EnginePlayer{})
	g.RunGameLoop()
}
