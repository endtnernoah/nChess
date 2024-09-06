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

	Zobrist uint64
}

func (p *Position) MakeMove(m Move) {
	// Update the stack
	currentState := State{}

	currentState.Pieces = make([]uint8, len(p.Pieces))
	currentState.Bitboards = make([]uint64, len(p.Bitboards))

	copy(currentState.Pieces, p.Pieces)
	copy(currentState.Bitboards, p.Bitboards)

	currentState.CastlingAvailability = p.CastlingRights
	currentState.EnPassantTargetSquare = p.EnPassantSquare
	currentState.HalfMoves = p.HalfMoves
	currentState.FullMoves = p.FullMoves
	currentState.Zobrist = p.Zobrist

	p.history = append(p.history, currentState)

	// Zobrist: Switch color
	p.Zobrist ^= ZobristColorToMove

	// Set new castling availability
	kingSideRookStart := 7
	queenSideRookStart := 0
	kingStart := 4
	kingSideBitIndex := 3
	queenSideBitIndex := 2

	if !p.WhiteToMove {
		kingSideRookStart += 56
		queenSideRookStart += 56
		kingStart += 56
		kingSideBitIndex = 1
		queenSideBitIndex = 0
	}

	// Zobrist: Hash out old castling rights
	p.Zobrist ^= ZobristCastlingRights[p.CastlingRights]

	if m.StartIndex == kingSideRookStart {
		p.CastlingRights = p.CastlingRights & ^(1 << kingSideBitIndex)
	} else if m.StartIndex == queenSideRookStart {
		p.CastlingRights = p.CastlingRights & ^(1 << queenSideBitIndex)
	} else if m.StartIndex == kingStart {
		p.CastlingRights = p.CastlingRights & ^(1 << kingSideBitIndex)
		p.CastlingRights = p.CastlingRights & ^(1 << queenSideBitIndex)
	}

	// Zobrist: Hash in new castling rights
	p.Zobrist ^= ZobristCastlingRights[p.CastlingRights]

	// Zobrist: Hashing out current EP Target Square
	if p.EnPassantSquare != -1 {
		p.Zobrist ^= ZobristEnPassant[p.EnPassantSquare]
	}

	if m.EnPassantPassedSquare != -1 {
		p.EnPassantSquare = m.EnPassantPassedSquare

		// Zobrist: Hashing in new EP Target Square
		p.Zobrist ^= ZobristEnPassant[p.EnPassantSquare]
	} else if p.EnPassantSquare != -1 {
		p.EnPassantSquare = -1
	}

	// Increase half move if not a pawn move and not a capture
	movedPieceType := p.Pieces[m.StartIndex] & 0b00111
	targetPiece := p.Pieces[m.TargetIndex]
	if movedPieceType == Pawn || targetPiece != 0 {
		p.HalfMoves = 0
	} else {
		p.HalfMoves += 1
	}

	// Increase the move number on blacks turns
	if !p.WhiteToMove {
		p.FullMoves += 1
	}

	// Actually move the piece on board
	movedPiece := p.Pieces[m.StartIndex]

	// Handle Castling
	if m.RookStartingSquare != -1 {

		// Move king to target square
		p.Pieces[m.StartIndex] = 0
		p.Pieces[m.TargetIndex] = movedPiece
		p.Bitboards[movedPiece] = (p.Bitboards[movedPiece] & ^(1 << m.StartIndex)) | (1 << m.TargetIndex)

		// Zobrist: Update moved king
		p.Zobrist ^= ZobristTable[m.StartIndex][movedPiece]
		p.Zobrist ^= ZobristTable[m.TargetIndex][movedPiece]

		movedRook := p.Pieces[m.RookStartingSquare]

		// Remove rook from pieces
		p.Pieces[m.RookStartingSquare] = 0

		kingSideTargetSquare := m.TargetIndex - 1
		queenSideTargetSquare := m.TargetIndex + 1

		isKingSideCastle := m.TargetIndex%8 == 6
		if isKingSideCastle {
			p.Pieces[kingSideTargetSquare] = movedRook
			p.Bitboards[movedRook] = (p.Bitboards[movedRook] & ^(1 << m.RookStartingSquare)) | (1 << kingSideTargetSquare)

			// Zobrist: Update moved rook
			p.Zobrist ^= ZobristTable[m.RookStartingSquare][movedRook]
			p.Zobrist ^= ZobristTable[kingSideTargetSquare][movedRook]
		} else {
			p.Pieces[queenSideTargetSquare] = movedRook
			p.Bitboards[movedRook] = (p.Bitboards[movedRook] & ^(1 << m.RookStartingSquare)) | (1 << queenSideTargetSquare)

			// Zobrist: Update moved rook
			p.Zobrist ^= ZobristTable[m.RookStartingSquare][movedRook]
			p.Zobrist ^= ZobristTable[queenSideTargetSquare][movedRook]
		}
	} else {
		// Remove piece from source square
		p.Pieces[m.StartIndex] = 0
		p.Bitboards[movedPiece] &= ^(1 << m.StartIndex)

		// Possibly remove captured piece
		capturedPiece := p.Pieces[m.TargetIndex]
		if capturedPiece != 0 && ((capturedPiece&0b11000)&(movedPiece&0b11000)) == 0 {
			p.Pieces[m.TargetIndex] = 0
			p.Bitboards[capturedPiece] &= ^(1 << m.TargetIndex)

			// Zobrist: Update captured piece
			p.Zobrist ^= ZobristTable[m.TargetIndex][capturedPiece]
		}

		// Possibly remove EP captured piece
		if m.EnPassantCaptureSquare != -1 {
			epCapturedPiece := p.Pieces[m.EnPassantCaptureSquare]
			if epCapturedPiece != 0 && ((epCapturedPiece&0b11000)&(movedPiece&0b11000)) == 0 {
				p.Pieces[m.EnPassantCaptureSquare] = 0
				p.Bitboards[epCapturedPiece] &= ^(1 << m.EnPassantCaptureSquare)

				// Zobrist: Update EP Capture
				p.Zobrist ^= ZobristTable[m.EnPassantCaptureSquare][epCapturedPiece]
			}
		}

		// Add new piece on the target square
		if m.PromotionPiece != 0 {
			// Add newly promoted piece
			p.Pieces[m.TargetIndex] = m.PromotionPiece
			p.Bitboards[m.PromotionPiece] |= 1 << m.TargetIndex

			// Zobrist: Update moved piece promotion
			p.Zobrist ^= ZobristTable[m.StartIndex][movedPiece]
			p.Zobrist ^= ZobristTable[m.TargetIndex][m.PromotionPiece]
		} else {
			// Updating piece position
			p.Pieces[m.TargetIndex] = movedPiece
			p.Bitboards[movedPiece] |= 1 << m.TargetIndex

			// Zobrist: Update moved piece
			p.Zobrist ^= ZobristTable[m.StartIndex][movedPiece]
			p.Zobrist ^= ZobristTable[m.TargetIndex][movedPiece]
		}
	}

	// Switch around the color
	p.OtherColorToMove()
}

func (p *Position) UnmakeMove() {
	if len(p.history) == 0 {
		return
	}

	n := len(p.history)
	lastState := p.history[n-1]

	p.Pieces = lastState.Pieces
	p.Bitboards = lastState.Bitboards
	p.CastlingRights = lastState.CastlingAvailability
	p.EnPassantSquare = lastState.EnPassantTargetSquare
	p.HalfMoves = lastState.HalfMoves
	p.FullMoves = lastState.FullMoves
	p.Zobrist = lastState.Zobrist

	p.history = p.history[:n-1]

	// Change color to move
	p.OtherColorToMove()
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
