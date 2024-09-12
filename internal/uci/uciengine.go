package uci

import (
	"bufio"
	"endtner.dev/nChess/internal/board"
	"fmt"
	"os"
	"strings"
)

type UCIEngine struct {
	currentPos *board.Position
}

func NewUCIEngine() *UCIEngine {
	return &UCIEngine{}
}

func (e *UCIEngine) UCILoop() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command := scanner.Text()
		if err := e.handleCommand(command); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}

func (e *UCIEngine) handleCommand(command string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "uci":
		fmt.Println("id name nChess")
		fmt.Println("id author Noah Endtner")
		fmt.Println("uciok")
	case "isready":
		fmt.Println("readyok")
	case "ucinewgame":
		fmt.Printf("")
	case "position":
		return e.handlePosition(parts[1:])
	case "go":
		return e.handleGo()
	case "quit":
		os.Exit(0)
	default:
		return fmt.Errorf("unknown command: %s", parts[0])
	}

	return nil
}
