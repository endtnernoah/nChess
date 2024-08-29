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

type Board struct {
	// Bitboards are stored from bottom left to top right, meaning A1 to H8
	// Bitboards are stored at the index of the piece they have
	bitboards []uint64

	Pieces []uint64 // Bonus: bitboard of allOccupiedFields at index 2

	bitboardStack [][]uint64
}

func New(fenString string) *Board {
	// Since fen starts at a8, we want to split it at the /, reverse the list and join it back with /
	fenRows := strings.Split(fenString, "/")
	slices.Reverse(fenRows)
	fenString = strings.Join(fenRows, "/")

	b := Board{}
	b.bitboards = make([]uint64, 0b10111)

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

		// Matching character, setting bits
		b.bitboards[piece.Value(rune(currentChar))] |= 1 << boardPosition

		boardPosition++
	}

	// Setting up the empty slices
	b.Pieces = make([]uint64, 3)

	// Running the first BB precomputation
	b.ComputeBitboards()

	return &b
}

func (b *Board) ToFEN() string {
	var fen strings.Builder
	emptySquares := 0

	// Piece placement
	for rank := 7; rank >= 0; rank-- {
		for file := 0; file < 8; file++ {
			index := rank*8 + file

			pieceValue := b.PieceAtIndex(index)

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
	copiedBitboards := make([]uint64, len(b.bitboards))
	copy(copiedBitboards, b.bitboards)
	b.bitboardStack = append(b.bitboardStack, copiedBitboards)

	// Handle Castling
	if m.RookStartingSquare != -1 {
		b.makeCastlingMove(m)
		return
	}

	// Get moved piece
	movedPiece := b.PieceAtIndex(m.StartIndex)

	// Remove piece from source square
	b.setPieceBitboard(movedPiece, b.PieceBitboard(movedPiece) & ^(1<<m.StartIndex))

	// Possibly remove captured piece
	capturedPiece := b.PieceAtIndex(m.TargetIndex)
	if capturedPiece != 0 && ((capturedPiece&0b11000)&(movedPiece&0b11000)) == 0 {
		b.setPieceBitboard(capturedPiece, b.PieceBitboard(capturedPiece) & ^(1<<m.TargetIndex))
	}

	// Possibly remove EP captured piece
	if m.EnPassantCaptureSquare != -1 {
		epCapturedPiece := b.PieceAtIndex(m.EnPassantCaptureSquare)
		if epCapturedPiece != 0 && ((epCapturedPiece&0b11000)&(movedPiece&0b11000)) == 0 {
			b.setPieceBitboard(epCapturedPiece, b.PieceBitboard(epCapturedPiece) & ^(1<<m.EnPassantCaptureSquare))
		}
	}

	// Add new piece on the target square
	if m.PromotionPiece != 0 {
		// Add newly promoted piece if flag is set
		b.setPieceBitboard(m.PromotionPiece, b.PieceBitboard(m.PromotionPiece)|(1<<m.TargetIndex))
	} else {
		// Add piece to its own bitboard
		b.setPieceBitboard(movedPiece, b.PieceBitboard(movedPiece)|(1<<m.TargetIndex))
	}
}

func (b *Board) makeCastlingMove(m move.Move) {
	// Get moved piece
	movedPiece := b.PieceAtIndex(m.StartIndex)

	// Move king to target square
	b.setPieceBitboard(movedPiece, b.PieceBitboard(movedPiece) & ^(1<<m.StartIndex))
	b.setPieceBitboard(movedPiece, b.PieceBitboard(movedPiece)|(1<<m.TargetIndex))

	movedRook := b.PieceAtIndex(m.RookStartingSquare)

	kingSideTargetSquare := m.TargetIndex - 1
	queenSideTargetSquare := m.TargetIndex + 1

	isKingSideCastle := m.TargetIndex%8 == 6
	if isKingSideCastle {
		b.setPieceBitboard(movedRook, b.PieceBitboard(movedRook) & ^(1<<m.RookStartingSquare))
		b.setPieceBitboard(movedRook, b.PieceBitboard(movedRook)|(1<<kingSideTargetSquare))
	} else {
		b.setPieceBitboard(movedRook, b.PieceBitboard(movedRook) & ^(1<<m.RookStartingSquare))
		b.setPieceBitboard(movedRook, b.PieceBitboard(movedRook)|(1<<queenSideTargetSquare))
	}
}

func (b *Board) UnmakeMove() {
	if len(b.bitboardStack) == 0 {
		return
	}

	b.bitboards = b.bitboardStack[len(b.bitboardStack)-1]
	b.bitboardStack = b.bitboardStack[:len(b.bitboardStack)-1]
}

func (b *Board) PieceBitboard(piece uint) uint64 {
	return b.bitboards[piece]
}

func (b *Board) setPieceBitboard(piece uint, bitboard uint64) { b.bitboards[piece] = bitboard }

func (b *Board) PieceAtIndex(index int) uint {
	if index < 0 || index > 63 {
		return 0 // Invalid index
	}

	for pieceType := uint(0); pieceType < 7; pieceType++ {
		if boardhelper.IsIndexBitSet(index, b.PieceBitboard(pieceType|piece.ColorWhite)) {
			return pieceType | piece.ColorWhite
		}
		if boardhelper.IsIndexBitSet(index, b.PieceBitboard(pieceType|piece.ColorBlack)) {
			return pieceType | piece.ColorBlack
		}
	}

	return 0 // No piece found
}

func (b *Board) ComputeBitboards() {
	var whitePieces uint64
	var blackPieces uint64

	for p, bitboard := range b.bitboards {
		if uint(p)&piece.ColorWhite == piece.ColorWhite {
			whitePieces |= bitboard
		}
		if uint(p)&piece.ColorBlack == piece.ColorBlack {
			blackPieces |= bitboard
		}
	}

	b.Pieces[0] = whitePieces
	b.Pieces[1] = blackPieces
	b.Pieces[2] = whitePieces | blackPieces
}

func (b *Board) IsPinnedMoveAlongRay(colorToMove uint, m move.Move) bool {
	var rayBitboard uint64

	ownKingIndex := bits.TrailingZeros64(b.PieceBitboard(colorToMove | piece.TypeKing))
	enemyPieces := b.Pieces[1-((colorToMove>>3)-1)]

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

	ownKingBitboard := b.PieceBitboard(colorToMove | piece.TypeKing)
	ownKingIndex := bits.TrailingZeros64(ownKingBitboard)

	// Can instantly return if there is no direct ray between ownKingIndex & enPassantCaptureSquare
	offset := boardhelper.CalculateRayOffset(ownKingIndex, m.EnPassantCaptureSquare)
	if offset == 0 {
		return false
	}

	enemyAttackers := b.PieceBitboard(enemyColor | piece.TypeQueen)
	isValidMoveFunction := boardhelper.IsValidStraightMove
	switch offset {
	case -1, 1, -8, 8:
		enemyAttackers |= b.PieceBitboard(enemyColor | piece.TypeRook)
	case -7, 7, -9, 9:
		enemyAttackers |= b.PieceBitboard(enemyColor | piece.TypeBishop)
		isValidMoveFunction = boardhelper.IsValidDiagonalMove
	default:
		return false
	}

	otherPieces := b.Pieces[2] & ^(enemyAttackers)

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
