package board

import "math/rand"

var ZobristColorToMove uint64
var ZobristTable [64][0b1111]uint64
var ZobristEnPassant [64]uint64
var ZobristCastlingRights [16]uint64

var ZobristReady = func() bool {
	r := rand.New(rand.NewSource(25042024))
	ZobristColorToMove = r.Uint64()

	// Initializing zobrist
	for i := range len(ZobristTable) {
		for j := range len(ZobristTable[0]) {
			ZobristTable[i][j] = r.Uint64()
		}
	}

	for i := range len(ZobristEnPassant) {
		ZobristEnPassant[i] = r.Uint64()
	}

	for i := range len(ZobristCastlingRights) {
		ZobristCastlingRights[i] = r.Uint64()
	}

	return true
}()

func GetZobrist(p *Position) uint64 {
	var zobrist uint64 = 0

	if !p.WhiteToMove {
		zobrist ^= ZobristColorToMove
	}

	zobrist ^= ZobristCastlingRights[p.CastlingRights]

	if p.EnPassantSquare != -1 {
		zobrist ^= ZobristEnPassant[p.EnPassantSquare]
	}

	for i, p := range p.Pieces {
		if p != 0 {
			zobrist ^= ZobristTable[i][p]
		}
	}

	return zobrist
}
