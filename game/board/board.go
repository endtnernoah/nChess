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

	WhiteAttackFields uint64
	BlackAttackFields uint64

	WhitePinnedPieces uint64
	BlackPinnedPieces uint64

	whitePieces uint64
	blackPieces uint64
	occupied    uint64

	bitboardStack [][]uint64
}

func New(fenString string) *Board {
	// Since fen starts at a8, we want to split it at the /, reverse the list and join it back with /
	fenRows := strings.Split(fenString, "/")
	slices.Reverse(fenRows)
	fenString = strings.Join(fenRows, "/")

	b := Board{}
	b.bitboards = make([]uint64, 0b11111)

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

		// We set that to a 0 so
		if rune(currentChar) == '/' {
			continue
		}

		// Matching character, setting bits
		b.bitboards[piece.Value(rune(currentChar))] |= 1 << boardPosition

		boardPosition++
	}

	// Update shit like the whitePieces, whiteAttacks, pinned Pieces...
	b.Update()

	return &b
}

func (b *Board) MakeMove(move move.Move) {
	// Put current bitboards on the stack
	copiedBitboards := make([]uint64, len(b.bitboards))
	copy(copiedBitboards, b.bitboards)
	b.bitboardStack = append(b.bitboardStack, copiedBitboards)

	// Get moved piece
	movedPiece := b.GetPieceAtIndex(move.StartIndex)

	// Remove piece from source square
	b.setPieceBitboard(movedPiece, b.PieceBitboard(movedPiece) & ^(1<<move.StartIndex))

	// Possibly remove captured piece
	capturedPiece := b.GetPieceAtIndex(move.TargetIndex)
	if capturedPiece != 0 && ((capturedPiece&0b11000)&(movedPiece&0b11000)) == 0 {
		b.setPieceBitboard(capturedPiece, b.PieceBitboard(capturedPiece) & ^(1<<move.TargetIndex))
	}

	// Add new piece on the target square
	if move.IsPromotion {
		newQueen := piece.TypeQueen | (movedPiece & 0b11000)
		b.setPieceBitboard(newQueen, b.PieceBitboard(newQueen)|(1<<move.TargetIndex))
	} else {
		b.setPieceBitboard(movedPiece, b.PieceBitboard(movedPiece)|(1<<move.TargetIndex))
	}

	// Update all bitboards
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

func (b *Board) GetPieceAtIndex(index int) uint {
	if index < 0 || index > 63 {
		return 0 // Invalid index
	}

	for pieceType := uint(0); pieceType < 6; pieceType++ {
		if boardhelper.IsIndexBitSet(index, b.PieceBitboard(pieceType|piece.ColorWhite)) {
			return pieceType | piece.ColorWhite
		}
		if boardhelper.IsIndexBitSet(index, b.PieceBitboard(pieceType|piece.ColorBlack)) {
			return pieceType | piece.ColorBlack
		}
	}

	return 0 // No piece found
}

func (b *Board) WhitePieces() uint64 {
	return b.whitePieces
}

func (b *Board) BlackPieces() uint64 {
	return b.blackPieces
}

func (b *Board) Occupied() uint64 {
	return b.occupied
}

func (b *Board) Update() {
	b.UpdateBitboards()
	b.UpdateWhiteAttackBitboard()
	b.UpdateBlackAttackBitboard()
	b.UpdateWhitePinnedPieces()
	b.UpdateBlackPinnedPieces()
}

func (b *Board) UpdateWhiteAttackBitboard() {
	whitePawnAttacks := pawnAttackBitboard(b.PieceBitboard(piece.ColorWhite|piece.TypePawn), b.whitePieces, 8)
	whiteStraightSlidingAttacks := straightSlidingAttackBitboard(
		b.PieceBitboard(piece.ColorWhite|piece.TypeRook)|b.PieceBitboard(piece.ColorWhite|piece.TypeQueen),
		b.whitePieces,
		b.blackPieces,
		8)
	whiteStraightKingAttacks := straightSlidingAttackBitboard(
		b.PieceBitboard(piece.ColorWhite|piece.TypeKing),
		b.whitePieces,
		b.blackPieces,
		1)
	whiteDiagonalSlidingAttacks := diagonalSlidingAttackBitboard(
		b.PieceBitboard(piece.ColorWhite|piece.TypeBishop)|b.PieceBitboard(piece.ColorWhite|piece.TypeQueen),
		b.whitePieces,
		b.blackPieces,
		8)
	whiteDiagonalKingAttacks := diagonalSlidingAttackBitboard(
		b.PieceBitboard(piece.ColorWhite|piece.TypeKing),
		b.whitePieces,
		b.blackPieces,
		1)
	whiteKnightAttacks := knightAttackBitboard(b.PieceBitboard(piece.ColorWhite|piece.TypeKnight), b.whitePieces)

	b.WhiteAttackFields = whitePawnAttacks |
		whiteStraightSlidingAttacks |
		whiteStraightKingAttacks |
		whiteDiagonalSlidingAttacks |
		whiteDiagonalKingAttacks |
		whiteKnightAttacks
}

func (b *Board) UpdateBlackAttackBitboard() {
	blackPawnAttacks := pawnAttackBitboard(b.PieceBitboard(piece.ColorBlack|piece.TypePawn), b.blackPieces, -8)
	blackStraightSlidingAttacks := straightSlidingAttackBitboard(
		b.PieceBitboard(piece.ColorBlack|piece.TypeRook)|b.PieceBitboard(piece.ColorWhite|piece.TypeQueen),
		b.blackPieces,
		b.whitePieces,
		8)
	blackStraightKingAttacks := straightSlidingAttackBitboard(
		b.PieceBitboard(piece.ColorBlack|piece.TypeKing),
		b.blackPieces,
		b.whitePieces,
		1)
	blackDiagonalSlidingAttacks := diagonalSlidingAttackBitboard(
		b.PieceBitboard(piece.ColorBlack|piece.TypeBishop)|b.PieceBitboard(piece.ColorWhite|piece.TypeQueen),
		b.blackPieces,
		b.whitePieces,
		8)
	blackDiagonalKingAttacks := diagonalSlidingAttackBitboard(
		b.PieceBitboard(piece.ColorBlack|piece.TypeKing),
		b.blackPieces,
		b.whitePieces,
		1)
	blackKnightAttacks := knightAttackBitboard(b.PieceBitboard(piece.ColorBlack|piece.TypeKnight), b.blackPieces)

	b.BlackAttackFields = blackPawnAttacks |
		blackStraightSlidingAttacks |
		blackStraightKingAttacks |
		blackDiagonalSlidingAttacks |
		blackDiagonalKingAttacks |
		blackKnightAttacks
}

func (b *Board) UpdateWhitePinnedPieces() {
	blackStraightAttackers := b.PieceBitboard(piece.ColorBlack|piece.TypeRook) | b.PieceBitboard(piece.ColorBlack|piece.TypeQueen)
	blackDiagonalAttackers := b.PieceBitboard(piece.ColorBlack|piece.TypeBishop) | b.PieceBitboard(piece.ColorBlack|piece.TypeQueen)

	straightPinnedPieces := straightPinnedPiecesBitboard(
		b.PieceBitboard(piece.ColorWhite|piece.TypeKing),
		b.whitePieces,
		blackStraightAttackers,
		b.blackPieces&^blackStraightAttackers)

	diagonalPinnedPieces := diagonalPinnedPiecesBitboard(
		b.PieceBitboard(piece.ColorWhite|piece.TypeKing),
		b.whitePieces,
		blackDiagonalAttackers,
		b.blackPieces&^blackDiagonalAttackers)

	b.WhitePinnedPieces = straightPinnedPieces | diagonalPinnedPieces
}

func (b *Board) UpdateBlackPinnedPieces() {
	whiteStraightAttackers := b.PieceBitboard(piece.ColorWhite|piece.TypeRook) | b.PieceBitboard(piece.ColorBlack|piece.TypeQueen)
	whiteDiagonalAttackers := b.PieceBitboard(piece.ColorWhite|piece.TypeBishop) | b.PieceBitboard(piece.ColorBlack|piece.TypeQueen)

	straightPinnedPieces := straightPinnedPiecesBitboard(
		b.PieceBitboard(piece.ColorBlack|piece.TypeKing),
		b.blackPieces,
		whiteStraightAttackers,
		b.whitePieces&^whiteStraightAttackers)

	diagonalPinnedPieces := diagonalPinnedPiecesBitboard(
		b.PieceBitboard(piece.ColorBlack|piece.TypeKing),
		b.blackPieces,
		whiteDiagonalAttackers,
		b.whitePieces&^whiteDiagonalAttackers)

	b.BlackPinnedPieces = straightPinnedPieces | diagonalPinnedPieces
}

func (b *Board) UpdateBitboards() {
	// Setting the occupied fields
	for p, bitboard := range b.bitboards {
		if uint(p)&piece.ColorWhite == piece.ColorWhite {
			b.whitePieces |= bitboard
		}
		if uint(p)&piece.ColorBlack == piece.ColorBlack {
			b.blackPieces |= bitboard
		}
	}

	b.occupied = b.whitePieces | b.blackPieces
}

func pawnAttackBitboard(pieceBitboard uint64, ownPiecesBitboard uint64, pawnIndexOffset int) uint64 {
	var pawnAttacks uint64

	for pieceBitboard != 0 {
		startIndex := bits.TrailingZeros64(pieceBitboard)

		side1 := startIndex + pawnIndexOffset - 1
		side2 := startIndex + pawnIndexOffset + 1

		if (side1 < 0 || side1 > 63) || (side2 < 0 || side2 > 63) {
			continue
		}

		// Add move if it is in the same target row and targets are not empty
		if side1/8 == (startIndex+pawnIndexOffset)/8 && !boardhelper.IsIndexBitSet(side1, ownPiecesBitboard) {
			pawnAttacks |= 1 << side1
		}
		if side2/8 == (startIndex+pawnIndexOffset)/8 && !boardhelper.IsIndexBitSet(side2, ownPiecesBitboard) {
			pawnAttacks |= 1 << side2
		}

		pieceBitboard &= pieceBitboard - 1
	}

	return pawnAttacks
}

func straightSlidingAttackBitboard(pieceBitboard uint64, ownPiecesBitboard uint64, enemyPiecesBitboard uint64, maxLength int) uint64 {
	straightIndexOffsets := []int{1, -1, 8, -8}
	var straightSlidingAttacks uint64

	for pieceBitboard != 0 {
		startIndex := bits.TrailingZeros64(pieceBitboard)

		for _, offset := range straightIndexOffsets {
			targetIndex := startIndex + offset

			length := 0
			// Go deep until we hit borders or our own pieces or maxLength
			for length < maxLength && boardhelper.IsValidStraightMove(startIndex, targetIndex) && !boardhelper.IsIndexBitSet(targetIndex, ownPiecesBitboard) {

				// We can go as deep as a capture
				if boardhelper.IsIndexBitSet(targetIndex, enemyPiecesBitboard) {
					straightSlidingAttacks |= 1 << targetIndex
					break
				}

				straightSlidingAttacks |= 1 << targetIndex
				targetIndex += offset
				length++
			}

		}

		pieceBitboard &= pieceBitboard - 1
	}

	return straightSlidingAttacks
}

func diagonalSlidingAttackBitboard(pieceBitboard uint64, ownPiecesBitboard uint64, enemyPiecesBitboard uint64, maxLength int) uint64 {
	diagonalIndexOffsets := []int{-7, 7, -9, 9}
	var diagonalSlidingAttacks uint64

	for pieceBitboard != 0 {
		startIndex := bits.TrailingZeros64(pieceBitboard)

		for _, offset := range diagonalIndexOffsets {
			targetIndex := startIndex + offset

			length := 0
			// Go deep until we hit borders or our own pieces or maxLength
			for length < maxLength && boardhelper.IsValidDiagonalMove(startIndex, targetIndex) && !boardhelper.IsIndexBitSet(targetIndex, ownPiecesBitboard) {

				// We can go as deep as a capture
				if boardhelper.IsIndexBitSet(targetIndex, enemyPiecesBitboard) {
					diagonalSlidingAttacks |= 1 << targetIndex
					break
				}

				diagonalSlidingAttacks |= 1 << targetIndex
				targetIndex += offset
				length++
			}

		}

		pieceBitboard &= pieceBitboard - 1
	}

	return diagonalSlidingAttacks
}

func knightAttackBitboard(pieceBitboard uint64, ownPiecesBitboard uint64) uint64 {
	knightIndexOffsets := []int{-6, 6, -10, 10, -15, 15, -17, 17}
	var knightAttacks uint64

	for pieceBitboard != 0 {
		startIndex := bits.TrailingZeros64(pieceBitboard)

		for _, offset := range knightIndexOffsets {
			if boardhelper.IsValidKnightMove(startIndex, startIndex+offset) && !boardhelper.IsIndexBitSet(startIndex+offset, ownPiecesBitboard) {
				knightAttacks |= 1 << (startIndex + offset)
			}
		}

		pieceBitboard &= pieceBitboard - 1
	}

	return knightAttacks
}

func straightPinnedPiecesBitboard(kingBitboard uint64, ownPiecesBitboard uint64, attackingEnemyPiecesBitboard uint64, otherEnemyPiecesBitboard uint64) uint64 {
	straightIndexOffsets := []int{-1, 1, -8, 8}
	var pinnedPieces uint64

	for kingBitboard != 0 {
		kingIndex := bits.TrailingZeros64(kingBitboard)

		for _, offset := range straightIndexOffsets {
			rayIndex := kingIndex + offset
			ownPieceIndex := -1

			// Go as long as the ray moves to valid fields
			for boardhelper.IsValidStraightMove(kingIndex, rayIndex) {

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
			}
		}

		kingBitboard &= kingBitboard - 1
	}

	return pinnedPieces
}

func diagonalPinnedPiecesBitboard(kingBitboard uint64, ownPiecesBitboard uint64, attackingEnemyPiecesBitboard uint64, otherEnemyPiecesBitboard uint64) uint64 {
	diagonalIndexOffsets := []int{-7, 7, -9, 9}
	var pinnedPieces uint64

	for kingBitboard != 0 {
		kingIndex := bits.TrailingZeros64(kingBitboard)

		for _, offset := range diagonalIndexOffsets {
			rayIndex := kingIndex + offset
			ownPieceIndex := -1

			// Go as long as the ray moves to valid fields
			for boardhelper.IsValidStraightMove(kingIndex, rayIndex) {

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
			}
		}

		kingBitboard &= kingBitboard - 1
	}

	return pinnedPieces
}
