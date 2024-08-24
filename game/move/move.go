package move

import (
	"endtner.dev/nChess/game/boardhelper"
	"fmt"
)

type Move struct {
	StartIndex             int
	TargetIndex            int
	EnPassantCaptureSquare int
	EnPassantPassedSquare  int
	RookStartingSquare     int
	IsPromotion            bool
}

func New(startIndex int, targetIndex int, enPassantCaptureSquare int, enPassantPassedSquare int, rookStartingSquare int, isPromotion bool) Move {
	return Move{
		StartIndex:             startIndex,
		TargetIndex:            targetIndex,
		EnPassantCaptureSquare: enPassantCaptureSquare,
		EnPassantPassedSquare:  enPassantPassedSquare,
		RookStartingSquare:     rookStartingSquare,
		IsPromotion:            isPromotion,
	}
}

func Print(m Move) {
	fmt.Printf("Move(from: %s, to: %s)\n", boardhelper.IndexToSquare(m.StartIndex), boardhelper.IndexToSquare(m.TargetIndex))
}
