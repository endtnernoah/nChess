package board

import (
	"endtner.dev/nChess/game/boardhelper"
	"endtner.dev/nChess/game/move"
	"endtner.dev/nChess/game/piece"
	"fmt"
	"math/bits"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

type State struct {
	Bitboards []uint64
	Pieces    []uint
}

type Board struct {
	// Bitboards are stored from bottom left to top right, meaning A1 to H8
	// Bitboards are stored at the index of the piece they have
	Bitboards          []uint64
	OccupancyBitboards []uint64
	Pieces             []uint

	stateStack []State
}

func New(fenString string) *Board {
	// Since fen starts at a8, we want to split it at the /, reverse the list and join it back with /
	fenRows := strings.Split(fenString, "/")
	slices.Reverse(fenRows)
	fenString = strings.Join(fenRows, "/")

	b := Board{}
	b.Bitboards = make([]uint64, 0b10111)
	b.Pieces = make([]uint, 64)

	// Setting up pieces
	boardPosition := 0
	for i := 0; i < len(fenString); i++ {
		currentChar := fenString[i]

		if unicode.IsNumber(rune(currentChar)) {
			data, err := strconv.Atoi(string(currentChar))
			if err != nil {
				fmt.Printf("Error parsing character '%q'", currentChar)
				panic(err)
			}

			// Skip n - 1 fields
			boardPosition += data

			continue
		}

		// Ignore this fucker
		if rune(currentChar) == '/' {
			continue
		}

		// Populating bitboards & pieces
		p := piece.Value(rune(currentChar))

		b.Pieces[boardPosition] = p
		b.Bitboards[p] |= 1 << boardPosition

		boardPosition++
	}

	// Setting up the empty slices
	b.OccupancyBitboards = make([]uint64, 3)

	// Running the first BB precomputation
	b.ComputeOccupancyBitboards()

	return &b
}

func (b *Board) ToFEN() string {
	var fen strings.Builder
	emptySquares := 0

	// Piece placement
	for rank := 7; rank >= 0; rank-- {
		for file := 0; file < 8; file++ {
			index := rank*8 + file

			pieceValue := b.Pieces[index]

			if pieceValue == 0 {
				emptySquares++
			} else {
				if emptySquares > 0 {
					fen.WriteString(strconv.Itoa(emptySquares))
					emptySquares = 0
				}
				fen.WriteString(piece.ToString(pieceValue))
			}
		}

		if emptySquares > 0 {
			fen.WriteString(strconv.Itoa(emptySquares))
			emptySquares = 0
		}

		if rank > 0 {
			fen.WriteRune('/')
		}
	}

	return fen.String()
}

func (b *Board) MakeMove(m move.Move) {
	// Put current bitboards on the stack
	s := State{}

	s.Pieces = make([]uint, len(b.Pieces))
	s.Bitboards = make([]uint64, len(b.Bitboards))

	copy(s.Pieces, b.Pieces)
	copy(s.Bitboards, b.Bitboards)

	b.stateStack = append(b.stateStack, s)

	// Handle Castling
	if m.RookStartingSquare != -1 {
		b.makeCastlingMove(m)
		return
	}

	// Get moved piece
	movedPiece := b.Pieces[m.StartIndex]

	// Remove piece from source square
	b.Pieces[m.StartIndex] = 0
	b.Bitboards[movedPiece] &= ^(1 << m.StartIndex)

	// Possibly remove captured piece
	capturedPiece := b.Pieces[m.TargetIndex]
	if capturedPiece != 0 && ((capturedPiece&0b11000)&(movedPiece&0b11000)) == 0 {
		b.Pieces[m.TargetIndex] = 0
		b.Bitboards[capturedPiece] &= ^(1 << m.TargetIndex)
	}

	// Possibly remove EP captured piece
	if m.EnPassantCaptureSquare != -1 {
		epCapturedPiece := b.Pieces[m.EnPassantCaptureSquare]
		if epCapturedPiece != 0 && ((epCapturedPiece&0b11000)&(movedPiece&0b11000)) == 0 {
			b.Pieces[m.TargetIndex] = 0
			b.Bitboards[epCapturedPiece] &= ^(1 << m.EnPassantCaptureSquare)
		}
	}

	// Add new piece on the target square
	if m.PromotionPiece != 0 {
		// Add newly promoted piece
		b.Pieces[m.TargetIndex] = m.PromotionPiece
		b.Bitboards[m.PromotionPiece] |= 1 << m.TargetIndex
	} else {
		// Updating piece position
		b.Pieces[m.TargetIndex] = movedPiece
		b.Bitboards[movedPiece] |= 1 << m.TargetIndex
	}
}

func (b *Board) makeCastlingMove(m move.Move) {
	// Get moved piece
	movedPiece := b.Pieces[m.StartIndex]

	// Move king to target square
	b.Pieces[m.StartIndex] = 0
	b.Pieces[m.TargetIndex] = movedPiece
	b.Bitboards[movedPiece] = (b.Bitboards[movedPiece] & ^(1 << m.StartIndex)) | (1 << m.TargetIndex)

	movedRook := b.Pieces[m.RookStartingSquare]

	// Remove rook from pieces
	b.Pieces[m.RookStartingSquare] = 0

	kingSideTargetSquare := m.TargetIndex - 1
	queenSideTargetSquare := m.TargetIndex + 1

	isKingSideCastle := m.TargetIndex%8 == 6
	if isKingSideCastle {
		b.Pieces[kingSideTargetSquare] = movedRook
		b.Bitboards[movedRook] = (b.Bitboards[movedRook] & ^(1 << m.RookStartingSquare)) | (1 << kingSideTargetSquare)
	} else {
		b.Pieces[queenSideTargetSquare] = movedRook
		b.Bitboards[movedRook] = (b.Bitboards[movedRook] & ^(1 << m.RookStartingSquare)) | (1 << queenSideTargetSquare)
	}
}

func (b *Board) UnmakeMove() {
	if len(b.stateStack) == 0 {
		return
	}

	n := len(b.stateStack)
	lastState := b.stateStack[n-1]

	b.Pieces = lastState.Pieces
	b.Bitboards = lastState.Bitboards

	b.stateStack = b.stateStack[:n-1]
}

func (b *Board) PieceAtIndex(index int) uint {
	if index < 0 || index > 63 {
		return 0 // Invalid index
	}

	for pieceType := uint(0); pieceType < 7; pieceType++ {
		if boardhelper.IsIndexBitSet(index, b.Bitboards[pieceType|piece.ColorWhite]) {
			return pieceType | piece.ColorWhite
		}
		if boardhelper.IsIndexBitSet(index, b.Bitboards[pieceType|piece.ColorBlack]) {
			return pieceType | piece.ColorBlack
		}
	}

	return 0 // No piece found
}

func (b *Board) ComputeOccupancyBitboards() {
	var whitePieces uint64
	var blackPieces uint64

	for p, bitboard := range b.Bitboards {
		if uint(p)&piece.ColorWhite == piece.ColorWhite {
			whitePieces |= bitboard
		}
		if uint(p)&piece.ColorBlack == piece.ColorBlack {
			blackPieces |= bitboard
		}
	}

	b.OccupancyBitboards[0] = whitePieces
	b.OccupancyBitboards[1] = blackPieces
	b.OccupancyBitboards[2] = whitePieces | blackPieces
}

func (b *Board) IsPinnedMoveAlongRay(colorToMove uint, m move.Move) bool {
	var rayBitboard uint64

	ownKingIndex := bits.TrailingZeros64(b.Bitboards[colorToMove|piece.TypeKing])
	enemyPieces := b.OccupancyBitboards[1-((colorToMove>>3)-1)]

	rayOffset := boardhelper.CalculateRayOffset(ownKingIndex, m.StartIndex)

	rayIndex := ownKingIndex + rayOffset

	for boardhelper.IsValidStraightMove(ownKingIndex, rayIndex) || boardhelper.IsValidDiagonalMove(ownKingIndex, rayIndex) {
		// Cannot move to our own square
		if rayIndex != m.StartIndex {
			rayBitboard |= 1 << rayIndex
		}

		// Since we know we are pinned, the first enemy piece has to be our attacker
		if boardhelper.IsIndexBitSet(rayIndex, enemyPieces) {
			break
		}

		rayIndex += rayOffset
	}

	return boardhelper.IsIndexBitSet(m.TargetIndex, rayBitboard)
}

func (b *Board) IsEnPassantMovePinned(colorToMove uint, m move.Move) bool {

	enemyColor := piece.ColorBlack
	if colorToMove != piece.ColorWhite {
		enemyColor = piece.ColorWhite
	}

	ownKingIndex := bits.TrailingZeros64(b.Bitboards[colorToMove|piece.TypeKing])

	// Can instantly return if there is no direct ray between ownKingIndex & enPassantCaptureSquare
	offset := boardhelper.CalculateRayOffset(ownKingIndex, m.EnPassantCaptureSquare)
	if offset == 0 {
		return false
	}

	enemyAttackers := b.Bitboards[enemyColor|piece.TypeQueen]
	isValidMoveFunction := boardhelper.IsValidStraightMove
	switch offset {
	case -1, 1, -8, 8:
		enemyAttackers |= b.Bitboards[enemyColor|piece.TypeRook]
	case -7, 7, -9, 9:
		enemyAttackers |= b.Bitboards[enemyColor|piece.TypeBishop]
		isValidMoveFunction = boardhelper.IsValidDiagonalMove
	default:
		return false
	}

	otherPieces := b.OccupancyBitboards[2] & ^(enemyAttackers)

	rayIndex := ownKingIndex + offset

	for isValidMoveFunction(ownKingIndex, rayIndex) {

		// Return if we hit the moves target index (new blocking piece)
		if rayIndex == m.TargetIndex {
			return false
		}

		// Return if we hit an enemy attacker
		if boardhelper.IsIndexBitSet(rayIndex, enemyAttackers) {
			return true
		}

		// Return if we hit any piece that is not on either enPassantCaptureSquare or m.StartIndex
		if rayIndex != m.EnPassantCaptureSquare &&
			rayIndex != m.StartIndex &&
			boardhelper.IsIndexBitSet(rayIndex, otherPieces) {
			return false
		}

		rayIndex += offset
	}

	// Ray was casted until the edge, we can return false
	return false
}
