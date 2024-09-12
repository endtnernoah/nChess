package engine

import (
	"endtner.dev/nChess/internal/board"
	"math"
	"time"
)

var (
	historyTable [64][64]int
	counterMoves [64][64]board.Move
)

func IterativeDeepeningSearch(p *board.Position, maxDepth int, timeLimit time.Duration) board.Move {
	tt := NewTranspositionTable()
	killerMoves := make([][2]board.Move, maxDepth+1)

	pv := make([][]board.Move, maxDepth+1)
	for i := range pv {
		pv[i] = make([]board.Move, maxDepth+1)
	}

	startTime := time.Now()
	var bestMove board.Move

	for depth := 1; depth <= maxDepth; depth++ {
		// Time management: check if we have enough time for the next iteration
		elapsedTime := time.Since(startTime)
		if elapsedTime > timeLimit {
			break
		}

		alpha := math.Inf(-1)
		beta := math.Inf(1)

		NegaMax(p, depth, alpha, beta, tt, killerMoves, pv)

		// Retrieve the best move from the transposition table
		if entry, found := tt.Probe(p.Zobrist); found && entry.Depth == depth {
			bestMove = entry.Move
		}

		// Time management: check if we have enough time for the next iteration
		elapsedTime = time.Since(startTime)
		if elapsedTime > timeLimit/2 {
			break
		}
	}

	return bestMove
}

func NegaMax(p *board.Position, depth int, alpha, beta float64, tt *TranspositionTable, killerMoves [][2]board.Move, pv [][]board.Move) float64 {
	// Original alpha value, used for updating the transposition table
	alpha0 := alpha

	// Transposition table lookup
	ttMove, shouldReturn, ttScore := tt.Query(p.Zobrist, depth, alpha, beta)
	if shouldReturn {
		return ttScore
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

		// min(a, b) = -max(-b, -a)
		score := -NegaMax(np, depth-1, -beta, -alpha, tt, killerMoves, pv)

		if score > currentEval {
			currentEval = score
			bestMove = m
		}
		alpha = math.Max(alpha, currentEval)

		// Move is too good (killer move), opponent will play another move
		if alpha >= beta {
			if p.Pieces[m.TargetIndex] == 0 {
				killerMoves[depth][1] = killerMoves[depth][0]
				killerMoves[depth][0] = m
			}
			break
		}
	}

	// Storing in TT
	tt.Store(p.Zobrist, depth, currentEval, alpha0, beta, bestMove)

	return alpha
}
