package engine

import (
	"endtner.dev/nChess/internal/board"
	"math"
)

const (
	TableSize = 1024 * 1024 * 1024 / 32 // Number of entries for 1GB table
)

type EntryType byte

const (
	ExactScore EntryType = iota
	LowerBound
	UpperBound
)

type Entry struct {
	Key   uint64     // Zobrist hash of the position
	Depth int        // Depth of the search when this entry was created
	Score float64    // Evaluation score
	Type  EntryType  // Type of the score (exact, lower bound, upper bound)
	Move  board.Move // Best move found for this position
}

type TranspositionTable struct {
	table []Entry
}

func NewTranspositionTable() *TranspositionTable {
	return &TranspositionTable{
		table: make([]Entry, TableSize),
	}
}

func (tt *TranspositionTable) Store(key uint64, depth int, score, alpha0, beta float64, move board.Move) {
	var entryType EntryType
	if score <= alpha0 {
		entryType = UpperBound
	} else if score >= beta {
		entryType = LowerBound
	} else {
		entryType = ExactScore
	}

	index := key % TableSize
	tt.table[index] = Entry{Key: key, Depth: depth, Score: score, Type: entryType, Move: move}
}

func (tt *TranspositionTable) Query(key uint64, depth int, alpha, beta float64) (board.Move, bool, float64) {
	ttMove := board.Move{}
	entry := tt.table[key%TableSize]
	found := entry.Key != 0

	if found && entry.Depth >= depth {
		ttMove = entry.Move

		if entry.Type == ExactScore {
			return ttMove, true, entry.Score
		} else if entry.Type == LowerBound {
			alpha = math.Max(alpha, entry.Score)
		} else if entry.Type == UpperBound {
			beta = math.Min(beta, entry.Score)
		}

		// Move already got evaluated better than the current search
		if alpha >= beta {
			return ttMove, true, entry.Score
		}
	}
	return ttMove, false, 0
}

func (tt *TranspositionTable) Probe(key uint64) (Entry, bool) {
	index := key % TableSize
	entry := tt.table[index]
	if entry.Key == key {
		return entry, true
	}
	return Entry{}, false
}
