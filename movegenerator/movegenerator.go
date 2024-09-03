package movegenerator

import (
	"endtner.dev/nChess/board"
	"endtner.dev/nChess/board/boardhelper"
	"endtner.dev/nChess/board/move"
	"endtner.dev/nChess/board/piece"
	"fmt"
	"math/bits"
	"os"
	"time"
)

/*
	MoveGenerator is now responsible for generating all Bitboards. THIS CODE DOES NOT RUN IN PARALLEL!

	When parallelized, other generators will override these variables for a given generator, so everything crashes.
	I tried instancing the generator per class, that created so much overhead that the performance increase was mitigated.
*/

/*
	Generators
*/

var TotalTimePrecompute time.Duration
var TotalTimeKingGeneration time.Duration
var TotalTimePawnGeneration time.Duration
var TotalTimeSlidingGeneration time.Duration
var TotalTimeKnightGeneration time.Duration

func LegalMoves(b *board.Board) []move.Move {
	pseudoLegalMoves := make([]move.Move, 218) // Maximum possible moves in a chess position is 218
	index := 0

	startPrecompute := time.Now()

	friendlyColor := b.FriendlyColor
	opponentColor := b.OpponentColor

	friendlyIndex := b.FriendlyIndex
	opponentIndex := b.OpponentIndex

	friendlyPieces := b.Bitboards[friendlyColor|piece.Pawn] | b.Bitboards[friendlyColor|piece.Knight] | b.Bitboards[friendlyColor|piece.Rook] | b.Bitboards[friendlyColor|piece.Bishop] | b.Bitboards[friendlyColor|piece.Queen] | b.Bitboards[friendlyColor|piece.King]
	opponentPieces := b.Bitboards[opponentColor|piece.Pawn] | b.Bitboards[opponentColor|piece.Knight] | b.Bitboards[opponentColor|piece.Rook] | b.Bitboards[opponentColor|piece.Bishop] | b.Bitboards[opponentColor|piece.Queen] | b.Bitboards[opponentColor|piece.King]
	allPieces := friendlyPieces | opponentPieces

	friendlyPawns := b.Bitboards[friendlyColor|piece.Pawn]
	friendlyKnights := b.Bitboards[friendlyColor|piece.Knight]
	friendlyRooks := b.Bitboards[friendlyColor|piece.Rook]
	friendlyOrthogonalSliders := friendlyRooks | b.Bitboards[friendlyColor|piece.Queen]
	friendlyDiagonalSliders := b.Bitboards[friendlyColor|piece.Bishop] | b.Bitboards[friendlyColor|piece.Queen]
	friendlyKingBitboard := b.Bitboards[friendlyColor|piece.King]
	friendlyKingIndex := b.FriendlyKingIndex

	if friendlyKingIndex == 64 {
		fmt.Println(b.ToFEN())
		fmt.Println(b.MoveHistory)
		os.Exit(0)
	}

	opponentOrthogonalSliders := b.Bitboards[opponentColor|piece.Rook] | b.Bitboards[opponentColor|piece.Queen]
	opponentDiagonalSliders := b.Bitboards[opponentColor|piece.Bishop] | b.Bitboards[opponentColor|piece.Queen]

	opponentSlidingAttacks := uint64(0)
	var UpdateSlideAttacks = func(pieces uint64, orthogonal bool) {
		blockers := allPieces & ^(1 << friendlyKingIndex)

		for pieces != 0 {
			opponentSlidingAttacks |= GetSliderMoves(bits.TrailingZeros64(pieces), blockers, orthogonal)
			pieces &= pieces - 1
		}
	}
	UpdateSlideAttacks(opponentOrthogonalSliders, true)
	UpdateSlideAttacks(opponentDiagonalSliders, false)
	opponentAttacks := opponentSlidingAttacks

	inCheck := false
	inDoubleCheck := false

	friendlyPinRays := uint64(0)
	validMoveMask := uint64(0)

	var ComputeAttackData = func() {

		// TODO: Check if there are no queens, if yes check if there are no rooks/bishops to loop over less directions

		for offsetIndex, offset := range DirectionalOffsets {
			isDiagonal := offsetIndex > 3

			sliders := opponentOrthogonalSliders
			if isDiagonal {
				sliders = opponentDiagonalSliders
			}

			// TODO: Can skip offset if there are no sliders in that direction

			rayMask := uint64(0)
			isFriendlyPieceAlongRay := false

			for distance := range DistanceToEdge[friendlyKingIndex][offsetIndex] {
				squareIndex := friendlyKingIndex + (offset * (distance + 1))
				p := b.Pieces[squareIndex]

				rayMask |= 1 << squareIndex

				if p == 0 {
					continue
				}

				if (p & 0b11000) == friendlyColor {
					// Break if it is the second friendly piece we encounter
					if !isFriendlyPieceAlongRay {
						isFriendlyPieceAlongRay = true
					} else {
						break
					}

				} else {
					// Check if piece is one of the current sliders
					if (sliders & (1 << squareIndex)) != 0 {
						if isFriendlyPieceAlongRay {
							friendlyPinRays |= rayMask // There is a friendly blocking piece, so it is a pin
						} else {
							validMoveMask |= rayMask // There is no friendly blocking piece, so it is a check
							inDoubleCheck = inCheck
							inCheck = true
						}
					}
					break // We either discovered a pin or a check OR the enemy piece at that index is blocking any attacks
				}
			}

			// Only king can move if we are in double check
			if inDoubleCheck {
				break
			}
		}

		opponentKnights := b.Bitboards[opponentColor|piece.Knight]
		for opponentKnights != 0 {
			knightIndex := bits.TrailingZeros64(opponentKnights)
			knightAttacks := ComputedKnightMoves[knightIndex]

			if (knightAttacks & friendlyKingBitboard) != 0 {
				inDoubleCheck = inCheck
				inCheck = true
				validMoveMask |= 1 << knightIndex
			}

			opponentAttacks |= knightAttacks
			opponentKnights &= opponentKnights - 1
		}

		opponentPawns := b.Bitboards[opponentColor|piece.Pawn]
		for opponentPawns != 0 {
			pawnIndex := bits.TrailingZeros64(opponentPawns)
			pawnAttacks := ComputedPawnAttacks[opponentIndex][pawnIndex]

			if (pawnAttacks & friendlyKingBitboard) != 0 {
				inDoubleCheck = inCheck
				inCheck = true
				validMoveMask |= 1 << pawnIndex
			}

			opponentAttacks |= pawnAttacks
			opponentPawns &= opponentPawns - 1
		}

		opponentAttacks |= ComputedKingMoves[bits.TrailingZeros64(b.Bitboards[opponentColor|piece.King])]

		if !inCheck {
			validMoveMask = ^uint64(0)
		}
	}
	ComputeAttackData()

	/*
		Validation functions
	*/

	var IsEnPassantMovePinned = func(startSquare, targetSquare, epCaptureSquare int) bool {
		epBlockers := uint64((1 << startSquare) | (1 << targetSquare) | (1 << epCaptureSquare))
		maskedBlockers := (allPieces | epBlockers) & ^(allPieces & epBlockers)

		if opponentOrthogonalSliders != 0 {
			rookAttacks := GetRookMoves(friendlyKingIndex, maskedBlockers)
			return (rookAttacks & opponentOrthogonalSliders) != 0
		}

		if opponentDiagonalSliders != 0 {
			bishopAttacks := GetBishopMoves(friendlyKingIndex, maskedBlockers)
			return (bishopAttacks & opponentDiagonalSliders) != 0
		}

		return false
	}

	/*
		Generator functions
	*/
	var KingMoves = func() {
		kingMoveMask := ComputedKingMoves[friendlyKingIndex] & ^friendlyPieces & ^opponentAttacks
		for kingMoveMask != 0 {
			pseudoLegalMoves[index] = move.New(friendlyKingIndex, bits.TrailingZeros64(kingMoveMask))
			index++

			kingMoveMask &= kingMoveMask - 1
		}

		initialKingIndex := 4
		if friendlyColor != piece.White {
			initialKingIndex = 60
		}

		// King is not on its original square, will not be allowed to castle
		if inCheck || friendlyKingIndex != initialKingIndex {
			return
		}

		kingSideAllowed := b.CastlingAvailability&0b1000 != 0
		queenSideAllowed := b.CastlingAvailability&0b0100 != 0

		var kingSideEmptyMask uint64 = 0b1100000
		var queenSideEmptyMask uint64 = 0b1110

		kingSideRookIndex := 7
		queenSideRookIndex := 0

		var kingSideAttackMask uint64 = 1<<friendlyKingIndex | 1<<(friendlyKingIndex+1) | 1<<(friendlyKingIndex+2)
		var queenSideAttackMask uint64 = 1<<friendlyKingIndex | 1<<(friendlyKingIndex-1) | 1<<(friendlyKingIndex-2)

		if !b.WhiteToMove {
			kingSideAllowed = b.CastlingAvailability&0b0010 != 0
			queenSideAllowed = b.CastlingAvailability&0b0001 != 0

			kingSideEmptyMask <<= 56
			queenSideEmptyMask <<= 56

			kingSideRookIndex += 56
			queenSideRookIndex += 56
		}

		if kingSideAllowed &&
			(kingSideAttackMask&opponentAttacks) == 0 && // King does not start or pass through attacked field
			(kingSideEmptyMask&allPieces) == 0 && // All fields are empty
			boardhelper.IsIndexBitSet(kingSideRookIndex, friendlyRooks) { // There is a rook on its field

			pseudoLegalMoves[index] = move.New(friendlyKingIndex, friendlyKingIndex+2, move.WithRookStartingSquare(kingSideRookIndex))
			index++
		}

		if queenSideAllowed &&
			(queenSideAttackMask&opponentAttacks) == 0 &&
			(queenSideEmptyMask&allPieces) == 0 &&
			boardhelper.IsIndexBitSet(queenSideRookIndex, friendlyRooks) {

			pseudoLegalMoves[index] = move.New(friendlyKingIndex, friendlyKingIndex-2, move.WithRookStartingSquare(queenSideRookIndex))
			index++
		}
	}

	var PawnMoves = func() {
		for friendlyPawns != 0 {
			// Get index of LSB
			pieceIndex := bits.TrailingZeros64(friendlyPawns)
			targetIndex := pieceIndex + b.PawnOffset

			// Continue if target index is out of bounds, just go to the next iteration
			if targetIndex < 0 || targetIndex > 63 {
				friendlyPawns &= friendlyPawns - 1
				continue
			}

			// Check if the pawn is on starting square
			isBaseRank := pieceIndex/8 == 1
			isPromotionRank := pieceIndex/8 == 6

			if !b.WhiteToMove {
				isBaseRank, isPromotionRank = isPromotionRank, isBaseRank
			}

			// Handling normal moves
			validMoves := ComputedPawnMoves[friendlyIndex][pieceIndex] & ^allPieces

			// Adding double pawn pushes
			if isBaseRank && validMoves != 0 {
				validMoves |= ComputedPawnMoves[friendlyIndex][pieceIndex+b.PawnOffset] & ^allPieces
			}

			// Generate all attacks
			attackMask := ComputedPawnAttacks[b.FriendlyIndex][pieceIndex]
			validAttacks := attackMask & opponentPieces

			// Generate en passant attacks
			validPawnMoveMask := validMoveMask
			if b.EnPassantTargetSquare != -1 {
				if 1<<(b.EnPassantTargetSquare-b.PawnOffset)&validMoveMask != 0 {
					validPawnMoveMask |= 1 << b.EnPassantTargetSquare
				}
				validAttacks |= attackMask & (1 << b.EnPassantTargetSquare)
			}

			validMoves |= validAttacks

			// Handle pins
			if (friendlyPinRays & (1 << pieceIndex)) != 0 {
				validMoves &= AlignmentMask[pieceIndex][friendlyKingIndex]
			}

			// Filter out non-protecting moves
			validMoves &= validPawnMoveMask

			// Writing moves
			for validMoves != 0 {
				moveTargetIndex := bits.TrailingZeros64(validMoves)

				if isPromotionRank {
					pseudoLegalMoves[index] = move.New(pieceIndex, moveTargetIndex, move.WithPromotion(b.FriendlyColor|piece.Knight))
					index++
					pseudoLegalMoves[index] = move.New(pieceIndex, moveTargetIndex, move.WithPromotion(b.FriendlyColor|piece.Bishop))
					index++
					pseudoLegalMoves[index] = move.New(pieceIndex, moveTargetIndex, move.WithPromotion(b.FriendlyColor|piece.Queen))
					index++
					pseudoLegalMoves[index] = move.New(pieceIndex, moveTargetIndex, move.WithPromotion(b.FriendlyColor|piece.Rook))
					index++
				} else {
					if moveTargetIndex == b.EnPassantTargetSquare {
						if !IsEnPassantMovePinned(pieceIndex, moveTargetIndex, moveTargetIndex-b.PawnOffset) {
							pseudoLegalMoves[index] = move.New(pieceIndex, moveTargetIndex, move.WithEnPassantCaptureSquare(moveTargetIndex-b.PawnOffset))
							index++
						}
					} else if moveTargetIndex == targetIndex+b.PawnOffset {
						pseudoLegalMoves[index] = move.New(pieceIndex, moveTargetIndex, move.WithEnPassantPassedSquare(moveTargetIndex-b.PawnOffset))
						index++
					} else {
						pseudoLegalMoves[index] = move.New(pieceIndex, moveTargetIndex)
						index++
					}
				}

				validMoves &= validMoves - 1
			}

			// Remove LSB of bitboard
			friendlyPawns &= friendlyPawns - 1
		}
	}

	var SlidingMoves = func() {
		for friendlyOrthogonalSliders|friendlyDiagonalSliders != 0 {
			pieceIndex := bits.TrailingZeros64(friendlyOrthogonalSliders | friendlyDiagonalSliders)

			if friendlyOrthogonalSliders&(1<<pieceIndex) != 0 {
				sliderMoves := GetRookMoves(pieceIndex, allPieces) & ^friendlyPieces & validMoveMask

				if (friendlyPinRays & (1 << pieceIndex)) != 0 {
					sliderMoves &= AlignmentMask[pieceIndex][friendlyKingIndex]
				}

				for sliderMoves != 0 {
					targetIndex := bits.TrailingZeros64(sliderMoves)

					pseudoLegalMoves[index] = move.New(pieceIndex, targetIndex)
					index++

					sliderMoves &= sliderMoves - 1
				}

				friendlyOrthogonalSliders &= friendlyOrthogonalSliders - 1
			}
			if friendlyDiagonalSliders&(1<<pieceIndex) != 0 {
				sliderMoves := GetBishopMoves(pieceIndex, allPieces) & ^friendlyPieces & validMoveMask

				if (friendlyPinRays & (1 << pieceIndex)) != 0 {
					sliderMoves &= AlignmentMask[pieceIndex][friendlyKingIndex]
				}

				for sliderMoves != 0 {
					targetIndex := bits.TrailingZeros64(sliderMoves)

					pseudoLegalMoves[index] = move.New(pieceIndex, targetIndex)
					index++

					sliderMoves &= sliderMoves - 1
				}

				friendlyDiagonalSliders &= friendlyDiagonalSliders - 1
			}
		}
	}

	var KnightMoves = func() {
		for friendlyKnights != 0 {
			pieceIndex := bits.TrailingZeros64(friendlyKnights)

			validMoves := ComputedKnightMoves[pieceIndex] & ^friendlyPieces & validMoveMask

			for validMoves != 0 {
				pseudoLegalMoves[index] = move.New(pieceIndex, bits.TrailingZeros64(validMoves))
				index++

				validMoves &= validMoves - 1
			}

			friendlyKnights &= friendlyKnights - 1
		}
	}

	/*
		Creating pseudo-legal moves, then filtering out illegal moves
	*/
	TotalTimePrecompute += time.Since(startPrecompute)

	startKingMoves := time.Now()
	KingMoves()
	TotalTimeKingGeneration += time.Since(startKingMoves)

	// Can directly return king moves if we are in multi check
	if inDoubleCheck {
		return pseudoLegalMoves[:index]
	}

	if inCheck {
		friendlyOrthogonalSliders &= ^friendlyPinRays
		friendlyDiagonalSliders &= ^friendlyPinRays
	}

	startPawnMoves := time.Now()
	PawnMoves()
	TotalTimePawnGeneration += time.Since(startPawnMoves)

	//fmt.Println(formatter.UnicodeBoardWithBorders(formatter.ToUnicodeBoard(map[uint64]string{friendlyPinRays: "P"})))
	startSlidingMoves := time.Now()
	SlidingMoves()
	TotalTimeSlidingGeneration += time.Since(startSlidingMoves)

	startKnightMoves := time.Now()
	friendlyKnights &= ^friendlyPinRays // Knights can never move if pinned
	KnightMoves()
	TotalTimeKnightGeneration += time.Since(startKnightMoves)

	return pseudoLegalMoves[:index]
}

/*
	Utility
*/

type TTablePerft map[uint64]int64

var tt = make([]TTablePerft, 10)
var isInitialized = func() bool {
	for i := range len(tt) {
		tt[i] = make(TTablePerft)
	}

	return true
}()

func Perft(b *board.Board, ply int, maxPly int) int64 {
	/*
		Perft Testing Utility
	*/

	if ply == 0 {
		return 1
	}

	legalMoves := LegalMoves(b)
	var totalNodes int64 = 0

	if nodes, found := tt[ply][b.Zobrist]; found {
		return nodes
	}

	// Not the official implementation, but works a lot faster
	if ply == 1 {
		return int64(len(legalMoves))
	}

	for _, m := range legalMoves {
		b.MakeMove(m)

		subNodes := Perft(b, ply-1, maxPly)
		totalNodes += subNodes

		if ply == maxPly {
			fmt.Printf("%s: %d\n", move.PrintSimple(m), subNodes)
		}

		b.UnmakeMove()
	}

	tt[ply][b.Zobrist] = totalNodes
	return totalNodes
}
