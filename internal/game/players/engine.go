package players

import (
	"endtner.dev/nChess/internal/board"
	"endtner.dev/nChess/internal/engine"
	"endtner.dev/nChess/internal/game"
	"time"
)

type EnginePlayer struct {
	PlayerType byte
}

func (e EnginePlayer) Init() {
	e.PlayerType = game.Engine
}

func (e EnginePlayer) GetPlayerType() byte {
	return e.PlayerType
}

func (e EnginePlayer) AwaitMove(p *board.Position, legalMoveTable *map[string]board.Move) board.Move {
	return engine.IterativeDeepeningSearch(p, 32, 15*time.Second)
}
