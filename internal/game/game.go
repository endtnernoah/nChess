package game

import (
	"endtner.dev/nChess/internal/board"
	"endtner.dev/nChess/internal/engine"
	"endtner.dev/nChess/internal/utils"
	"fmt"
)

/*
	This will hold logic for:
	- Turn based move making
	- Holding all rules for a game (50 move rule, 75 move rule...)
	- Checking the state for checkmate...
	- Allows for 2 player interface objects that will hold something like a getNextMove(*position, *legalMoves????)
	- Somehow needs to validate if a move is valid?
	- Maybe the game will validate that
	- AI Interface needs to hold the engine for search
*/

type Game struct {
	playerWhite     AbstractPlayer
	playerBlack     AbstractPlayer
	currentPosition *board.Position
}

func NewGame(playerWhite AbstractPlayer, playerBlack AbstractPlayer) *Game {
	return &Game{playerWhite: playerWhite, playerBlack: playerBlack, currentPosition: utils.FromFen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")}
}

func (g *Game) RunGameLoop() {
	g.playerWhite.Init()
	g.playerBlack.Init()

	for !g.currentPosition.IsTerminal {
		utils.Display(g.currentPosition)

		playerToMove := g.playerBlack
		colorToMove := "Black"
		if g.currentPosition.WhiteToMove {
			playerToMove = g.playerWhite
			colorToMove = "White"
		}

		legalMoves := engine.LegalMoves(g.currentPosition)
		legalMoveTable := make(map[string]board.Move)

		for _, m := range legalMoves {
			legalMoveTable[board.MoveToString(m)] = m
		}
		if playerToMove.GetPlayerType() == Human {
			fmt.Printf("[%s] Enter move: ", colorToMove)
		} else {
			fmt.Printf("[%s] Thinking...\n", colorToMove)
		}
		playedMove := playerToMove.AwaitMove(g.currentPosition, &legalMoveTable)
		g.currentPosition = g.currentPosition.MakeMove(playedMove)
	}
	fmt.Println()
	fmt.Printf("Game terminated by: %s\n", g.currentPosition.TerminalReason)
}
