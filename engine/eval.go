package engine

import (
	"math"
	"math/bits"

	"github.com/dylhunn/dragontoothmg"
)

const (
	mateVal   int16 = math.MinInt16 / 2
	pawnVal         = 100
	knightVal       = 300
	bishopVal       = 320
	rookVal         = 500
	queenVal        = 900
)

func Eval(board dragontoothmg.Board) int16 {
	return whiteToMove(board) * (pieceEval(board.White) - pieceEval(board.Black))
}

func pieceEval(b dragontoothmg.Bitboards) int16 {
	score := bits.OnesCount64(b.Pawns)*pawnVal +
		bits.OnesCount64(b.Knights)*knightVal +
		bits.OnesCount64(b.Bishops)*bishopVal +
		bits.OnesCount64(b.Rooks)*rookVal +
		bits.OnesCount64(b.Queens)*queenVal
	return int16(score)

}
