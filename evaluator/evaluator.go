package evaluator

import (
	"endtner.dev/nChess/board"
	"endtner.dev/nChess/board/piece"
	"math/bits"
)

// Weights
const (
	// Material weights per Larry Kaufmann, 2012 (https://www.talkchess.com/forum/viewtopic.php?topic_view=threads&p=487051&t=45512)
	PawnValue   = 100
	KnightValue = 350
	BishopValue = 350
	RookValue   = 525
	QueenValue  = 1000
)

func Evaluate(b *board.Board) int {
	score := 0

	pawnDiff := bits.OnesCount64(b.Bitboards[b.FriendlyColor|piece.Pawn]) - bits.OnesCount64(b.Bitboards[b.OpponentColor|piece.Pawn])
	knightDiff := bits.OnesCount64(b.Bitboards[b.FriendlyColor|piece.Knight]) - bits.OnesCount64(b.Bitboards[b.OpponentColor|piece.Knight])
	bishopDiff := bits.OnesCount64(b.Bitboards[b.FriendlyColor|piece.Bishop]) - bits.OnesCount64(b.Bitboards[b.OpponentColor|piece.Bishop])
	rookDiff := bits.OnesCount64(b.Bitboards[b.FriendlyColor|piece.Rook]) - bits.OnesCount64(b.Bitboards[b.OpponentColor|piece.Rook])
	queenDiff := bits.OnesCount64(b.Bitboards[b.FriendlyColor|piece.Queen]) - bits.OnesCount64(b.Bitboards[b.OpponentColor|piece.Queen])

	score += 0 +
		pawnDiff*PawnValue +
		knightDiff*KnightValue +
		bishopDiff*BishopValue +
		rookDiff*RookValue +
		queenDiff*QueenValue

	return score
}
