package engine

import "endtner.dev/nChess/internal/board"

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

func (tt *TranspositionTable) Store(key uint64, depth int, score float64, entryType EntryType, move board.Move) {
	index := key % TableSize
	tt.table[index] = Entry{Key: key, Depth: depth, Score: score, Type: entryType, Move: move}
}

func (tt *TranspositionTable) Probe(key uint64) (Entry, bool) {
	index := key % TableSize
	entry := tt.table[index]
	if entry.Key == key {
		return entry, true
	}
	return Entry{}, false
}
