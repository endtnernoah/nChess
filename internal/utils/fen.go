package utils

import (
	"endtner.dev/nChess/internal/board"
	"fmt"
	"math/bits"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

func FromFen(fenString string) *board.Position {
	p := board.Position{}

	// Setting up game from fen
	fenFields := strings.Split(fenString, " ")

	// Setting up old
	figurePositionRows := strings.Split(fenFields[0], "/")
	slices.Reverse(figurePositionRows)
	figurePositions := strings.Join(figurePositionRows, "/")

	p.Bitboards = make([]uint64, 0b1111)
	p.Pieces = make([]uint8, 64)

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
		pc := board.Value(rune(currentChar))

		p.Pieces[boardPosition] = pc
		p.Bitboards[pc] |= 1 << boardPosition

		boardPosition++
	}

	// Checking who is to move
	p.WhiteToMove = fenFields[1] == "w"

	p.FriendlyColor = board.White
	p.OpponentColor = board.Black

	p.PawnOffset = 8
	p.PromotionRank = 7

	if !p.WhiteToMove {
		p.FriendlyColor, p.OpponentColor = p.OpponentColor, p.FriendlyColor
		p.PawnOffset = -8
		p.PromotionRank = 0
	}

	p.FriendlyIndex = int(p.FriendlyColor >> 3)
	p.OpponentIndex = 1 - p.FriendlyIndex

	p.FriendlyKingIndex = bits.TrailingZeros64(p.Bitboards[p.FriendlyColor|board.King])
	p.OpponentKingIndex = bits.TrailingZeros64(p.Bitboards[p.OpponentColor|board.King])

	// Castling availability
	castlingAvailabilityFlags := fenFields[2]
	if strings.Contains(castlingAvailabilityFlags, "K") {
		p.CastlingRights |= 0b1000
	}
	if strings.Contains(castlingAvailabilityFlags, "Q") {
		p.CastlingRights |= 0b0100
	}
	if strings.Contains(castlingAvailabilityFlags, "k") {
		p.CastlingRights |= 0b0010
	}
	if strings.Contains(castlingAvailabilityFlags, "q") {
		p.CastlingRights |= 0b0001
	}

	// EP Target Square
	if fenFields[3] != "-" {
		p.EnPassantSquare = board.SquareToIndex(fenFields[3])
	} else {
		p.EnPassantSquare = -1
	}

	// Half move count
	data, err := strconv.Atoi(fenFields[4])
	if err != nil {
		fmt.Println("Failed parsing halfMove number")
		panic(err)
	}
	p.HalfMoves = data

	// Move count
	data, err = strconv.Atoi(fenFields[5])
	if err != nil {
		fmt.Println("Failed parsing FullMoves number")
		panic(err)
	}
	p.FullMoves = data

	p.Zobrist = board.GetZobrist(&p)

	return &p
}

func ToFEN(p *board.Position) string {
	var fen strings.Builder

	emptySquares := 0

	// Piece placement
	for rank := 7; rank >= 0; rank-- {
		for file := 0; file < 8; file++ {
			index := rank*8 + file

			pieceValue := p.Pieces[index]

			if pieceValue == 0 {
				emptySquares++
			} else {
				if emptySquares > 0 {
					fen.WriteString(strconv.Itoa(emptySquares))
					emptySquares = 0
				}
				fen.WriteString(board.ToString(pieceValue))
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
	if p.WhiteToMove {
		fen.WriteString("w")
	} else {
		fen.WriteString("p")
	}

	// Castling availability
	fen.WriteString(" ")
	castlingRights := ""
	if p.CastlingRights&(1<<3) != 0 {
		castlingRights += "K"
	}
	if p.CastlingRights&(1<<2) != 0 {
		castlingRights += "Q"
	}
	if p.CastlingRights&(1<<1) != 0 {
		castlingRights += "k"
	}
	if p.CastlingRights&1 != 0 {
		castlingRights += "q"
	}
	if castlingRights == "" {
		fen.WriteString("-")
	} else {
		fen.WriteString(castlingRights)
	}

	// En passant target square
	fen.WriteString(" ")
	if p.EnPassantSquare == -1 {
		fen.WriteString("-")
	} else {
		fen.WriteString(board.IndexToSquare(p.EnPassantSquare))
	}

	// Half-Move clock
	fen.WriteString(" ")
	fen.WriteString(strconv.Itoa(p.HalfMoves))

	// Full-Move number
	fen.WriteString(" ")
	fen.WriteString(strconv.Itoa(p.FullMoves))

	return fen.String()
}
