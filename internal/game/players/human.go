package players

import (
	"bufio"
	"endtner.dev/nChess/internal/board"
	"endtner.dev/nChess/internal/game"
	"fmt"
	"os"
	"strings"
)

type HumanPlayer struct {
	PlayerType byte
}

func (h HumanPlayer) Init() {
	h.PlayerType = game.Human
}

func (h HumanPlayer) GetPlayerType() byte {
	return h.PlayerType
}

func (h HumanPlayer) AwaitMove(p *board.Position, legalMoveTable *map[string]board.Move) board.Move {
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input:", err)
		return h.AwaitMove(p, legalMoveTable)
	}

	moveString := strings.TrimSpace(text)

	if entry, found := (*legalMoveTable)[moveString]; found {
		return entry
	} else {
		fmt.Printf("'%s'\n", moveString)
		fmt.Println(*legalMoveTable)
		fmt.Print("Illegal move. Enter a new move: ")
		return h.AwaitMove(p, legalMoveTable)
	}
}
