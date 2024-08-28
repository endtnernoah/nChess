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

	AttackFields []uint64 // accessIndex := (3 >> color) - 1, enemyAccessIndex := 1 - accessIndex
	PinnedPieces []uint64
	Pieces       []uint64 // Bonus: bitboard of allOccupiedFields at index 2

	bitboardStack [][]uint64

	BlackPawnAttacks uint64
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
	b.AttackFields = make([]uint64, 2)
	b.PinnedPieces = make([]uint64, 2)

	// Running the first BB precomputation
	b.Update()

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

	// Update all bitboards
	b.Update()
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

	b.Update()
}

func (b *Board) UnmakeMove() {
	if len(b.bitboardStack) == 0 {
		return
	}

	b.bitboards = b.bitboardStack[len(b.bitboardStack)-1]
	b.bitboardStack = b.bitboardStack[:len(b.bitboardStack)-1]

	b.Update()
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

func (b *Board) Update() {
	b.updateBitboards()
	b.updateWhiteAttackBitboard()
	b.updateBlackAttackBitboard()
	b.updateWhitePinnedPieces()
	b.updateBlackPinnedPieces()
}

func (b *Board) updateBitboards() {
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

func (b *Board) updateWhiteAttackBitboard() {
	whitePawnAttacks := b.pawnAttackBitboard(piece.ColorWhite)
	whiteStraightSlidingAttacks := b.straightSlidingAttackBitboard(piece.ColorWhite)
	whiteDiagonalSlidingAttacks := b.diagonalSlidingAttackBitboard(piece.ColorWhite)
	whiteKnightAttacks := b.knightAttackBitboard(piece.ColorWhite)

	b.AttackFields[0] = whitePawnAttacks |
		whiteStraightSlidingAttacks |
		whiteDiagonalSlidingAttacks |
		whiteKnightAttacks
}

func (b *Board) updateBlackAttackBitboard() {
	blackPawnAttacks := b.pawnAttackBitboard(piece.ColorBlack)
	blackStraightSlidingAttacks := b.straightSlidingAttackBitboard(piece.ColorBlack)
	blackDiagonalSlidingAttacks := b.diagonalSlidingAttackBitboard(piece.ColorBlack)
	blackKnightAttacks := b.knightAttackBitboard(piece.ColorBlack)

	b.BlackPawnAttacks = blackPawnAttacks

	b.AttackFields[1] = blackPawnAttacks |
		blackStraightSlidingAttacks |
		blackDiagonalSlidingAttacks |
		blackKnightAttacks
}

func (b *Board) updateWhitePinnedPieces() {
	straightPinnedPieces := b.straightPinnedPiecesBitboard(piece.ColorWhite)
	diagonalPinnedPieces := b.diagonalPinnedPiecesBitboard(piece.ColorWhite)

	b.PinnedPieces[0] = straightPinnedPieces | diagonalPinnedPieces
}

func (b *Board) updateBlackPinnedPieces() {
	straightPinnedPieces := b.straightPinnedPiecesBitboard(piece.ColorBlack)
	diagonalPinnedPieces := b.diagonalPinnedPiecesBitboard(piece.ColorBlack)

	b.PinnedPieces[1] = straightPinnedPieces | diagonalPinnedPieces
}

func (b *Board) CalculateProtectMoves(colorToMove uint) (int, uint64) {
	numAttackers := 0
	var protectMovesBitboard uint64

	enemyColor := piece.ColorBlack
	if colorToMove == piece.ColorBlack {
		enemyColor = piece.ColorWhite
	}

	indexOffsetsStraight := []int{1, -1, 8, -8}
	indexOffsetsDiagonal := []int{-7, 7, -9, 9}
	indexOffsetsKnight := []int{-6, 6, -10, 10, -15, 15, -17, 17}
	indexOffsetsPawns := []int{7, 9}

	if colorToMove == piece.ColorBlack {
		indexOffsetsPawns = []int{-7, -9}
	}

	kingBitboard := b.PieceBitboard(colorToMove | piece.TypeKing)

	enemyStraightAttackers := b.PieceBitboard(enemyColor|piece.TypeRook) | b.PieceBitboard(enemyColor|piece.TypeQueen)
	enemyDiagonalAttackers := b.PieceBitboard(enemyColor|piece.TypeBishop) | b.PieceBitboard(enemyColor|piece.TypeQueen)
	enemyKnightsBitboard := b.PieceBitboard(enemyColor | piece.TypeKnight)
	enemyPawnsBitboard := b.PieceBitboard(enemyColor | piece.TypePawn)

	otherPiecesStraight := b.Pieces[2] & ^enemyStraightAttackers
	otherPiecesDiagonal := b.Pieces[2] & ^enemyDiagonalAttackers

	kingIndex := bits.TrailingZeros64(kingBitboard)

	for _, offset := range indexOffsetsStraight {
		var currentOffsetBitboard uint64
		rayIndex := kingIndex + offset

		for boardhelper.IsValidStraightMove(kingIndex, rayIndex) {
			currentOffsetBitboard |= 1 << rayIndex

			if boardhelper.IsIndexBitSet(rayIndex, otherPiecesStraight) {
				break
			}

			if boardhelper.IsIndexBitSet(rayIndex, enemyStraightAttackers) {
				protectMovesBitboard |= currentOffsetBitboard
				numAttackers++
				break
			}

			rayIndex += offset
		}
	}

	for _, offset := range indexOffsetsDiagonal {
		var currentOffsetBitboard uint64
		rayIndex := kingIndex + offset

		for boardhelper.IsValidDiagonalMove(kingIndex, rayIndex) {
			currentOffsetBitboard |= 1 << rayIndex

			if boardhelper.IsIndexBitSet(rayIndex, otherPiecesDiagonal) {
				break
			}

			if boardhelper.IsIndexBitSet(rayIndex, enemyDiagonalAttackers) {
				protectMovesBitboard |= currentOffsetBitboard
				numAttackers++
				break
			}

			rayIndex += offset
		}
	}

	for _, offset := range indexOffsetsKnight {
		rayIndex := kingIndex + offset

		if !boardhelper.IsValidKnightMove(kingIndex, rayIndex) {
			continue
		}

		if !boardhelper.IsIndexBitSet(rayIndex, enemyKnightsBitboard) {
			continue
		}

		protectMovesBitboard |= 1 << rayIndex
		numAttackers++
	}

	for _, offset := range indexOffsetsPawns {
		rayIndex := kingIndex + offset

		if !boardhelper.IsValidDiagonalMove(kingIndex, rayIndex) {
			continue
		}

		if !boardhelper.IsIndexBitSet(rayIndex, enemyPawnsBitboard) {
			continue
		}

		protectMovesBitboard |= 1 << rayIndex
		numAttackers++
	}

	return numAttackers, protectMovesBitboard
}

func (b *Board) pawnAttackBitboard(colorToMove uint) uint64 {
	var pawnAttacks uint64

	pawnIndexOffset := 8

	if colorToMove != piece.ColorWhite {
		pawnIndexOffset = -8
	}

	pieceBitboard := b.PieceBitboard(colorToMove | piece.TypePawn)

	for pieceBitboard != 0 {
		startIndex := bits.TrailingZeros64(pieceBitboard)

		side1 := startIndex + pawnIndexOffset - 1
		side2 := startIndex + pawnIndexOffset + 1

		// Add move if it is in the same target row and targets are not empty
		if boardhelper.IsValidDiagonalMove(startIndex, side1) {
			pawnAttacks |= 1 << side1
		}
		if boardhelper.IsValidDiagonalMove(startIndex, side2) {
			pawnAttacks |= 1 << side2
		}

		pieceBitboard &= pieceBitboard - 1
	}

	return pawnAttacks
}

func (b *Board) straightSlidingAttackBitboard(colorToMove uint) uint64 {
	straightIndexOffsets := []int{1, -1, 8, -8}
	var straightSlidingAttacks uint64

	pieceBitboard := b.PieceBitboard(colorToMove|piece.TypeRook) | b.PieceBitboard(colorToMove|piece.TypeQueen) | b.PieceBitboard(colorToMove|piece.TypeKing)
	ownPiecesBitboard := b.Pieces[(colorToMove>>3)-1]
	enemyPiecesBitboard := b.Pieces[1-((colorToMove>>3)-1)]

	enemyColor := piece.ColorBlack
	if colorToMove != piece.ColorWhite {
		enemyColor = piece.ColorWhite
	}
	enemyKingIndex := bits.TrailingZeros64(b.PieceBitboard(enemyColor | piece.TypeKing))

	for pieceBitboard != 0 {
		startIndex := bits.TrailingZeros64(pieceBitboard)

		maxLength := 8
		if (b.PieceAtIndex(startIndex) & 0b00111) == piece.TypeKing {
			maxLength = 1
		}

		for _, offset := range straightIndexOffsets {
			targetIndex := startIndex + offset

			length := 0
			// Go deep until we hit maxLength or crossed borders
			for length < maxLength && boardhelper.IsValidStraightMove(startIndex, targetIndex) {

				straightSlidingAttacks |= 1 << targetIndex

				// If we hit our own piece, we break
				if boardhelper.IsIndexBitSet(targetIndex, ownPiecesBitboard) {
					break
				}

				// If we hit an enemy piece except the enemy king, we break
				if boardhelper.IsIndexBitSet(targetIndex, enemyPiecesBitboard) && (targetIndex != enemyKingIndex) {
					break
				}

				targetIndex += offset
				length++
			}

		}

		pieceBitboard &= pieceBitboard - 1
	}

	return straightSlidingAttacks
}

func (b *Board) diagonalSlidingAttackBitboard(colorToMove uint) uint64 {
	diagonalIndexOffsets := []int{-7, 7, -9, 9}
	var diagonalSlidingAttacks uint64

	pieceBitboard := b.PieceBitboard(colorToMove|piece.TypeBishop) | b.PieceBitboard(colorToMove|piece.TypeQueen) | b.PieceBitboard(colorToMove|piece.TypeKing)
	ownPiecesBitboard := b.Pieces[(colorToMove>>3)-1]
	enemyPiecesBitboard := b.Pieces[1-((colorToMove>>3)-1)]

	enemyColor := piece.ColorBlack
	if colorToMove != piece.ColorWhite {
		enemyColor = piece.ColorWhite
	}
	enemyKingIndex := bits.TrailingZeros64(b.PieceBitboard(enemyColor | piece.TypeKing))

	for pieceBitboard != 0 {
		startIndex := bits.TrailingZeros64(pieceBitboard)

		maxLength := 8
		if (b.PieceAtIndex(startIndex) & 0b00111) == piece.TypeKing {
			maxLength = 1
		}

		for _, offset := range diagonalIndexOffsets {
			targetIndex := startIndex + offset

			length := 0
			// Go deep until we hit maxlength or crossed border
			for length < maxLength && boardhelper.IsValidDiagonalMove(startIndex, targetIndex) {

				diagonalSlidingAttacks |= 1 << targetIndex

				// If we hit our own piece, we break
				if boardhelper.IsIndexBitSet(targetIndex, ownPiecesBitboard) {
					break
				}

				// If we hit an enemy piece except the enemy king, we break
				if boardhelper.IsIndexBitSet(targetIndex, enemyPiecesBitboard) && (targetIndex != enemyKingIndex) {
					break
				}

				targetIndex += offset
				length++
			}

		}

		pieceBitboard &= pieceBitboard - 1
	}

	return diagonalSlidingAttacks
}

func (b *Board) knightAttackBitboard(colorToMove uint) uint64 {
	knightIndexOffsets := []int{-6, 6, -10, 10, -15, 15, -17, 17}
	var knightAttacks uint64

	pieceBitboard := b.PieceBitboard(colorToMove | piece.TypeKnight)

	for pieceBitboard != 0 {
		startIndex := bits.TrailingZeros64(pieceBitboard)

		for _, offset := range knightIndexOffsets {
			if boardhelper.IsValidKnightMove(startIndex, startIndex+offset) {
				knightAttacks |= 1 << (startIndex + offset)
			}
		}

		pieceBitboard &= pieceBitboard - 1
	}

	return knightAttacks
}

func (b *Board) straightPinnedPiecesBitboard(colorToMove uint) uint64 {
	straightIndexOffsets := []int{-1, 1, -8, 8}
	var pinnedPieces uint64

	kingIndex := bits.TrailingZeros64(b.PieceBitboard(colorToMove | piece.TypeKing))
	ownPiecesBitboard := b.Pieces[(colorToMove>>3)-1]
	enemyPiecesBitboard := b.Pieces[1-((colorToMove>>3)-1)]

	enemyColor := piece.ColorWhite
	if colorToMove == piece.ColorWhite {
		enemyColor = piece.ColorBlack
	}

	attackingEnemyPiecesBitboard := b.PieceBitboard(enemyColor|piece.TypeRook) | b.PieceBitboard(enemyColor|piece.TypeQueen)
	otherEnemyPiecesBitboard := enemyPiecesBitboard & ^attackingEnemyPiecesBitboard

	for _, offset := range straightIndexOffsets {

		rayIndex := kingIndex + offset
		ownPieceIndex := -1

		// Go as long as the ray moves to valid fields
		depth := 1
		for depth < 8 && boardhelper.IsValidStraightMove(kingIndex, rayIndex) {

			// Hit our own piece
			if boardhelper.IsIndexBitSet(rayIndex, ownPiecesBitboard) {

				// Hit own piece 2 times in a row
				if ownPieceIndex != -1 {
					break
				}

				// Hit own piece first time
				ownPieceIndex = rayIndex
			}

			// Hit enemy attacking piece
			if boardhelper.IsIndexBitSet(rayIndex, attackingEnemyPiecesBitboard) {

				// If we have hit our own piece before, that piece is pinned
				if ownPieceIndex != -1 {
					pinnedPieces |= 1 << ownPieceIndex
					break
				}

				// Fun fact: we are in check if the code comes to this comment
			}

			// Hit any other enemy piece
			if boardhelper.IsIndexBitSet(rayIndex, otherEnemyPiecesBitboard) {
				break // Can just exit, no matter if we hit our own piece first, there is an enemy piece in the way of possible attackers
			}

			rayIndex += offset
			depth++
		}
	}

	return pinnedPieces
}

func (b *Board) diagonalPinnedPiecesBitboard(colorToMove uint) uint64 {
	diagonalIndexOffsets := []int{-7, 7, -9, 9}
	var pinnedPieces uint64

	kingIndex := bits.TrailingZeros64(b.PieceBitboard(colorToMove | piece.TypeKing))
	ownPiecesBitboard := b.Pieces[(colorToMove>>3)-1]
	enemyPiecesBitboard := b.Pieces[1-((colorToMove>>3)-1)]

	enemyColor := piece.ColorWhite
	if colorToMove == piece.ColorWhite {
		enemyColor = piece.ColorBlack
	}

	attackingEnemyPiecesBitboard := b.PieceBitboard(enemyColor|piece.TypeBishop) | b.PieceBitboard(enemyColor|piece.TypeQueen)
	otherEnemyPiecesBitboard := enemyPiecesBitboard & ^attackingEnemyPiecesBitboard

	for _, offset := range diagonalIndexOffsets {
		rayIndex := kingIndex + offset
		ownPieceIndex := -1

		// Go as long as the ray moves to valid fields
		for boardhelper.IsValidDiagonalMove(kingIndex, rayIndex) {

			// Hit our own piece
			if boardhelper.IsIndexBitSet(rayIndex, ownPiecesBitboard) {

				// Hit own piece 2 times in a row
				if ownPieceIndex != -1 {
					break
				}

				// Hit own piece first time
				ownPieceIndex = rayIndex
			}

			// Hit enemy attacking piece
			if boardhelper.IsIndexBitSet(rayIndex, attackingEnemyPiecesBitboard) {

				// If we have hit our own piece before, that piece is pinned
				if ownPieceIndex != -1 {
					pinnedPieces |= 1 << ownPieceIndex
					break
				}

				// Fun fact: we are in check if the code comes to this comment
			}

			// Hit any other enemy piece
			if boardhelper.IsIndexBitSet(rayIndex, otherEnemyPiecesBitboard) {
				break // Can just exit, no matter if we hit our own piece first, there is an enemy piece in the way of possible attackers
			}

			rayIndex += offset
		}
	}

	return pinnedPieces
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
