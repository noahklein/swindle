package engine

import (
	"math"
	"math/bits"

	"github.com/dylhunn/dragontoothmg"
)

const (
	mateVal   int16 = math.MinInt16 / 2
	pawnVal         = 100
	knightVal       = 320
	bishopVal       = 330
	rookVal         = 500
	queenVal        = 900
)

func Eval(board *dragontoothmg.Board) int16 {
	score := pieceEval(&board.White) - pieceEval(&board.Black)

	phase := gamePhase(board)

	// Give bonus points for piece positions.
	for square := uint8(0); square < 64; square++ {
		color, piece := At(board, square)
		if color == Empty {
			continue
		}

		posBonus := midGameTable[pieceColor(int(piece), color)][square]
		if phase == EndGame {
			posBonus = endGameTable[pieceColor(int(piece), color)][square]
		}

		if color == White {
			score += posBonus
		} else {
			score -= posBonus
		}
	}

	return whiteToMove(board) * score
}

func pieceEval(b *dragontoothmg.Bitboards) int16 {
	score := bits.OnesCount64(b.Pawns)*pawnVal +
		bits.OnesCount64(b.Knights)*knightVal +
		bits.OnesCount64(b.Bishops)*bishopVal +
		bits.OnesCount64(b.Rooks)*rookVal +
		bits.OnesCount64(b.Queens)*queenVal
	return int16(score)
}

func pieceCount(b *dragontoothmg.Bitboards) int {
	return bits.OnesCount64(b.Knights) +
		bits.OnesCount64(b.Bishops) +
		bits.OnesCount64(b.Rooks) +
		bits.OnesCount64(b.Queens)
}

// TODO: improve endgame detection.
func gamePhase(b *dragontoothmg.Board) GamePhase {
	if pieceCount(&b.White)+pieceCount(&b.Black) < 7 {
		return EndGame
	}
	return MiddleGame
}
