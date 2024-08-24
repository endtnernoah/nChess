package game

import (
	"endtner.dev/nChess/game/board"
	"endtner.dev/nChess/game/boardhelper"
	"endtner.dev/nChess/game/formatter"
	"endtner.dev/nChess/game/move"
	"endtner.dev/nChess/game/piece"
	"fmt"
	"math/bits"
	"strconv"
	"strings"
)

type Game struct {
	b *board.Board

	// From fen
	whiteToMove           bool
	castlingAvailability  uint // Bits set like KQkq
	enPassantTargetSquare int
	halfMoves             int
	moveCount             int
}

func New(fenString string) *Game {
	g := Game{}

	// Setting up game from fen
	fenFields := strings.Split(fenString, " ")

	// Setting up board
	g.b = board.New(fenFields[0])

	// Checking who is to move
	g.whiteToMove = fenFields[1] == "w"

	// Castling availability
	castlingAvailabilityFlags := fenFields[2]
	if strings.Contains(castlingAvailabilityFlags, "K") {
		g.castlingAvailability |= 0b1000
	}
	if strings.Contains(castlingAvailabilityFlags, "Q") {
		g.castlingAvailability |= 0b0100
	}
	if strings.Contains(castlingAvailabilityFlags, "k") {
		g.castlingAvailability |= 0b0010
	}
	if strings.Contains(castlingAvailabilityFlags, "q") {
		g.castlingAvailability |= 0b0001
	}

	// EP Target Square
	if fenFields[3] != "-" {
		g.enPassantTargetSquare = boardhelper.SquareToIndex(fenFields[3])
	}

	// Half move count
	data, err := strconv.Atoi(fenFields[4])
	if err != nil {
		fmt.Println("Failed parsing halfMove number")
		panic(err)
	}
	g.halfMoves = data

	// Move count
	data, err = strconv.Atoi(fenFields[5])
	if err != nil {
		fmt.Println("Failed parsing moveCount number")
		panic(err)
	}
	g.moveCount = data

	return &g
}

func (g *Game) Board() *board.Board {
	return g.b
}

func (g *Game) DisplayBoard() {
	unicodeBoard := formatter.ToUnicodeBoard(formatter.BitboardMappingAll(g.b))
	fmt.Println(formatter.FormatUnicodeBoard(unicodeBoard))
}

func (g *Game) DisplayBoardPretty() {
	unicodeBoard := formatter.ToUnicodeBoard(formatter.BitboardMappingAll(g.b))
	fmt.Println(formatter.FormatUnicodeBoardWithBorders(unicodeBoard))
}

func (g *Game) GenerateLegalMoves() []move.Move {
	// Index offsets for each move
	pawnIndexOffset := 8
	straightIndexOffsets := []int{-1, 1, -8, 8}
	diagonalIndexOffsets := []int{-7, 7, -9, 9}
	knightIndexOffsets := []int{-6, 6, -10, 10, -15, 15, -17, 17}

	colorToMove := piece.ColorWhite
	ownOccupiedPieces := g.b.WhitePieces()
	ownPinnedPieces := g.b.WhitePinnedPieces
	enemyOccupiedPieces := g.b.BlackPieces()
	enemyAttackFields := g.b.BlackAttackFields
	promotionRank := 7

	if !g.whiteToMove {
		pawnIndexOffset = -8 // Changes based on walking direction

		colorToMove = piece.ColorBlack
		ownOccupiedPieces = g.b.BlackPieces()
		ownPinnedPieces = g.b.BlackPinnedPieces
		enemyOccupiedPieces = g.b.WhitePieces()
		enemyAttackFields = g.b.WhiteAttackFields
		promotionRank = 0
	}

	pawnBitboard := g.b.PieceBitboard(colorToMove | piece.TypePawn)
	rookBitboard := g.b.PieceBitboard(colorToMove | piece.TypeRook)
	knightBitboard := g.b.PieceBitboard(colorToMove | piece.TypeKnight)
	bishopBitboard := g.b.PieceBitboard(colorToMove | piece.TypeBishop)
	queenBitboard := g.b.PieceBitboard(colorToMove | piece.TypeQueen)
	kingBitboard := g.b.PieceBitboard(colorToMove | piece.TypeKing)

	// Starting move generation
	legalMoves := make([]move.Move, 0)

	// Generate pawn moves
	for pawnBitboard != 0 {
		// Get index of LSB
		startIndex := bits.TrailingZeros64(pawnBitboard)
		targetIndex := startIndex + pawnIndexOffset

		// Continue if piece is pinned
		if boardhelper.IsIndexBitSet(startIndex, ownPinnedPieces) {
			pawnBitboard &= pawnBitboard - 1
			continue
		}

		// Continue if target index is out of bounds, just go to the next iteration
		if targetIndex < 0 || targetIndex > 63 {
			pawnBitboard &= pawnBitboard - 1
			continue
		}

		// Add move if target square is empty
		if !boardhelper.IsIndexBitSet(targetIndex, ownOccupiedPieces|enemyOccupiedPieces) {
			legalMoves = append(legalMoves, move.New(startIndex, targetIndex, -1, -1, -1, targetIndex/8 == promotionRank))
		}

		// Check if the pawn is on starting square
		isStartingSquare := startIndex >= 8 && startIndex < 16
		if !g.whiteToMove {
			isStartingSquare = startIndex >= 48 && startIndex < 56
		}

		// Can move 2 rows from starting square
		if isStartingSquare {
			if !boardhelper.IsIndexBitSet(targetIndex+pawnIndexOffset, ownOccupiedPieces|enemyOccupiedPieces) {
				legalMoves = append(legalMoves, move.New(startIndex, targetIndex+pawnIndexOffset, -1, targetIndex, -1, false))
			}
		}

		// Check for possible captures
		side1 := targetIndex - 1
		side2 := targetIndex + 1

		// Add move if it is in the same target row and targets are not empty
		if side1/8 == targetIndex/8 && boardhelper.IsIndexBitSet(side1, enemyOccupiedPieces) {
			legalMoves = append(legalMoves, move.New(startIndex, side1, -1, -1, -1, targetIndex/8 == promotionRank))
		}
		if side2/8 == targetIndex/8 && boardhelper.IsIndexBitSet(side2, enemyOccupiedPieces) {
			legalMoves = append(legalMoves, move.New(startIndex, side2, -1, -1, -1, targetIndex/8 == promotionRank))
		}

		// Add move if either side can capture en passant
		if side1 == g.enPassantTargetSquare && (side1/8 == g.enPassantTargetSquare/8) {
			legalMoves = append(legalMoves, move.New(startIndex, side1, side1-pawnIndexOffset, -1, -1, false))
		}
		if side2 == g.enPassantTargetSquare && (side2/8 == g.enPassantTargetSquare/8) {
			legalMoves = append(legalMoves, move.New(startIndex, side2, side2-pawnIndexOffset, -1, -1, false))
		}

		// Remove LSB of bitboard
		pawnBitboard &= pawnBitboard - 1
	}

	// Generate rook moves
	for rookBitboard != 0 {
		startIndex := bits.TrailingZeros64(rookBitboard)

		if boardhelper.IsIndexBitSet(startIndex, ownPinnedPieces) {
			rookBitboard &= rookBitboard - 1
			continue
		}

		// Go as deep as possible
		for _, offset := range straightIndexOffsets {
			targetIndex := startIndex + offset

			// Go deep into direction
			for boardhelper.IsValidStraightMove(startIndex, targetIndex) && !boardhelper.IsIndexBitSet(targetIndex, ownOccupiedPieces) {

				// Break on possible capture for direction
				if boardhelper.IsIndexBitSet(targetIndex, enemyOccupiedPieces) {
					legalMoves = append(legalMoves, move.New(startIndex, targetIndex, -1, -1, -1, false))
					break
				}

				legalMoves = append(legalMoves, move.New(startIndex, targetIndex, -1, -1, -1, false))
				targetIndex += offset
			}

		}

		// Remove LSB
		rookBitboard &= rookBitboard - 1
	}

	for knightBitboard != 0 {
		// Get index of LSB
		startIndex := bits.TrailingZeros64(knightBitboard)

		if boardhelper.IsIndexBitSet(startIndex, ownPinnedPieces) {
			knightBitboard &= knightBitboard - 1
			continue
		}

		for _, offset := range knightIndexOffsets {
			targetIndex := startIndex + offset

			if !boardhelper.IsValidKnightMove(startIndex, targetIndex) || boardhelper.IsIndexBitSet(targetIndex, ownOccupiedPieces) {
				continue
			}

			legalMoves = append(legalMoves, move.New(startIndex, targetIndex, -1, -1, -1, false))
		}

		knightBitboard &= knightBitboard - 1
	}

	// Generate bishop moves
	for bishopBitboard != 0 {
		startIndex := bits.TrailingZeros64(bishopBitboard)

		if boardhelper.IsIndexBitSet(startIndex, ownPinnedPieces) {
			bishopBitboard &= bishopBitboard - 1
		}

		// Go as deep as possible
		for _, offset := range diagonalIndexOffsets {
			targetIndex := startIndex + offset

			for boardhelper.IsValidDiagonalMove(startIndex, targetIndex) && !boardhelper.IsIndexBitSet(targetIndex, ownOccupiedPieces) {
				if boardhelper.IsIndexBitSet(targetIndex, enemyOccupiedPieces) {
					legalMoves = append(legalMoves, move.New(startIndex, targetIndex, -1, -1, -1, false))
					break
				}

				legalMoves = append(legalMoves, move.New(startIndex, targetIndex, -1, -1, -1, false))
				targetIndex += offset
			}

		}

		bishopBitboard &= bishopBitboard - 1
	}

	for queenBitboard != 0 {
		// Get index of LSB
		startIndex := bits.TrailingZeros64(queenBitboard)

		if boardhelper.IsIndexBitSet(startIndex, ownPinnedPieces) {
			queenBitboard &= queenBitboard - 1
			continue
		}

		// Go as deep as possible
		for _, offset := range straightIndexOffsets {
			targetIndex := startIndex + offset

			for boardhelper.IsValidStraightMove(startIndex, targetIndex) && !boardhelper.IsIndexBitSet(targetIndex, ownOccupiedPieces) {
				if boardhelper.IsIndexBitSet(targetIndex, enemyOccupiedPieces) {
					legalMoves = append(legalMoves, move.New(startIndex, targetIndex, -1, -1, -1, false))
					break
				}

				legalMoves = append(legalMoves, move.New(startIndex, targetIndex, -1, -1, -1, false))
				targetIndex += offset
			}

		}

		for _, offset := range diagonalIndexOffsets {
			targetIndex := startIndex + offset

			for boardhelper.IsValidDiagonalMove(startIndex, targetIndex) && !boardhelper.IsIndexBitSet(targetIndex, ownOccupiedPieces) {
				if boardhelper.IsIndexBitSet(targetIndex, enemyOccupiedPieces) {
					legalMoves = append(legalMoves, move.New(startIndex, targetIndex, -1, -1, -1, false))
					break
				}

				legalMoves = append(legalMoves, move.New(startIndex, targetIndex, -1, -1, -1, false))
				targetIndex += offset
			}

		}

		queenBitboard &= queenBitboard - 1
	}

	for kingBitboard != 0 {
		// Get index of LSB
		startIndex := bits.TrailingZeros64(kingBitboard)

		for _, offset := range straightIndexOffsets {
			targetIndex := startIndex + offset

			if !boardhelper.IsValidStraightMove(startIndex, targetIndex) || boardhelper.IsIndexBitSet(targetIndex, ownOccupiedPieces) {
				continue
			}

			if boardhelper.IsIndexBitSet(targetIndex, enemyAttackFields) {
				continue
			}

			legalMoves = append(legalMoves, move.New(startIndex, targetIndex, -1, -1, -1, false))
		}

		for _, offset := range diagonalIndexOffsets {
			targetIndex := startIndex + offset

			if !boardhelper.IsValidDiagonalMove(startIndex, targetIndex) || boardhelper.IsIndexBitSet(targetIndex, ownOccupiedPieces) {
				continue
			}

			// Cannot move king into attack field
			if boardhelper.IsIndexBitSet(targetIndex, enemyAttackFields) {
				continue
			}

			legalMoves = append(legalMoves, move.New(startIndex, targetIndex, -1, -1, -1, false))
		}

		// Castling
		kingSideAllowed := g.castlingAvailability&0b1000 != 0
		queenSideAllowed := g.castlingAvailability&0b0100 != 0

		var emptyFieldsKingSide uint64 = 0b0000011
		var emptyFieldsQueenSide uint64 = 0b1110

		kingSideRook := 7
		queenSideRook := 0

		if !g.whiteToMove {
			kingSideAllowed = g.castlingAvailability&0b0010 != 0
			queenSideAllowed = g.castlingAvailability&0b0001 != 0

			emptyFieldsKingSide = emptyFieldsKingSide << 56
			emptyFieldsQueenSide = emptyFieldsQueenSide << 56

			kingSideRook = 63
			queenSideRook = 56
		}

		// White king to g1, rook to f1; black king to g8, rook to f8
		if kingSideAllowed && !boardhelper.IsIndexBitSet(startIndex, enemyAttackFields) {
			if (emptyFieldsKingSide&g.b.Occupied() == 0) &&
				!boardhelper.IsIndexBitSet(startIndex+1, enemyAttackFields) &&
				!boardhelper.IsIndexBitSet(startIndex+2, enemyAttackFields) {

				legalMoves = append(legalMoves, move.New(startIndex, startIndex+2, -1, -1, kingSideRook, false))
			}
		}

		// White king to c1, rook to d1; black king to c8, rook to d8
		if queenSideAllowed && !boardhelper.IsIndexBitSet(startIndex, enemyAttackFields) {
			if (emptyFieldsQueenSide&g.b.Occupied() == 0) &&
				!boardhelper.IsIndexBitSet(startIndex-1, enemyAttackFields) &&
				!boardhelper.IsIndexBitSet(startIndex-2, enemyAttackFields) {

				legalMoves = append(legalMoves, move.New(startIndex, startIndex-2, -1, -1, queenSideRook, false))
			}
		}

		kingBitboard &= kingBitboard - 1
	}

	return legalMoves
}
