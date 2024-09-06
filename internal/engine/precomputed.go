package engine

import (
	"endtner.dev/nChess/internal/board"
)

var DirectionalOffsets = []int{8, -8, -1, 1, 7, -7, 9, -9}
var KnightOffsets = []int{-6, 6, -10, 10, -15, 15, -17, 17}

var DistanceToEdge = func() [][]int {
	distances := make([][]int, 64)

	for file := 0; file < 8; file++ {
		for rank := 0; rank < 8; rank++ {
			numNorth := 7 - rank
			numSouth := rank
			numWest := file
			numEast := 7 - file

			squareIndex := rank*8 + file

			distances[squareIndex] = []int{
				numNorth,
				numSouth,
				numWest,
				numEast,
				Min(numNorth, numWest),
				Min(numSouth, numEast),
				Min(numNorth, numEast),
				Min(numSouth, numWest),
			}
		}
	}

	return distances
}()

var ComputedKnightMoves = func() []uint64 {
	moves := make([]uint64, 64)

	for startIndex := range 64 {
		for _, offset := range KnightOffsets {
			targetIndex := startIndex + offset
			if board.IsValidKnightMove(startIndex, targetIndex) {
				moves[startIndex] |= 1 << targetIndex
			}
		}
	}

	return moves
}()

var ComputedKingMoves = func() []uint64 {
	moves := make([]uint64, 64)

	for startIndex := range 64 {
		for i, offset := range DirectionalOffsets {
			if DistanceToEdge[startIndex][i] != 0 {
				moves[startIndex] |= 1 << (startIndex + offset)
			}
		}
	}

	return moves
}()

var ComputedPawnMoves = func() [][]uint64 {
	moves := make([][]uint64, 2)
	moves[0] = make([]uint64, 64)
	moves[1] = make([]uint64, 64)

	for startIndex := range 64 {
		targetIndexWhite := startIndex + 8
		targetIndexBlack := startIndex - 8

		if board.IsValidStraightMove(startIndex, targetIndexWhite) {
			moves[0][startIndex] |= 1 << targetIndexWhite
		}
		if board.IsValidStraightMove(startIndex, targetIndexBlack) {
			moves[1][startIndex] |= 1 << targetIndexBlack
		}
	}

	return moves
}()

var ComputedPawnAttacks = func() [][]uint64 {
	moves := make([][]uint64, 2)
	moves[0] = make([]uint64, 64)
	moves[1] = make([]uint64, 64)

	for startIndex := range 64 {
		targetIndexWhite := startIndex + 8
		targetIndexBlack := startIndex - 8

		if board.IsValidDiagonalMove(startIndex, targetIndexWhite+1) {
			moves[0][startIndex] |= 1 << (targetIndexWhite + 1)
		}
		if board.IsValidDiagonalMove(startIndex, targetIndexWhite-1) {
			moves[0][startIndex] |= 1 << (targetIndexWhite - 1)
		}

		if board.IsValidDiagonalMove(startIndex, targetIndexBlack+1) {
			moves[1][startIndex] |= 1 << (targetIndexBlack + 1)
		}
		if board.IsValidDiagonalMove(startIndex, targetIndexBlack-1) {
			moves[1][startIndex] |= 1 << (targetIndexBlack - 1)
		}
	}

	return moves
}()

var AlignmentMask = func() [64][64]uint64 {
	alignMask := [64][64]uint64{}

	for squareA := 0; squareA < 64; squareA++ {
		fileA, rankA := squareA%8, squareA/8

		for squareB := 0; squareB < 64; squareB++ {
			fileB, rankB := squareB%8, squareB/8

			deltaFile := fileB - fileA
			deltaRank := rankB - rankA

			dirFile := sign(deltaFile)
			dirRank := sign(deltaRank)

			for i := -7; i <= 7; i++ {
				file := fileA + dirFile*i
				rank := rankA + dirRank*i

				if file >= 0 && file < 8 && rank >= 0 && rank < 8 {
					square := rank*8 + file
					alignMask[squareA][squareB] |= 1 << uint(square)
				}
			}
		}
	}

	return alignMask
}()

func Min(x int, y int) int {
	if x < y {
		return x
	}
	return y
}

func sign(x int) int {
	if x < 0 {
		return -1
	}
	if x > 0 {
		return 1
	}
	return 0
}
