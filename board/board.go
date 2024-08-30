package board

import (
	"endtner.dev/nChess/board/boardhelper"
	"endtner.dev/nChess/board/move"
	"endtner.dev/nChess/board/piece"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

type State struct {
	Bitboards []uint64
	Pieces    []uint8

	CastlingAvailability  uint8
	EnPassantTargetSquare int
	HalfMoves             int
	FullMoves             int
}

type Board struct {
	// Bitboards are stored from bottom left to top right, meaning A1 to H8
	// Bitboards are stored at the index of the piece they have
	Bitboards []uint64
	Pieces    []uint8

	// From fen
	WhiteToMove           bool
	CastlingAvailability  uint8 // Bits set like KQkq
	EnPassantTargetSquare int
	HalfMoves             int
	FullMoves             int

	stateStack []State
}

func New(fenString string) *Board {
	b := Board{}

	// Setting up game from fen
	fenFields := strings.Split(fenString, " ")

	// Setting up old
	figurePositionRows := strings.Split(fenFields[0], "/")
	slices.Reverse(figurePositionRows)
	figurePositions := strings.Join(figurePositionRows, "/")

	b.Bitboards = make([]uint64, 0b10111)
	b.Pieces = make([]uint8, 64)

	// Setting up pieces
	boardPosition := 0
	for i := 0; i < len(figurePositions); i++ {
		currentChar := figurePositions[i]

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

	// Checking who is to move
	b.WhiteToMove = fenFields[1] == "w"

	// Castling availability
	castlingAvailabilityFlags := fenFields[2]
	if strings.Contains(castlingAvailabilityFlags, "K") {
		b.CastlingAvailability |= 0b1000
	}
	if strings.Contains(castlingAvailabilityFlags, "Q") {
		b.CastlingAvailability |= 0b0100
	}
	if strings.Contains(castlingAvailabilityFlags, "k") {
		b.CastlingAvailability |= 0b0010
	}
	if strings.Contains(castlingAvailabilityFlags, "q") {
		b.CastlingAvailability |= 0b0001
	}

	// EP Target Square
	if fenFields[3] != "-" {
		b.EnPassantTargetSquare = boardhelper.SquareToIndex(fenFields[3])
	} else {
		b.EnPassantTargetSquare = -1
	}

	// Half move count
	data, err := strconv.Atoi(fenFields[4])
	if err != nil {
		fmt.Println("Failed parsing halfMove number")
		panic(err)
	}
	b.HalfMoves = data

	// Move count
	data, err = strconv.Atoi(fenFields[5])
	if err != nil {
		fmt.Println("Failed parsing FullMoves number")
		panic(err)
	}
	b.FullMoves = data

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

	// Active color
	fen.WriteString(" ")
	if b.WhiteToMove {
		fen.WriteString("w")
	} else {
		fen.WriteString("b")
	}

	// Castling availability
	fen.WriteString(" ")
	castlingRights := ""
	if b.CastlingAvailability&(1<<3) != 0 {
		castlingRights += "K"
	}
	if b.CastlingAvailability&(1<<2) != 0 {
		castlingRights += "Q"
	}
	if b.CastlingAvailability&(1<<1) != 0 {
		castlingRights += "k"
	}
	if b.CastlingAvailability&1 != 0 {
		castlingRights += "q"
	}
	if castlingRights == "" {
		fen.WriteString("-")
	} else {
		fen.WriteString(castlingRights)
	}

	// En passant target square
	fen.WriteString(" ")
	if b.EnPassantTargetSquare == -1 {
		fen.WriteString("-")
	} else {
		fen.WriteString(boardhelper.IndexToSquare(b.EnPassantTargetSquare))
	}

	// Half-Move clock
	fen.WriteString(" ")
	fen.WriteString(strconv.Itoa(b.HalfMoves))

	// Full-Move number
	fen.WriteString(" ")
	fen.WriteString(strconv.Itoa(b.FullMoves))

	return fen.String()
}

func (b *Board) MakeMove(m move.Move) {
	// Update the stack
	s := State{}

	s.Pieces = make([]uint8, len(b.Pieces))
	s.Bitboards = make([]uint64, len(b.Bitboards))

	copy(s.Pieces, b.Pieces)
	copy(s.Bitboards, b.Bitboards)

	s.CastlingAvailability = b.CastlingAvailability
	s.EnPassantTargetSquare = b.EnPassantTargetSquare
	s.HalfMoves = b.HalfMoves
	s.FullMoves = b.FullMoves

	b.stateStack = append(b.stateStack, s)

	// Set new castling availability
	kingSideRookStart := 7
	queenSideRookStart := 0
	kingStart := 4
	kingSideBitIndex := 3
	queenSideBitIndex := 2

	if !b.WhiteToMove {
		kingSideRookStart += 56
		queenSideRookStart += 56
		kingStart += 56
		kingSideBitIndex = 1
		queenSideBitIndex = 0
	}

	if m.StartIndex == kingSideRookStart {
		b.CastlingAvailability = b.CastlingAvailability & ^(1 << kingSideBitIndex)
	}
	if m.StartIndex == queenSideRookStart {
		b.CastlingAvailability = b.CastlingAvailability & ^(1 << queenSideBitIndex)
	}
	if m.StartIndex == kingStart {
		b.CastlingAvailability = b.CastlingAvailability & ^(1 << kingSideBitIndex)
		b.CastlingAvailability = b.CastlingAvailability & ^(1 << queenSideBitIndex)
	}

	// Setting EP Target Square
	if m.EnPassantPassedSquare != -1 {
		b.EnPassantTargetSquare = m.EnPassantPassedSquare
	} else {
		b.EnPassantTargetSquare = -1
	}

	// Increase half move if not a pawn move and not a capture
	movedPieceType := b.Pieces[m.StartIndex] & 0b00111
	targetPiece := b.Pieces[m.TargetIndex]
	if movedPieceType == piece.Pawn || targetPiece != 0 {
		b.HalfMoves = 0
	} else {
		b.HalfMoves += 1
	}

	// Increase the move number on blacks turns
	if !b.WhiteToMove {
		b.FullMoves += 1
	}

	// Actually move the piece on board
	movedPiece := b.Pieces[m.StartIndex]

	// Handle Castling
	if m.RookStartingSquare != -1 {

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
	} else {
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

	// Switch around the color
	b.OtherColorToMove()
}

func (b *Board) UnmakeMove() {
	if len(b.stateStack) == 0 {
		return
	}

	n := len(b.stateStack)
	lastState := b.stateStack[n-1]

	b.Pieces = lastState.Pieces
	b.Bitboards = lastState.Bitboards
	b.CastlingAvailability = lastState.CastlingAvailability
	b.EnPassantTargetSquare = lastState.EnPassantTargetSquare
	b.HalfMoves = lastState.HalfMoves
	b.FullMoves = lastState.FullMoves

	b.stateStack = b.stateStack[:n-1]

	// Change color to move
	b.OtherColorToMove()
}

func (b *Board) OtherColorToMove() { b.WhiteToMove = !b.WhiteToMove }
