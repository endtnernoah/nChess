package engine

import (
	"endtner.dev/nChess/internal/board"
	"fmt"
	"math/bits"
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

func LegalMoves(p *board.Position) []board.Move {
	pseudoLegalMoves := make([]board.Move, 218) // Maximum possible moves in a chess position is 218
	index := 0

	startPrecompute := time.Now()

	friendlyColor := p.FriendlyColor
	opponentColor := p.OpponentColor

	friendlyIndex := p.FriendlyIndex
	opponentIndex := p.OpponentIndex

	friendlyPieces := p.Bitboards[friendlyColor|board.Pawn] | p.Bitboards[friendlyColor|board.Knight] | p.Bitboards[friendlyColor|board.Rook] | p.Bitboards[friendlyColor|board.Bishop] | p.Bitboards[friendlyColor|board.Queen] | p.Bitboards[friendlyColor|board.King]
	opponentPieces := p.Bitboards[opponentColor|board.Pawn] | p.Bitboards[opponentColor|board.Knight] | p.Bitboards[opponentColor|board.Rook] | p.Bitboards[opponentColor|board.Bishop] | p.Bitboards[opponentColor|board.Queen] | p.Bitboards[opponentColor|board.King]
	allPieces := friendlyPieces | opponentPieces

	friendlyPawns := p.Bitboards[friendlyColor|board.Pawn]
	friendlyKnights := p.Bitboards[friendlyColor|board.Knight]
	friendlyRooks := p.Bitboards[friendlyColor|board.Rook]
	friendlyOrthogonalSliders := friendlyRooks | p.Bitboards[friendlyColor|board.Queen]
	friendlyDiagonalSliders := p.Bitboards[friendlyColor|board.Bishop] | p.Bitboards[friendlyColor|board.Queen]
	friendlyKingBitboard := p.Bitboards[friendlyColor|board.King]
	friendlyKingIndex := p.FriendlyKingIndex

	opponentOrthogonalSliders := p.Bitboards[opponentColor|board.Rook] | p.Bitboards[opponentColor|board.Queen]
	opponentDiagonalSliders := p.Bitboards[opponentColor|board.Bishop] | p.Bitboards[opponentColor|board.Queen]

	opponentSlidingAttacks := uint64(0)
	var UpdateSlideAttacks = func(pieces uint64, orthogonal bool) {
		blockers := allPieces & ^(1 << friendlyKingIndex)

		for pieces != 0 {
			opponentSlidingAttacks |= PGetSliderMoves(bits.TrailingZeros64(pieces), blockers, orthogonal)
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
				p := p.Pieces[squareIndex]

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

		opponentKnights := p.Bitboards[opponentColor|board.Knight]
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

		opponentPawns := p.Bitboards[opponentColor|board.Pawn]
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

		opponentAttacks |= ComputedKingMoves[bits.TrailingZeros64(p.Bitboards[opponentColor|board.King])]

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
			rookAttacks := PGetRookMoves(friendlyKingIndex, maskedBlockers)
			return (rookAttacks & opponentOrthogonalSliders) != 0
		}

		if opponentDiagonalSliders != 0 {
			bishopAttacks := PGetBishopMoves(friendlyKingIndex, maskedBlockers)
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
			pseudoLegalMoves[index] = board.NewMove(friendlyKingIndex, bits.TrailingZeros64(kingMoveMask))
			index++

			kingMoveMask &= kingMoveMask - 1
		}

		initialKingIndex := 4
		if friendlyColor != board.White {
			initialKingIndex = 60
		}

		// King is not on its original square, will not be allowed to castle
		if inCheck || friendlyKingIndex != initialKingIndex {
			return
		}

		kingSideAllowed := p.CastlingRights&0b1000 != 0
		queenSideAllowed := p.CastlingRights&0b0100 != 0

		var kingSideEmptyMask uint64 = 0b1100000
		var queenSideEmptyMask uint64 = 0b1110

		kingSideRookIndex := 7
		queenSideRookIndex := 0

		var kingSideAttackMask uint64 = 1<<friendlyKingIndex | 1<<(friendlyKingIndex+1) | 1<<(friendlyKingIndex+2)
		var queenSideAttackMask uint64 = 1<<friendlyKingIndex | 1<<(friendlyKingIndex-1) | 1<<(friendlyKingIndex-2)

		if !p.WhiteToMove {
			kingSideAllowed = p.CastlingRights&0b0010 != 0
			queenSideAllowed = p.CastlingRights&0b0001 != 0

			kingSideEmptyMask <<= 56
			queenSideEmptyMask <<= 56

			kingSideRookIndex += 56
			queenSideRookIndex += 56
		}

		if kingSideAllowed &&
			(kingSideAttackMask&opponentAttacks) == 0 && // King does not start or pass through attacked field
			(kingSideEmptyMask&allPieces) == 0 && // All fields are empty
			board.IsIndexBitSet(kingSideRookIndex, friendlyRooks) { // There is a rook on its field

			pseudoLegalMoves[index] = board.NewMove(friendlyKingIndex, friendlyKingIndex+2, board.WithRookStartingSquare(kingSideRookIndex))
			index++
		}

		if queenSideAllowed &&
			(queenSideAttackMask&opponentAttacks) == 0 &&
			(queenSideEmptyMask&allPieces) == 0 &&
			board.IsIndexBitSet(queenSideRookIndex, friendlyRooks) {

			pseudoLegalMoves[index] = board.NewMove(friendlyKingIndex, friendlyKingIndex-2, board.WithRookStartingSquare(queenSideRookIndex))
			index++
		}
	}

	var PawnMoves = func() {
		for friendlyPawns != 0 {
			// Get index of LSB
			pieceIndex := bits.TrailingZeros64(friendlyPawns)
			targetIndex := pieceIndex + p.PawnOffset

			// Continue if target index is out of bounds, just go to the next iteration
			if targetIndex < 0 || targetIndex > 63 {
				friendlyPawns &= friendlyPawns - 1
				continue
			}

			// Check if the pawn is on starting square
			isBaseRank := pieceIndex/8 == 1
			isPromotionRank := pieceIndex/8 == 6

			if !p.WhiteToMove {
				isBaseRank, isPromotionRank = isPromotionRank, isBaseRank
			}

			// Handling normal moves
			validMoves := ComputedPawnMoves[friendlyIndex][pieceIndex] & ^allPieces

			// Adding double pawn pushes
			if isBaseRank && validMoves != 0 {
				validMoves |= ComputedPawnMoves[friendlyIndex][pieceIndex+p.PawnOffset] & ^allPieces
			}

			// Generate all attacks
			attackMask := ComputedPawnAttacks[p.FriendlyIndex][pieceIndex]
			validAttacks := attackMask & opponentPieces

			// Generate en passant attacks
			validPawnMoveMask := validMoveMask
			if p.EnPassantSquare != -1 {
				if 1<<(p.EnPassantSquare-p.PawnOffset)&validMoveMask != 0 {
					validPawnMoveMask |= 1 << p.EnPassantSquare
				}
				validAttacks |= attackMask & (1 << p.EnPassantSquare)
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
					pseudoLegalMoves[index] = board.NewMove(pieceIndex, moveTargetIndex, board.WithPromotion(p.FriendlyColor|board.Knight))
					index++
					pseudoLegalMoves[index] = board.NewMove(pieceIndex, moveTargetIndex, board.WithPromotion(p.FriendlyColor|board.Bishop))
					index++
					pseudoLegalMoves[index] = board.NewMove(pieceIndex, moveTargetIndex, board.WithPromotion(p.FriendlyColor|board.Queen))
					index++
					pseudoLegalMoves[index] = board.NewMove(pieceIndex, moveTargetIndex, board.WithPromotion(p.FriendlyColor|board.Rook))
					index++
				} else {
					if moveTargetIndex == p.EnPassantSquare {
						if !IsEnPassantMovePinned(pieceIndex, moveTargetIndex, moveTargetIndex-p.PawnOffset) {
							pseudoLegalMoves[index] = board.NewMove(pieceIndex, moveTargetIndex, board.WithEnPassantCaptureSquare(moveTargetIndex-p.PawnOffset))
							index++
						}
					} else if moveTargetIndex == targetIndex+p.PawnOffset {
						pseudoLegalMoves[index] = board.NewMove(pieceIndex, moveTargetIndex, board.WithEnPassantPassedSquare(moveTargetIndex-p.PawnOffset))
						index++
					} else {
						pseudoLegalMoves[index] = board.NewMove(pieceIndex, moveTargetIndex)
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
				sliderMoves := PGetRookMoves(pieceIndex, allPieces) & ^friendlyPieces & validMoveMask

				if (friendlyPinRays & (1 << pieceIndex)) != 0 {
					sliderMoves &= AlignmentMask[pieceIndex][friendlyKingIndex]
				}

				for sliderMoves != 0 {
					targetIndex := bits.TrailingZeros64(sliderMoves)

					pseudoLegalMoves[index] = board.NewMove(pieceIndex, targetIndex)
					index++

					sliderMoves &= sliderMoves - 1
				}

				friendlyOrthogonalSliders &= friendlyOrthogonalSliders - 1
			}
			if friendlyDiagonalSliders&(1<<pieceIndex) != 0 {
				sliderMoves := PGetBishopMoves(pieceIndex, allPieces) & ^friendlyPieces & validMoveMask

				if (friendlyPinRays & (1 << pieceIndex)) != 0 {
					sliderMoves &= AlignmentMask[pieceIndex][friendlyKingIndex]
				}

				for sliderMoves != 0 {
					targetIndex := bits.TrailingZeros64(sliderMoves)

					pseudoLegalMoves[index] = board.NewMove(pieceIndex, targetIndex)
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
				pseudoLegalMoves[index] = board.NewMove(pieceIndex, bits.TrailingZeros64(validMoves))
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

	p.UpdateTerminalState(len(pseudoLegalMoves) != 0, (opponentAttacks>>friendlyKingIndex)&1 != 0)

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

func Perft(p *board.Position, ply int, maxPly int) int64 {
	/*
		Perft Testing Utility
	*/

	if ply == 0 {
		return 1
	}

	legalMoves := LegalMoves(p)
	var totalNodes int64 = 0

	if nodes, found := tt[ply][p.Zobrist]; found {
		return nodes
	}

	// Not the official implementation, but works a lot faster
	if ply == 1 {
		return int64(len(legalMoves))
	}

	for _, m := range legalMoves {
		np := p.MakeMove(m)

		subNodes := Perft(np, ply-1, maxPly)
		totalNodes += subNodes

		if ply == maxPly {
			fmt.Printf("%s: %d\n", board.MoveToString(m), subNodes)
		}
	}

	tt[ply][p.Zobrist] = totalNodes
	return totalNodes
}
