package piece

var White uint8 = 0b01000
var Black uint8 = 0b10000

var Pawn uint8 = 0b00001
var Rook uint8 = 0b00010
var Knight uint8 = 0b00011
var Bishop uint8 = 0b00100
var Queen uint8 = 0b00101
var King uint8 = 0b00110

func Value(pieceChar rune) uint8 {
	switch pieceChar {
	case 'r':
		return Black | Rook
	case 'n':
		return Black | Knight
	case 'b':
		return Black | Bishop
	case 'q':
		return Black | Queen
	case 'k':
		return Black | King
	case 'p':
		return Black | Pawn

	case 'R':
		return White | Rook
	case 'N':
		return White | Knight
	case 'B':
		return White | Bishop
	case 'Q':
		return White | Queen
	case 'K':
		return White | King
	case 'P':
		return White | Pawn
	default:
		return 0
	}
}

func ToString(pieceValue uint8) string {
	pieceType := pieceValue & 0b00111
	color := pieceValue & 0b11000

	var char rune
	switch pieceType {
	case Pawn:
		char = 'p'
	case Rook:
		char = 'r'
	case Knight:
		char = 'n'
	case Bishop:
		char = 'b'
	case Queen:
		char = 'q'
	case King:
		char = 'k'
	default:
		return ""
	}

	if color == White {
		return string(char - 32) // Convert to uppercase
	}
	return string(char)
}
