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

// Piece-Square-Tables
var (
	PawnTableMidgame = []int{
		0, 0, 0, 0, 0, 0, 0, 0,
		50, 50, 50, 50, 50, 50, 50, 50,
		10, 10, 20, 30, 30, 20, 10, 10,
		5, 5, 10, 25, 25, 10, 5, 5,
		0, 0, 0, 20, 20, 0, 0, 0,
		5, -5, -10, 0, 0, -10, -5, 5,
		5, 10, 10, -20, -20, 10, 10, 5,
		0, 0, 0, 0, 0, 0, 0, 0,
	}
	KnightTableMidgame = []int{
		-50, -40, -30, -30, -30, -30, -40, -50,
		-40, -20, 0, 0, 0, 0, -20, -40,
		-30, 0, 10, 15, 15, 10, 0, -30,
		-30, 5, 15, 20, 20, 15, 5, -30,
		-30, 0, 15, 20, 20, 15, 0, -30,
		-30, 5, 10, 15, 15, 10, 5, -30,
		-40, -20, 0, 5, 5, 0, -20, -40,
		-50, -40, -30, -30, -30, -30, -40, -50,
	}
	BishopTableMidgame = []int{
		-20, -10, -10, -10, -10, -10, -10, -20,
		-10, 0, 0, 0, 0, 0, 0, -10,
		-10, 0, 5, 10, 10, 5, 0, -10,
		-10, 5, 5, 10, 10, 5, 5, -10,
		-10, 0, 10, 10, 10, 10, 0, -10,
		-10, 10, 10, 10, 10, 10, 10, -10,
		-10, 5, 0, 0, 0, 0, 5, -10,
		-20, -10, -10, -10, -10, -10, -10, -20,
	}
	RookTableMidgame = []int{
		0, 0, 0, 0, 0, 0, 0, 0,
		5, 10, 10, 10, 10, 10, 10, 5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		0, 0, 0, 5, 5, 0, 0, 0,
	}
	QueenTableMidgame = []int{
		-20, -10, -10, -5, -5, -10, -10, -20,
		-10, 0, 0, 0, 0, 0, 0, -10,
		-10, 0, 5, 5, 5, 5, 0, -10,
		-5, 0, 5, 5, 5, 5, 0, -5,
		0, 0, 5, 5, 5, 5, 0, -5,
		-10, 5, 5, 5, 5, 5, 0, -10,
		-10, 0, 5, 0, 0, 0, 0, -10,
		-20, -10, -10, -5, -5, -10, -10, -20,
	}
	KingTableMidgame = []int{
		-30, -40, -40, -50, -50, -40, -40, -30,
		-30, -40, -40, -50, -50, -40, -40, -30,
		-30, -40, -40, -50, -50, -40, -40, -30,
		-30, -40, -40, -50, -50, -40, -40, -30,
		-20, -30, -30, -40, -40, -30, -30, -20,
		-10, -20, -20, -20, -20, -20, -20, -10,
		20, 20, 0, 0, 0, 0, 20, 20,
		20, 30, 10, 0, 0, 10, 30, 20,
	}
	KingTableEndgame = []int{
		-50, -40, -30, -20, -20, -30, -40, -50,
		-30, -20, -10, 0, 0, -10, -20, -30,
		-30, -10, 20, 30, 30, 20, -10, -30,
		-30, -10, 30, 40, 40, 30, -10, -30,
		-30, -10, 30, 40, 40, 30, -10, -30,
		-30, -10, 20, 30, 30, 20, -10, -30,
		-30, -30, 0, 0, 0, 0, -30, -30,
		-50, -30, -30, -30, -30, -30, -30, -50,
	}
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
	gamePhase := calculateGamePhase(p)

	score += evaluateMaterial(p)
	score += evaluatePieceSquareTables(p, gamePhase)

	return float64(score) / 100
}

func evaluateMaterial(p *board.Position) int {
	score := 0
	for piece := board.Pawn; piece <= board.Queen; piece++ {
		friendlyCount := bits.OnesCount64(p.Bitboards[p.FriendlyColor|piece])
		opponentCount := bits.OnesCount64(p.Bitboards[p.OpponentColor|piece])
		score += (friendlyCount - opponentCount) * PieceValue(piece)
	}
	return score
}

func evaluatePieceSquareTables(p *board.Position, gamePhase float64) int {
	score := 0

	for piece := board.Pawn; piece <= board.King; piece++ {
		friendlyPieces := p.Bitboards[p.FriendlyColor|piece]
		opponentPieces := p.Bitboards[p.OpponentColor|piece]

		for friendlyPieces != 0 {
			square := bits.TrailingZeros64(friendlyPieces)
			score += getPieceSquareBonus(piece, square, p.FriendlyColor, gamePhase)
			friendlyPieces &= friendlyPieces - 1
		}

		for opponentPieces != 0 {
			square := bits.TrailingZeros64(opponentPieces)
			score -= getPieceSquareBonus(piece, square, p.OpponentColor, gamePhase)
			opponentPieces &= opponentPieces - 1
		}
	}

	return score
}

func getPieceSquareBonus(piece uint8, square int, color uint8, gamePhase float64) int {
	if color == board.Black {
		square = 63 - square // Flip square for black pieces
	}

	switch piece {
	case board.Pawn:
		return PawnTableMidgame[square]
	case board.Knight:
		return KnightTableMidgame[square]
	case board.Bishop:
		return BishopTableMidgame[square]
	case board.Rook:
		return RookTableMidgame[square]
	case board.Queen:
		return QueenTableMidgame[square]
	case board.King:
		return int(float64(KingTableMidgame[square])*(1-gamePhase) + float64(KingTableEndgame[square])*gamePhase)
	default:
		return 0
	}
}

func calculateGamePhase(p *board.Position) float64 {
	totalMaterial := 0
	for piece := board.Knight; piece <= board.Queen; piece++ {
		totalMaterial += bits.OnesCount64(p.Bitboards[board.White|piece]|p.Bitboards[board.Black|piece]) * PieceValue(piece)
	}

	// Define thresholds for midgame and endgame
	midgameThreshold := 2 * (4*PieceValue(board.Knight) + 4*PieceValue(board.Bishop) + 4*PieceValue(board.Rook) + 2*PieceValue(board.Queen))
	endgameThreshold := 2 * (3*PieceValue(board.Knight) + 3*PieceValue(board.Bishop) + 2*PieceValue(board.Rook) + 1*PieceValue(board.Queen))

	if totalMaterial >= midgameThreshold {
		return 0 // Midgame
	} else if totalMaterial <= endgameThreshold {
		return 1 // Endgame
	} else {
		// Linear interpolation between midgame and endgame
		return 1 - float64(totalMaterial-endgameThreshold)/float64(midgameThreshold-endgameThreshold)
	}
}
