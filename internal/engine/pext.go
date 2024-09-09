package engine

import (
	"endtner.dev/nChess/internal/board"
)

// PextEntry represents an entry in the Pext lookup table
type PextEntry struct {
	Mask   uint64
	Offset int
}

var RookMasks [64]uint64
var BishopMasks [64]uint64
var RookPextTable [64]PextEntry
var BishopPextTable [64]PextEntry
var PRookMoveTable []uint64
var PBishopMoveTable []uint64

func init() {
	for square := 0; square < 64; square++ {
		RookMasks[square] = Mask(square, board.Rook)
		BishopMasks[square] = Mask(square, board.Bishop)
	}

	totalSizeRooks := 0
	totalSizeBishops := 0

	for square := 0; square < 64; square++ {
		rookMask := RookMasks[square]
		bishopMask := BishopMasks[square]

		rookOccupancies := GenerateOccupancies(rookMask)
		bishopOccupancies := GenerateOccupancies(bishopMask)

		maxIndexRook := 0
		maxIndexBishop := 0

		for _, occ := range rookOccupancies {
			index := int(Pext(occ, rookMask))
			if index > maxIndexRook {
				maxIndexRook = index
			}
		}

		for _, occ := range bishopOccupancies {
			index := int(Pext(occ, bishopMask))
			if index > maxIndexBishop {
				maxIndexBishop = index
			}
		}

		RookPextTable[square] = PextEntry{Mask: rookMask, Offset: totalSizeRooks}
		BishopPextTable[square] = PextEntry{Mask: bishopMask, Offset: totalSizeBishops}

		totalSizeRooks += maxIndexRook + 1
		totalSizeBishops += maxIndexBishop + 1
	}

	// Now create the optimized tables
	PRookMoveTable = make([]uint64, totalSizeRooks)
	PBishopMoveTable = make([]uint64, totalSizeBishops)

	// Fill the tables
	for square := 0; square < 64; square++ {
		rookEntry := RookPextTable[square]
		bishopEntry := BishopPextTable[square]

		rookOccupancies := GenerateOccupancies(rookEntry.Mask)
		bishopOccupancies := GenerateOccupancies(bishopEntry.Mask)

		for _, occ := range rookOccupancies {
			moves := ValidMoves(square, board.Rook, occ)
			index := Pext(occ, rookEntry.Mask)
			PRookMoveTable[rookEntry.Offset+int(index)] = moves
		}

		for _, occ := range bishopOccupancies {
			moves := ValidMoves(square, board.Bishop, occ)
			index := Pext(occ, bishopEntry.Mask)
			PBishopMoveTable[bishopEntry.Offset+int(index)] = moves
		}
	}
}

//go:noescape
func Pext(x, mask uint64) uint64

func PGetRookMoves(square int, occupation uint64) uint64 {
	entry := RookPextTable[square]
	index := Pext(occupation, entry.Mask)
	return PRookMoveTable[entry.Offset+int(index)]
}

func PGetBishopMoves(square int, occupation uint64) uint64 {
	entry := BishopPextTable[square]
	index := Pext(occupation, entry.Mask)
	return PBishopMoveTable[entry.Offset+int(index)]
}

func PGetSliderMoves(square int, occupation uint64, isRook bool) uint64 {
	if isRook {
		return PGetRookMoves(square, occupation)
	}
	return PGetBishopMoves(square, occupation)
}
