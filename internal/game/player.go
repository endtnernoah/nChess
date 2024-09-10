package game

import "endtner.dev/nChess/internal/board"

/*
	Will hold an interface that can be used for players. This will allow for multiple match types to happen
	e.g. Player vs Player, Player vs Engine, Engine vs Engine
*/

const (
	Human byte = iota
	Engine
)

type AbstractPlayer interface {
	AwaitMove(p *board.Position, legalMoveTable *map[string]board.Move) board.Move
	Init()
	GetPlayerType() byte
}
