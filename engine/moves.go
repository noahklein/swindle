package engine

import (
	"github.com/noahklein/dragon"
)

const (
	killerScore = 16
)

var (
	promotionScore = [7]int16{0, 0, 16, 17, 36, 49, 0}
	mvvLvaTable    = [7][7]int16{
		{0, 0, 0, 0, 0, 0, 0},       // victim 0, Not a capture.
		{0, 15, 14, 13, 12, 11, 10}, // victim P, attacker 0, P, N, B, R, Q, K
		{0, 25, 24, 23, 22, 21, 20}, // victim N, attacker 0, P, N, B, R, Q, K
		{0, 35, 34, 33, 32, 31, 30}, // victim B, attacker 0, P, N, B, R, Q, K
		{0, 45, 44, 43, 42, 41, 40}, // victim R, attacker 0, P, N, B, R, Q, K
		{0, 55, 54, 53, 52, 51, 50}, // victim Q, attacker 0, P, N, B, R, Q, K
		{0, 0, 0, 0, 0, 0, 0},       // victim K, King can't be captured.
	}
)

type moveScore struct {
	move  dragon.Move
	score int16
}

// MoveSort sorts moves using cheap heuristics, e.g. search captures and promotions
//  before other moves. Searching better moves first helps us prune nodes with beta cutoffs.
type MoveSort struct {
	moveScores []moveScore
}

// Score moves and store them. Call Next(i) from 0 to len(moves) to get the sorted list.
func (e *Engine) newMoveSorter(moves []dragon.Move) *MoveSort {
	ms := MoveSort{
		moveScores: make([]moveScore, len(moves)),
	}

	pv, pvOk := e.PVMove()
	kms := e.killer.Get(e.ply)
	for i, move := range moves {
		ms.moveScores[i].move = move

		if pvOk && move == pv {
			ms.moveScores[i].score = 10000
			continue
		}

		attacker, _ := e.squares.PieceType(move.From())
		victim, _ := e.squares.PieceType(move.To())
		ms.moveScores[i].score = mvvLvaTable[victim][attacker]
		ms.moveScores[i].score += promotionScore[move.Promote()]

		if move == kms[0] || move == kms[1] {
			ms.moveScores[i].score += killerScore
		}
	}
	return &ms
}

// Next gets the next best move.
func (ms *MoveSort) Next(start int) dragon.Move {
	for i := start; i < len(ms.moveScores); i++ {
		if ms.moveScores[i].score > ms.moveScores[start].score {
			ms.moveScores[start], ms.moveScores[i] = ms.moveScores[i], ms.moveScores[start]
		}
	}

	return ms.moveScores[start].move
}
