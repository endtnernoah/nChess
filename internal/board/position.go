package board

import (
	"math/bits"
)

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

	IsTerminal     bool
	TerminalReason string

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
		IsTerminal:        p.IsTerminal,
		TerminalReason:    p.TerminalReason,
		LastPos:           p.LastPos,
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

func (p *Position) UpdateTerminalState(hasLegalMoves, isInCheck bool) {
	p.IsTerminal = false
	p.TerminalReason = ""

	if !hasLegalMoves {
		if isInCheck {
			p.IsTerminal = true
			if p.WhiteToMove {
				p.TerminalReason = "Black wins by checkmate"
			} else {
				p.TerminalReason = "White wins by checkmate"
			}
		} else {
			p.IsTerminal = true
			p.TerminalReason = "Draw by stalemate"
		}
	} else if p.IsInsufficientMaterial() {
		p.IsTerminal = true
		p.TerminalReason = "Draw by insufficient material"
	} else if p.IsFiftyMoveRule() {
		p.IsTerminal = true
		p.TerminalReason = "Draw by fifty-move rule"
	}
}

func (p *Position) IsInsufficientMaterial() bool {
	whitePieces := p.Bitboards[White|Knight] | p.Bitboards[White|Bishop] | p.Bitboards[White|Rook] | p.Bitboards[White|Queen] | p.Bitboards[White|Pawn]
	blackPieces := p.Bitboards[Black|Knight] | p.Bitboards[Black|Bishop] | p.Bitboards[Black|Rook] | p.Bitboards[Black|Queen] | p.Bitboards[Black|Pawn]

	// King vs King
	if whitePieces == 0 && blackPieces == 0 {
		return true
	}

	// King and Bishop vs King or King and Knight vs King
	if (whitePieces == p.Bitboards[White|Bishop] || whitePieces == p.Bitboards[White|Knight]) && blackPieces == 0 {
		return true
	}
	if (blackPieces == p.Bitboards[Black|Bishop] || blackPieces == p.Bitboards[Black|Knight]) && whitePieces == 0 {
		return true
	}

	return false
}

func (p *Position) IsFiftyMoveRule() bool {
	return p.HalfMoves >= 100 // 50 full moves = 100 half moves
}
