package board

import (
	"math/bits"
)

type State struct {
	Bitboards []uint64
	Pieces    []uint8

	CastlingAvailability  uint8
	EnPassantTargetSquare int
	HalfMoves             int
	FullMoves             int
	Zobrist               uint64
}

type Position struct {
	// Bitboards are stored from bottom left to top right, meaning A1 to H8
	// Bitboards are stored at the index of the piece they have
	Bitboards []uint64
	Pieces    []uint8

	// Current player
	WhiteToMove       bool
	FriendlyColor     uint8
	OpponentColor     uint8
	FriendlyIndex     int
	OpponentIndex     int
	FriendlyKingIndex int
	OpponentKingIndex int
	PawnOffset        int
	PromotionRank     int

	// Metadata
	CastlingRights  uint8 // Bits set like KQkq
	EnPassantSquare int
	HalfMoves       int
	FullMoves       int

	history []State
	LastPos *Position

	Zobrist uint64
}

func (p *Position) Copy() *Position {
	np := Position{
		Bitboards:         make([]uint64, len(p.Bitboards)),
		Pieces:            make([]uint8, len(p.Pieces)),
		WhiteToMove:       p.WhiteToMove,
		FriendlyColor:     p.FriendlyColor,
		OpponentColor:     p.OpponentColor,
		FriendlyIndex:     p.FriendlyIndex,
		OpponentIndex:     p.OpponentIndex,
		FriendlyKingIndex: p.FriendlyKingIndex,
		OpponentKingIndex: p.OpponentKingIndex,
		PawnOffset:        p.PawnOffset,
		PromotionRank:     p.PromotionRank,
		CastlingRights:    p.CastlingRights,
		EnPassantSquare:   p.EnPassantSquare,
		HalfMoves:         p.HalfMoves,
		FullMoves:         p.FullMoves,
		Zobrist:           p.Zobrist,
	}
	copy(np.Bitboards, p.Bitboards)
	copy(np.Pieces, p.Pieces)
	return &np
}

func (p *Position) OtherColorToMove() {
	p.WhiteToMove = !p.WhiteToMove

	p.FriendlyColor, p.OpponentColor = p.OpponentColor, p.FriendlyColor
	p.FriendlyIndex, p.OpponentIndex = p.OpponentIndex, p.FriendlyIndex

	p.FriendlyKingIndex = bits.TrailingZeros64(p.Bitboards[p.FriendlyColor|King])
	p.OpponentKingIndex = bits.TrailingZeros64(p.Bitboards[p.OpponentColor|King])

	p.PawnOffset *= -1

	p.PromotionRank = 7 - p.PromotionRank
}
