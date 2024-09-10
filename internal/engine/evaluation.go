package engine

import (
	"endtner.dev/nChess/internal/board"
	"math/bits"
)

// Material weights per Larry Kaufmann, 2012 (https://www.talkchess.com/forum/viewtopic.php?topic_view=threads&p=487051&t=45512)
const (
	PawnValue   = 100
	KnightValue = 350
	BishopValue = 350
	RookValue   = 525
	QueenValue  = 1000
)

func PieceValue(piece uint8) int {
	switch piece {
	case board.Pawn:
		return PawnValue
	case board.Knight:
		return KnightValue
	case board.Bishop:
		return BishopValue
	case board.Rook:
		return RookValue
	case board.Queen:
		return QueenValue
	default:
		return 0
	}
}

func Evaluate(p *board.Position) float64 {
	score := 0

	pawnDiff := bits.OnesCount64(p.Bitboards[p.FriendlyColor|board.Pawn]) - bits.OnesCount64(p.Bitboards[p.OpponentColor|board.Pawn])
	knightDiff := bits.OnesCount64(p.Bitboards[p.FriendlyColor|board.Knight]) - bits.OnesCount64(p.Bitboards[p.OpponentColor|board.Knight])
	bishopDiff := bits.OnesCount64(p.Bitboards[p.FriendlyColor|board.Bishop]) - bits.OnesCount64(p.Bitboards[p.OpponentColor|board.Bishop])
	rookDiff := bits.OnesCount64(p.Bitboards[p.FriendlyColor|board.Rook]) - bits.OnesCount64(p.Bitboards[p.OpponentColor|board.Rook])
	queenDiff := bits.OnesCount64(p.Bitboards[p.FriendlyColor|board.Queen]) - bits.OnesCount64(p.Bitboards[p.OpponentColor|board.Queen])

	score += 0 +
		pawnDiff*PawnValue +
		knightDiff*KnightValue +
		bishopDiff*BishopValue +
		rookDiff*RookValue +
		queenDiff*QueenValue

	return float64(score / 100)
}
