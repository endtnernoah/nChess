package engine

import "endtner.dev/nChess/internal/board"

import (
	"sort"
)

const (
	HashMove   = 10000 // Highest priority for moves from the transposition table
	Capture    = 100   // Priority for capture moves
	Promotion  = 90    // Priority for pawn promotions
	KillerMove = 80    // Priority for killer moves
)

type ScoredMove struct {
	Move  board.Move
	Score int
}

func OrderMoves(p *board.Position, moves []board.Move, ttMove board.Move, killerMoves [2]board.Move) []board.Move {
	scoredMoves := make([]ScoredMove, len(moves))

	for i, move := range moves {
		score := 0

		// Hash move (from transposition table)
		if move == ttMove {
			score += HashMove
		}

		// Captures
		if p.Pieces[move.TargetIndex] != 0 {
			score += Capture
			// MVV-LVA (Most Valuable Victim - Least Valuable Attacker)
			score += MVV_LVA(p, move)
		}

		// Promotions
		if move.PromotionPiece != 0 {
			score += Promotion
		}

		// Killer moves
		if move == killerMoves[0] {
			score += KillerMove
		} else if move == killerMoves[1] {
			score += KillerMove - 1
		}

		// History heuristic could be added here

		scoredMoves[i] = ScoredMove{Move: move, Score: score}
	}

	// Sort moves based on their scores
	sort.Slice(scoredMoves, func(i, j int) bool {
		return scoredMoves[i].Score > scoredMoves[j].Score
	})

	// Extract sorted moves
	sortedMoves := make([]board.Move, len(moves))
	for i, sm := range scoredMoves {
		sortedMoves[i] = sm.Move
	}

	return sortedMoves
}

// MVV_LVA = (Most Valuable Victim - Least Valuable Attacker)
func MVV_LVA(p *board.Position, m board.Move) int {
	victim := p.Pieces[m.TargetIndex]
	attacker := p.Pieces[m.StartIndex]
	return PieceValue(victim) - PieceValue(attacker)/100
}
