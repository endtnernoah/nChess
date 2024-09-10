package engine

import (
	"endtner.dev/nChess/internal/board"
	"math"
)

func NegaMax(p *board.Position, depth int, alpha, beta float64, tt *TranspositionTable) float64 {
	alpha0 := alpha

	entry, found := tt.Probe(p.Zobrist)
	if found && entry.Depth >= depth {
		if entry.Type == ExactScore {
			return entry.Score
		} else if entry.Type == LowerBound {
			alpha = math.Max(alpha, entry.Score)
		} else if entry.Type == UpperBound {
			beta = math.Min(beta, entry.Score)
		}
		if alpha >= beta {
			return entry.Score
		}
	}

	if depth == 0 || p.IsTerminal {
		return Evaluate(p)
	}

	var bestMove board.Move
	currentEval := math.Inf(-1)
	for _, m := range LegalMoves(p) {
		np := p.MakeMove(m)
		score := -NegaMax(np, depth-1, -beta, -alpha, tt)
		if score > currentEval {
			currentEval = score
			bestMove = m
		}
		alpha = math.Max(alpha, currentEval)
		if alpha >= beta {
			break
		}
	}

	var entryType EntryType
	if currentEval <= alpha0 {
		entryType = UpperBound
	} else if currentEval >= beta {
		entryType = LowerBound
	} else {
		entryType = ExactScore
	}
	tt.Store(p.Zobrist, depth, currentEval, entryType, bestMove)

	return currentEval
}

func Search(p *board.Position, depth int) board.Move {
	tt := NewTranspositionTable()
	bestMove := board.Move{}
	bestValue := math.Inf(-1)
	alpha := math.Inf(-1)
	beta := math.Inf(1)

	for _, m := range LegalMoves(p) {
		np := p.MakeMove(m)
		value := -NegaMax(np, depth-1, -beta, -alpha, tt)
		if value > bestValue {
			bestValue = value
			bestMove = m
		}
		alpha = math.Max(alpha, value)
	}

	return bestMove
}
