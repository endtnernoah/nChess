package engine

import (
	"endtner.dev/nChess/internal/board"
	"fmt"
	"math"
	"time"
)

func IterativeDeepeningSearch(p *board.Position, maxDepth int, timeLimit time.Duration) board.Move {
	tt := NewTranspositionTable()
	killerMoves := make([][2]board.Move, maxDepth+1)

	startTime := time.Now()
	var bestMove board.Move
	bestValue := math.Inf(-1)

	for depth := 1; depth <= maxDepth; depth++ {
		// Time management: check if we have enough time for the next iteration
		elapsedTime := time.Since(startTime)
		if elapsedTime > timeLimit {
			break
		}

		alpha := math.Inf(-1)
		beta := math.Inf(1)

		value := NegaMax(p, depth, alpha, beta, tt, killerMoves)

		// Retrieve the best move from the transposition table
		if entry, found := tt.Probe(p.Zobrist); found && entry.Depth == depth {
			bestMove = entry.Move
		}

		bestValue = value

		// You can add logging here to show progress
		fmt.Printf("Depth %d: bestMove = %s, score = %.2f\n", depth, board.MoveToString(bestMove), bestValue)
	}

	return bestMove
}

func Search(p *board.Position, depth int) board.Move {
	// Initializing transposition table & killer moves
	tt := NewTranspositionTable()
	killerMoves := make([][2]board.Move, depth+1)

	bestMove := board.Move{}
	bestValue := math.Inf(-1)
	alpha := math.Inf(-1)
	beta := math.Inf(1)

	for _, m := range LegalMoves(p) {
		np := p.MakeMove(m)
		value := -NegaMax(np, depth-1, -beta, -alpha, tt, killerMoves)
		if value > bestValue {
			bestValue = value
			bestMove = m
		}
		alpha = math.Max(alpha, value)
	}

	return bestMove
}

func NegaMax(p *board.Position, depth int, alpha, beta float64, tt *TranspositionTable, killerMoves [][2]board.Move) float64 {
	// Original alpha value, used for updating the transposition table
	alpha0 := alpha

	// Transposition table lookup
	ttMove := board.Move{}
	entry, found := tt.Probe(p.Zobrist)
	if found && entry.Depth >= depth {
		ttMove = entry.Move
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

	// Retuning if depth is reached OR the position is terminal
	if depth == 0 || p.IsTerminal {
		return Evaluate(p)
	}

	// Ordering the moves
	orderedMoves := OrderMoves(p, LegalMoves(p), ttMove, killerMoves[depth])

	// Alpha-Beta-Pruned search
	var bestMove board.Move
	currentEval := math.Inf(-1)
	for _, m := range orderedMoves {
		np := p.MakeMove(m)
		score := -NegaMax(np, depth-1, -beta, -alpha, tt, killerMoves)
		if score > currentEval {
			currentEval = score
			bestMove = m
		}
		alpha = math.Max(alpha, currentEval)

		// Pruning the branch
		if alpha >= beta {
			if p.Pieces[m.TargetIndex] == 0 {
				killerMoves[depth][1] = killerMoves[depth][0]
				killerMoves[depth][0] = m
			}
			break
		}
	}

	// Storing in tt
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
