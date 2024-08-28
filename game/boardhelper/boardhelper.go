package boardhelper

import (
	"unicode"
)

func IsIndexBitSet(bitIndex int, bitboard uint64) bool {
	return (bitboard & (1 << bitIndex)) != 0
}

func SquareToIndex(square string) int {
	if len(square) != 2 {
		return 0
	}

	file := int(unicode.ToLower(rune(square[0]))) - int('a')
	rank := int(square[1]) - int('1')

	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return 0
	}

	return rank*8 + file
}

func IndexToSquare(index int) string {
	if index < 0 || index > 63 {
		return ""
	}

	file := index % 8
	rank := index / 8

	return string(rune('a'+file)) + string(rune('1'+rank))
}

func IsValidStraightMove(startIndex int, targetIndex int) bool {
	// Check if move is vertical
	isVertical := startIndex%8 == targetIndex%8

	// Check if move is horizontal
	sourceRank := startIndex / 8
	targetRank := targetIndex / 8
	isHorizontal := sourceRank == targetRank

	// The move must be either vertical or horizontal, but not both
	isStraight := isVertical || isHorizontal

	// Check bounds
	isWithinBounds := targetIndex >= 0 && targetIndex <= 63

	return isStraight && isWithinBounds
}

func IsValidDiagonalMove(startIndex, targetIndex int) bool {
	abs := func(n int) int {
		if n < 0 {
			return -n
		}
		return n
	}

	// Check if the move is within bounds
	if targetIndex < 0 || targetIndex > 63 {
		return false
	}

	// Calculate the difference between source and target
	diff := abs(targetIndex - startIndex)

	// Calculate ranks and files for source and target
	sourceRank, sourceFile := startIndex/8, startIndex%8
	targetRank, targetFile := targetIndex/8, targetIndex%8

	// Check if the move is diagonal
	rankDiff := abs(targetRank - sourceRank)
	fileDiff := abs(targetFile - sourceFile)

	// A move is diagonal if the change in rank equals the change in file
	isDiagonal := rankDiff == fileDiff

	// Check if the move doesn't wrap around the board
	isNoWrap := diff%7 == 0 || diff%9 == 0

	return isDiagonal && isNoWrap
}

func IsValidKnightMove(sourceIndex, targetIndex int) bool {
	abs := func(n int) int {
		if n < 0 {
			return -n
		}
		return n
	}

	// Check if the move is within bounds
	if targetIndex < 0 || targetIndex > 63 {
		return false
	}

	// Calculate ranks and files for source and target
	sourceRank, sourceFile := sourceIndex/8, sourceIndex%8
	targetRank, targetFile := targetIndex/8, targetIndex%8

	// Calculate the differences in rank and file
	rankDiff := abs(targetRank - sourceRank)
	fileDiff := abs(targetFile - sourceFile)

	// A knight move is valid if:
	// 1. It moves 2 squares in one direction and 1 square in the perpendicular direction
	// 2. The total of rank difference and file difference is 3
	return (rankDiff == 2 && fileDiff == 1) || (rankDiff == 1 && fileDiff == 2)
}

func CalculateRayOffset(fromIndex, toIndex int) int {
	diff := toIndex - fromIndex

	// Handle diagonal pins
	if diff%7 == 0 {
		if diff > 0 {
			return 7 // Pinned piece is on the bottom-left to top-right diagonal
		}
		return -7 // Pinned piece is on the top-left to bottom-right diagonal
	}

	if diff%9 == 0 {
		if diff > 0 {
			return 9 // Pinned piece is on the top-left to bottom-right diagonal
		}
		return -9 // Pinned piece is on the bottom-left to top-right diagonal
	}

	// Handle horizontal pins
	if diff%8 == 0 {
		if diff > 0 {
			return 8 // Pinned piece is below the king
		}
		return -8 // Pinned piece is above the king
	}

	// Handle vertical pins
	if diff > -8 && diff < 8 {
		if diff > 0 {
			return 1 // Pinned piece is to the right of the king
		}
		return -1 // Pinned piece is to the left of the king
	}

	// If we get here, the pieces aren't on a straight line
	return 0
}
