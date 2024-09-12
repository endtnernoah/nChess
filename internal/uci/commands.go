package uci

import (
	"endtner.dev/nChess/internal/board"
	"endtner.dev/nChess/internal/engine"
	"endtner.dev/nChess/internal/utils"
	"fmt"
	"strings"
	"time"
)

/*
	This will handle the commands for the UCI Engine. Possibly will be refactored
*/

func (e *UCIEngine) handleGo() error {
	fmt.Printf("bestmove %s\n", board.MoveToString(engine.IterativeDeepeningSearch(e.currentPos, 32, 10*time.Second)))
	return nil
}

func (e *UCIEngine) handlePosition(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("invalid position command")
	}

	if args[0] == "startpos" {
		args = args[1:]
		e.currentPos = utils.FromFen(utils.StartPosition)
	} else if args[0] == "fen" {
		if len(args) < 6 {
			return fmt.Errorf("invalid FEN")
		}
		fen := strings.Join(args[1:7], " ")
		e.currentPos = utils.FromFen(fen)
		args = args[7:]
	} else {
		return fmt.Errorf("invalid position command")
	}

	if len(args) > 0 && args[0] == "moves" {
		for _, moveStr := range args[1:] {
			legalMoves := engine.LegalMoves(e.currentPos)

			moveFound := false
			for _, m := range legalMoves {
				if board.MoveToString(m) == moveStr {
					moveFound = true
					e.currentPos = e.currentPos.MakeMove(m)
					break
				}
			}
			if !moveFound {
				return fmt.Errorf("move %s not possible", moveStr)
			}
		}
	}

	fmt.Printf("info current position: %s\n", utils.ToFEN(e.currentPos))

	return nil
}
