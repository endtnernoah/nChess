package game

import (
	"endtner.dev/nChess/game/board"
	"endtner.dev/nChess/game/boardhelper"
	"endtner.dev/nChess/game/formatter"
	"endtner.dev/nChess/game/move"
	"endtner.dev/nChess/game/movegenerator"
	"endtner.dev/nChess/game/piece"
	"fmt"
	"math/bits"
	"strconv"
	"strings"
	"sync"
)

type State struct {
	castlingAvailability  uint
	enPassantTargetSquare int
	halfMoves             int
	moveCount             int
}

type Game struct {
	b *board.Board

	// From fen
	whiteToMove           bool
	castlingAvailability  uint // Bits set like KQkq
	enPassantTargetSquare int
	halfMoves             int
	moveCount             int

	stateStack []State
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
	} else {
		g.enPassantTargetSquare = -1
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

func (g *Game) ToFEN() string {
	var fen strings.Builder

	fen.WriteString(g.Board().ToFEN())

	// Active color
	fen.WriteString(" ")
	if g.whiteToMove {
		fen.WriteString("w")
	} else {
		fen.WriteString("b")
	}

	// Castling availability
	fen.WriteString(" ")
	castlingRights := ""
	if g.castlingAvailability&(1<<3) != 0 {
		castlingRights += "K"
	}
	if g.castlingAvailability&(1<<2) != 0 {
		castlingRights += "Q"
	}
	if g.castlingAvailability&(1<<1) != 0 {
		castlingRights += "k"
	}
	if g.castlingAvailability&1 != 0 {
		castlingRights += "q"
	}
	if castlingRights == "" {
		fen.WriteString("-")
	} else {
		fen.WriteString(castlingRights)
	}

	// En passant target square
	fen.WriteString(" ")
	if g.enPassantTargetSquare == -1 {
		fen.WriteString("-")
	} else {
		fen.WriteString(boardhelper.IndexToSquare(g.enPassantTargetSquare))
	}

	// Halfmove clock
	fen.WriteString(" ")
	fen.WriteString(strconv.Itoa(g.halfMoves))

	// Fullmove number
	fen.WriteString(" ")
	fen.WriteString(strconv.Itoa(g.moveCount))

	return fen.String()
}

func (g *Game) MakeMove(m move.Move) {
	// Update the stack
	g.stateStack = append(g.stateStack, State{g.castlingAvailability, g.enPassantTargetSquare, g.halfMoves, g.moveCount})

	// Set new castling availability
	kingSideRookStart := 7
	queenSideRookStart := 0
	kingStart := 4
	kingSideBitIndex := 3
	queenSideBitIndex := 2

	if !g.whiteToMove {
		kingSideRookStart += 56
		queenSideRookStart += 56
		kingStart += 56
		kingSideBitIndex = 1
		queenSideBitIndex = 0
	}

	if m.StartIndex == kingSideRookStart {
		g.castlingAvailability = g.castlingAvailability & ^(1 << kingSideBitIndex)
	}
	if m.StartIndex == queenSideRookStart {
		g.castlingAvailability = g.castlingAvailability & ^(1 << queenSideBitIndex)
	}
	if m.StartIndex == kingStart {
		g.castlingAvailability = g.castlingAvailability & ^(1 << kingSideBitIndex)
		g.castlingAvailability = g.castlingAvailability & ^(1 << queenSideBitIndex)
	}

	// Setting EP Target Square
	if m.EnPassantPassedSquare != -1 {
		g.enPassantTargetSquare = m.EnPassantPassedSquare
	} else {
		g.enPassantTargetSquare = -1
	}

	// Increase half move if not a pawn move and not a capture
	movedPieceType := g.b.PieceAtIndex(m.StartIndex) & 0b00111
	targetPiece := g.b.PieceAtIndex(m.TargetIndex)
	if movedPieceType == piece.TypePawn || targetPiece != 0 {
		g.halfMoves = 0
	} else {
		g.halfMoves += 1
	}

	// Increase the move number on blacks turns
	if !g.whiteToMove {
		g.moveCount += 1
	}

	// Make the move on board
	g.Board().MakeMove(m)

	// Switch around the color
	g.OtherColorToMove()
}

func (g *Game) UnmakeMove() {

	latestGameState := g.stateStack[len(g.stateStack)-1]

	g.castlingAvailability = latestGameState.castlingAvailability
	g.enPassantTargetSquare = latestGameState.enPassantTargetSquare
	g.halfMoves = latestGameState.halfMoves
	g.moveCount = latestGameState.moveCount

	g.stateStack = g.stateStack[:len(g.stateStack)-1]

	// Unmake move on board
	g.Board().UnmakeMove()

	g.OtherColorToMove()
}

func (g *Game) OtherColorToMove() { g.whiteToMove = !g.whiteToMove }

func (g *Game) WhiteToMove() bool { return g.whiteToMove }

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

func (g *Game) GeneratePseudoLegalMoves() []move.Move {
	/*
		Generating all pseudo-legal moves in parallel and then joining them from the channel to a list
	*/

	pseudoLegalMovesChan := make(chan []move.Move)
	var waitGroup sync.WaitGroup
	waitGroup.Add(5)

	colorToMove := piece.ColorWhite

	if !g.whiteToMove {
		colorToMove = piece.ColorBlack
	}

	// Pawn moves
	go func(b *board.Board, colorToMove uint, enPassantTargetSquare int) {
		defer waitGroup.Done()
		pseudoLegalMovesChan <- movegenerator.GeneratePawnMoves(b, colorToMove, enPassantTargetSquare)
	}(g.b, colorToMove, g.enPassantTargetSquare)

	// Straight sliding moves
	go func(b *board.Board, colorToMove uint) {
		defer waitGroup.Done()
		pseudoLegalMovesChan <- movegenerator.GenerateStraightSlidingMoves(b, colorToMove)
	}(g.b, colorToMove)

	// Diagonal sliding moves
	go func(b *board.Board, colorToMove uint) {
		defer waitGroup.Done()
		pseudoLegalMovesChan <- movegenerator.GenerateDiagonalSlidingMoves(b, colorToMove)
	}(g.b, colorToMove)

	// Knight moves
	go func(b *board.Board, colorToMove uint) {
		defer waitGroup.Done()
		pseudoLegalMovesChan <- movegenerator.GenerateKnightMoves(b, colorToMove)
	}(g.b, colorToMove)

	// King moves
	go func(b *board.Board, colorToMove uint, castlingAvailability uint) {
		defer waitGroup.Done()
		pseudoLegalMovesChan <- movegenerator.GenerateKingMoves(b, colorToMove, castlingAvailability)
	}(g.b, colorToMove, g.castlingAvailability)

	// Wait for all generators to finish
	go func() {
		waitGroup.Wait()
		close(pseudoLegalMovesChan)
	}()

	// Joining to a list, returning
	var pseudoLegalMoves []move.Move
	for m := range pseudoLegalMovesChan {
		pseudoLegalMoves = append(pseudoLegalMoves, m...)
	}

	return pseudoLegalMoves
}

func (g *Game) GenerateLegalMoves() []move.Move {
	/*
		Filtering out all illegal moves
	*/

	var legalMoves []move.Move

	pseudoLegalMoves := g.GeneratePseudoLegalMoves()

	colorToMove := piece.ColorWhite

	if !g.whiteToMove {
		colorToMove = piece.ColorBlack
	}

	ownKingBitboard := g.b.PieceBitboard(colorToMove | piece.TypeKing)
	ownPinnedPieces := g.b.PinnedPieces[(colorToMove>>3)-1]

	ownKingIndex := bits.TrailingZeros64(ownKingBitboard)

	enemyAttackFields := g.b.AttackFields[1-((colorToMove>>3)-1)]

	checkCount := 0
	possibleProtectMoves := ^uint64(0)

	if boardhelper.IsIndexBitSet(ownKingIndex, enemyAttackFields) {
		checkCount, possibleProtectMoves = g.b.CalculateProtectMoves(colorToMove)
	}

	for _, m := range pseudoLegalMoves {

		// Only move along pin ray if piece is pinned
		if boardhelper.IsIndexBitSet(m.StartIndex, ownPinnedPieces) && !g.b.IsPinnedMoveAlongRay(colorToMove, m) {
			continue
		}

		// If the king is in check
		if checkCount > 0 {

			// King is in single check
			if checkCount == 1 {

				// If a move is not to any of the proctect square OR not a king move
				if !boardhelper.IsIndexBitSet(m.TargetIndex, possibleProtectMoves) && (ownKingIndex != m.StartIndex) {

					// Move can not enPassantCapture the checking pawn, not allowed
					if m.EnPassantCaptureSquare == -1 {
						continue
					}

					// Move CAN capture enPassant, but not the attacking pawn, not allowed
					if m.EnPassantCaptureSquare != -1 && !boardhelper.IsIndexBitSet(m.EnPassantCaptureSquare, possibleProtectMoves) {
						continue
					}
				}
			}

			// If the king is in double (or higher) check, only allow king moves
			if checkCount > 1 && ownKingIndex != m.StartIndex {
				continue
			}
		}

		// If we do an enPassant Capture, make sure it does not leave our own king in check
		if m.EnPassantCaptureSquare != -1 && g.b.IsEnPassantMovePinned(colorToMove, m) {
			continue
		}

		legalMoves = append(legalMoves, m)
	}

	return legalMoves
}

func (g *Game) Perft(ply int) int64 {
	/*
		Perft Testing Utility
	*/

	if ply == 0 {
		return 1
	}

	legalMoves := g.GenerateLegalMoves()
	var totalNodes int64 = 0

	// Not the official implementation, but works a lot faster
	if ply == 1 {
		return int64(len(legalMoves))
	}

	for _, m := range legalMoves {
		g.MakeMove(m)

		subNodes := g.Perft(ply - 1)
		totalNodes += subNodes

		if ply == 10 {
			fmt.Printf("%s: %d\n", move.PrintSimple(m), subNodes)
		}

		g.UnmakeMove()
	}

	return totalNodes
}
