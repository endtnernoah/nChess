package board

import (
	"fmt"
)

type Move struct {
	StartIndex             int
	TargetIndex            int
	EnPassantCaptureSquare int
	EnPassantPassedSquare  int
	RookStartingSquare     int
	PromotionPiece         uint8
}

type OptionalParameter func(*Move)

func WithEnPassantCaptureSquare(square int) OptionalParameter {
	return func(m *Move) {
		m.EnPassantCaptureSquare = square
	}
}

func WithEnPassantPassedSquare(square int) OptionalParameter {
	return func(m *Move) {
		m.EnPassantPassedSquare = square
	}
}

func WithRookStartingSquare(square int) OptionalParameter {
	return func(m *Move) {
		m.RookStartingSquare = square
	}
}

func WithPromotion(promotionPiece uint8) OptionalParameter {
	return func(m *Move) {
		m.PromotionPiece = promotionPiece
	}
}

func New(startIndex int, targetIndex int, optionalParameters ...OptionalParameter) Move {
	m := Move{
		StartIndex:             startIndex,
		TargetIndex:            targetIndex,
		EnPassantCaptureSquare: -1,
		EnPassantPassedSquare:  -1,
		RookStartingSquare:     -1,
		PromotionPiece:         0,
	}

	for _, optionalParameter := range optionalParameters {
		optionalParameter(&m)
	}

	return m
}

func Print(m Move) string {
	return fmt.Sprintf("Move(from: %s, to: %s)\n", IndexToSquare(m.StartIndex), IndexToSquare(m.TargetIndex))
}

func PrintSimple(m Move) string {
	return fmt.Sprintf("%s%s%s", IndexToSquare(m.StartIndex), IndexToSquare(m.TargetIndex), ToString(m.PromotionPiece))
}
