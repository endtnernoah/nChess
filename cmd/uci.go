package main

import "endtner.dev/nChess/internal/uci"

func main() {
	engine := uci.NewUCIEngine()
	engine.UCILoop()
}
