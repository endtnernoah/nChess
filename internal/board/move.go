package board

import (
	"fmt"
)

type Move struct {
	StartIndex             int
	TargetIndex            int
	EnPassantCaptureSquare int
	EnPassantPassedSquare  int
	RookStartingSquare     int
	PromotionPiece         uint8
}

type OptionalParameter func(*Move)

func WithEnPassantCaptureSquare(square int) OptionalParameter {
	return func(m *Move) {
		m.EnPassantCaptureSquare = square
	}
}

func WithEnPassantPassedSquare(square int) OptionalParameter {
	return func(m *Move) {
		m.EnPassantPassedSquare = square
	}
}

func WithRookStartingSquare(square int) OptionalParameter {
	return func(m *Move) {
		m.RookStartingSquare = square
	}
}

func WithPromotion(promotionPiece uint8) OptionalParameter {
	return func(m *Move) {
		m.PromotionPiece = promotionPiece
	}
}

func NewMove(startIndex int, targetIndex int, optionalParameters ...OptionalParameter) Move {
	m := Move{
		StartIndex:             startIndex,
		TargetIndex:            targetIndex,
		EnPassantCaptureSquare: -1,
		EnPassantPassedSquare:  -1,
		RookStartingSquare:     -1,
		PromotionPiece:         0,
	}

	for _, optionalParameter := range optionalParameters {
		optionalParameter(&m)
	}

	return m
}

func MoveToString(m Move) string {
	return fmt.Sprintf("%s%s%s", IndexToSquare(m.StartIndex), IndexToSquare(m.TargetIndex), ToString(m.PromotionPiece))
}

// Making a move

func (p *Position) MakeMove(m Move) *Position {
	np := p.Copy()
	np.LastPos = p

	// Zobrist: Switch color
	np.Zobrist ^= ZobristColorToMove

	// Set new castling availability
	kingSideRookStart := 7
	queenSideRookStart := 0
	kingStart := 4
	kingSideBitIndex := 3
	queenSideBitIndex := 2

	if !np.WhiteToMove {
		kingSideRookStart += 56
		queenSideRookStart += 56
		kingStart += 56
		kingSideBitIndex = 1
		queenSideBitIndex = 0
	}

	// Zobrist: Hash out old castling rights
	np.Zobrist ^= ZobristCastlingRights[np.CastlingRights]

	if m.StartIndex == kingSideRookStart {
		np.CastlingRights = np.CastlingRights & ^(1 << kingSideBitIndex)
	} else if m.StartIndex == queenSideRookStart {
		np.CastlingRights = np.CastlingRights & ^(1 << queenSideBitIndex)
	} else if m.StartIndex == kingStart {
		np.CastlingRights = np.CastlingRights & ^(1 << kingSideBitIndex)
		np.CastlingRights = np.CastlingRights & ^(1 << queenSideBitIndex)
	}

	// Zobrist: Hash in new castling rights
	np.Zobrist ^= ZobristCastlingRights[np.CastlingRights]

	// Zobrist: Hashing out current EP Target Square
	if np.EnPassantSquare != -1 {
		np.Zobrist ^= ZobristEnPassant[np.EnPassantSquare]
	}

	if m.EnPassantPassedSquare != -1 {
		np.EnPassantSquare = m.EnPassantPassedSquare

		// Zobrist: Hashing in new EP Target Square
		np.Zobrist ^= ZobristEnPassant[np.EnPassantSquare]
	} else if np.EnPassantSquare != -1 {
		np.EnPassantSquare = -1
	}

	// Increase half move if not a pawn move and not a capture
	movedPieceType := np.Pieces[m.StartIndex] & 0b00111
	targetPiece := np.Pieces[m.TargetIndex]
	if movedPieceType == Pawn || targetPiece != 0 {
		np.HalfMoves = 0
	} else {
		np.HalfMoves += 1
	}

	// Increase the move number on blacks turns
	if !np.WhiteToMove {
		np.FullMoves += 1
	}

	// Actually move the piece on board
	movedPiece := np.Pieces[m.StartIndex]

	// Handle Castling
	if m.RookStartingSquare != -1 {

		// Move king to target square
		np.Pieces[m.StartIndex] = 0
		np.Pieces[m.TargetIndex] = movedPiece
		np.Bitboards[movedPiece] = (np.Bitboards[movedPiece] & ^(1 << m.StartIndex)) | (1 << m.TargetIndex)

		// Zobrist: Update moved king
		np.Zobrist ^= ZobristTable[m.StartIndex][movedPiece]
		np.Zobrist ^= ZobristTable[m.TargetIndex][movedPiece]

		movedRook := np.Pieces[m.RookStartingSquare]

		// Remove rook from pieces
		np.Pieces[m.RookStartingSquare] = 0

		kingSideTargetSquare := m.TargetIndex - 1
		queenSideTargetSquare := m.TargetIndex + 1

		isKingSideCastle := m.TargetIndex%8 == 6
		if isKingSideCastle {
			np.Pieces[kingSideTargetSquare] = movedRook
			np.Bitboards[movedRook] = (np.Bitboards[movedRook] & ^(1 << m.RookStartingSquare)) | (1 << kingSideTargetSquare)

			// Zobrist: Update moved rook
			np.Zobrist ^= ZobristTable[m.RookStartingSquare][movedRook]
			np.Zobrist ^= ZobristTable[kingSideTargetSquare][movedRook]
		} else {
			np.Pieces[queenSideTargetSquare] = movedRook
			np.Bitboards[movedRook] = (np.Bitboards[movedRook] & ^(1 << m.RookStartingSquare)) | (1 << queenSideTargetSquare)

			// Zobrist: Update moved rook
			np.Zobrist ^= ZobristTable[m.RookStartingSquare][movedRook]
			np.Zobrist ^= ZobristTable[queenSideTargetSquare][movedRook]
		}
	} else {
		// Remove piece from source square
		np.Pieces[m.StartIndex] = 0
		np.Bitboards[movedPiece] &= ^(1 << m.StartIndex)

		// Possibly remove captured piece
		capturedPiece := np.Pieces[m.TargetIndex]
		if capturedPiece != 0 && ((capturedPiece&0b11000)&(movedPiece&0b11000)) == 0 {
			np.Pieces[m.TargetIndex] = 0
			np.Bitboards[capturedPiece] &= ^(1 << m.TargetIndex)

			// Zobrist: Update captured piece
			np.Zobrist ^= ZobristTable[m.TargetIndex][capturedPiece]
		}

		// Possibly remove EP captured piece
		if m.EnPassantCaptureSquare != -1 {
			epCapturedPiece := np.Pieces[m.EnPassantCaptureSquare]
			if epCapturedPiece != 0 && ((epCapturedPiece&0b11000)&(movedPiece&0b11000)) == 0 {
				np.Pieces[m.EnPassantCaptureSquare] = 0
				np.Bitboards[epCapturedPiece] &= ^(1 << m.EnPassantCaptureSquare)

				// Zobrist: Update EP Capture
				np.Zobrist ^= ZobristTable[m.EnPassantCaptureSquare][epCapturedPiece]
			}
		}

		// Add new piece on the target square
		if m.PromotionPiece != 0 {
			// Add newly promoted piece
			np.Pieces[m.TargetIndex] = m.PromotionPiece
			np.Bitboards[m.PromotionPiece] |= 1 << m.TargetIndex

			// Zobrist: Update moved piece promotion
			np.Zobrist ^= ZobristTable[m.StartIndex][movedPiece]
			np.Zobrist ^= ZobristTable[m.TargetIndex][m.PromotionPiece]
		} else {
			// Updating piece position
			np.Pieces[m.TargetIndex] = movedPiece
			np.Bitboards[movedPiece] |= 1 << m.TargetIndex

			// Zobrist: Update moved piece
			np.Zobrist ^= ZobristTable[m.StartIndex][movedPiece]
			np.Zobrist ^= ZobristTable[m.TargetIndex][movedPiece]
		}
	}

	// Switch around the color
	np.OtherColorToMove()

	return np
}

func (p *Position) UnmakeMove() *Position {
	return p.LastPos
}
