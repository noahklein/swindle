package engine

import (
	"math/bits"

	"github.com/noahklein/chess/bitboard"
	"github.com/noahklein/dragon"
)

const (
	pawnVal   = 100
	knightVal = 320
	bishopVal = 330
	rookVal   = 500
	queenVal  = 900
	kingVal   = queenVal * 2
)

var (
	doubledPawn  = [2]int16{-10, -10}
	isolatedPawn = [2]int16{-10, -10}

	// Passed pawn bonuses for each rank.
	passedPawn = [8]int16{0, 10, 30, 50, 75, 100, 150, 200}
	// Bonus for rooks on open and semi-open files.
	rookOpen     = [2]int16{10, 20}
	rookSemiOpen = [2]int16{5, 7}
)

var PieceValue = [...]int16{0, pawnVal, knightVal, bishopVal, rookVal, queenVal, kingVal}

// How much to increment the game phase counter for each piece type.
var gamePhaseInc = [...]int16{0, 1, 1, 2, 4, 0}

func Eval(board *dragon.Board) int16 {
	// Game phase is incremented for each piece.
	var phase int16

	var (
		// [White, Black]
		mgScore [2]int16
		egScore [2]int16

		pieces = [2]dragon.Bitboards{board.White, board.Black}
	)

	// Give bonus points for piece positions and other heuristics.
	for square := uint8(0); square < 64; square++ {
		piece, isWhite := dragon.GetPieceType(square, board)
		if piece == dragon.Nothing {
			continue
		}

		color := 0
		if !isWhite {
			color = 1
		}

		switch piece {
		case dragon.Pawn:
			fileMask := bitboard.FileMask[square]
			// Doubled pawn penalty.
			ourPawnsOnFile := pieces[color].Pawns & fileMask
			if ourPawnsOnFile > 1 {
				mgScore[color] += doubledPawn[0]
				egScore[color] += doubledPawn[1]
			}
			// Passed pawn bonus.
			if bitboard.PassedMask[color][square]&pieces[other(color)].Pawns == 0 {
				rank := bitboard.Rank(square)
				if color == 1 {
					rank = 8 - rank
				}
				mgScore[color] += passedPawn[rank]
				egScore[color] += passedPawn[rank]
			}
			// Isolated pawn penalty.
			if bitboard.AdjacentMask[square]&pieces[color].Pawns == 0 {
				mgScore[color] += isolatedPawn[0]
				egScore[color] += isolatedPawn[1]
			}

		case dragon.Rook:
			pawnsOnFile := (board.White.Pawns | board.Black.Pawns) & bitboard.FileMask[square]
			if pawnsOnFile == 0 {
				mgScore[color] += rookOpen[0]
				egScore[color] += rookOpen[1]
			} else if pawnsOnFile == 1 {
				mgScore[color] += rookSemiOpen[0]
				egScore[color] += rookSemiOpen[1]
			}
		case dragon.King:
			// King-safety.
			pawnsOnFile := (board.White.Pawns | board.Black.Pawns) & bitboard.FileMask[square]
			if pawnsOnFile == 0 {
				mgScore[color] -= rookOpen[0]
				egScore[color] -= rookOpen[1]
			} else if pawnsOnFile == 1 {
				mgScore[color] -= rookSemiOpen[0]
				egScore[color] -= rookSemiOpen[1]
			}
		}

		pc := pieceColor(piece, isWhite)

		mgScore[color] += MidGameTable[pc][square]
		egScore[color] += EndGameTable[pc][square]
		phase += gamePhaseInc[piece-1]
	}

	material := pieceEval(&board.White) - pieceEval(&board.Black)

	// Tapered evaluation: weigh midgame and endgame scores to smoothly transition between
	// game phases.
	mg := mgScore[0] - mgScore[1]
	eg := egScore[0] - mgScore[1]

	mgWeight := min(phase, 24)
	egWeight := 24 - mgWeight

	phaseScore := (mg*mgWeight + eg*egWeight) / 24
	return whiteToMove(board) * (material + phaseScore)
}

// Number from 0 to 24 indicating how much material is on the board.
func materialCount(b *dragon.Board) int16 {
	var phase int16
	for sq := uint8(0); sq < 64; sq++ {
		piece, _ := dragon.GetPieceType(sq, b)
		if piece == dragon.Nothing {
			continue
		}

		phase += gamePhaseInc[piece-1]
	}
	return min(phase, 24)
}

func pieceEval(b *dragon.Bitboards) int16 {
	score := bits.OnesCount64(b.Pawns)*pawnVal +
		bits.OnesCount64(b.Knights)*knightVal +
		bits.OnesCount64(b.Bishops)*bishopVal +
		bits.OnesCount64(b.Rooks)*rookVal +
		bits.OnesCount64(b.Queens)*queenVal
	return int16(score)
}

func badCapture(attacker, victim int16) bool {
	// Pawn captures don't lose material.
	if attacker == dragon.Pawn {
		return false
	}

	attVal, vicVal := PieceValue[attacker], PieceValue[victim]
	return vicVal < attVal-50
}

const (
	maxMate = 400
	NotMate = 500
)

// Converts an eval score into ply till mate. Returns NotMate if not mating.
func mateScore(score int16, ply int16) int16 {
	var mate int16 = NotMate
	plyTillMate := -mateVal - abs(score) - ply
	if plyTillMate < maxMate {
		mate = plyTillMate / 2

		if score < 0 {
			mate = -mate
		}
	}

	return mate
}

func other(color int) int { return 1 - color }

// Branchless abs. Only works if MinInt16 <= a <= MaxInt16.
func abs(n int16) int16 {
	sgn := n >> 15
	n ^= sgn
	return n - sgn
}

// Branchless max. Only works if MinInt16 <= a - b <= MaxInt16.
func max(a, b int16) int16 {
	diff := a - b
	dsgn := diff >> 15
	return a - (diff & dsgn)
}

// Branchless min. Only works if MinInt16 <= a - b <= MaxInt16.
func min(a, b int16) int16 {
	diff := a - b
	dsgn := diff >> 15
	return b + (diff & dsgn)
}
