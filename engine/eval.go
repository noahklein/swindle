package engine

import (
	"math/bits"

	"github.com/noahklein/dragon"
)

const (
	pawnVal   = 100
	knightVal = 320
	bishopVal = 330
	rookVal   = 500
	queenVal  = 900
)

func Eval(board *dragon.Board) int16 {
	score := pieceEval(&board.White) - pieceEval(&board.Black)
	// var score int16

	phase := gamePhase(board)

	// Give bonus points for piece positions.
	for square := uint8(0); square < 64; square++ {
		piece, isWhite := dragon.GetPieceType(square, board)
		if piece == dragon.Nothing {
			continue
		}

		posBonus := MidGameTable[pieceColor(int(piece), isWhite)][square]
		if phase == EndGame {
			posBonus = EndGameTable[pieceColor(int(piece), isWhite)][square]
		}

		if isWhite {
			score += posBonus
		} else {
			score -= posBonus
		}
	}

	return whiteToMove(board) * score
}

func pieceEval(b *dragon.Bitboards) int16 {
	score := bits.OnesCount64(b.Pawns)*pawnVal +
		bits.OnesCount64(b.Knights)*knightVal +
		bits.OnesCount64(b.Bishops)*bishopVal +
		bits.OnesCount64(b.Rooks)*rookVal +
		bits.OnesCount64(b.Queens)*queenVal
	return int16(score)
}

func pieceCount(b *dragon.Bitboards) int {
	return bits.OnesCount64(b.Knights) +
		bits.OnesCount64(b.Bishops) +
		bits.OnesCount64(b.Rooks) +
		bits.OnesCount64(b.Queens)
}

// TODO: improve endgame detection.
func gamePhase(b *dragon.Board) GamePhase {
	if b.White.Queens == 0 && b.Black.Queens == 0 {
		return EndGame
	}
	if pieceCount(&b.White)+pieceCount(&b.Black) < 6 {
		return EndGame
	}
	return MiddleGame
}

func abs(n int16) int16 {
	if n < 0 {
		return -n
	}
	return n
}
