package piece

var ColorWhite uint = 0b01000
var ColorBlack uint = 0b10000

var TypePawn uint = 0b00001
var TypeRook uint = 0b00010
var TypeKnight uint = 0b00011
var TypeBishop uint = 0b00100
var TypeQueen uint = 0b00101
var TypeKing uint = 0b00110

func PieceValue(pieceChar rune) uint {
	switch pieceChar {
	case 'r':
		return ColorBlack | TypeRook
	case 'n':
		return ColorBlack | TypeKnight
	case 'b':
		return ColorBlack | TypeBishop
	case 'q':
		return ColorBlack | TypeQueen
	case 'k':
		return ColorBlack | TypeKing
	case 'p':
		return ColorBlack | TypePawn

	case 'R':
		return ColorWhite | TypeRook
	case 'N':
		return ColorWhite | TypeKnight
	case 'B':
		return ColorWhite | TypeBishop
	case 'Q':
		return ColorWhite | TypeQueen
	case 'K':
		return ColorWhite | TypeKing
	case 'P':
		return ColorWhite | TypePawn
	default:
		return 0
	}
}
