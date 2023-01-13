package engine

import (
	"sort"

	"github.com/noahklein/dragon"
)

// Sort moves using cheap heuristics, e.g. search captures and promotions before other moves.
// Searching better moves first helps us prune nodes with beta cutoffs.
func (e *Engine) sortMoves(moves []dragon.Move) []dragon.Move {
	var (
		out, killers, checks, captures, others []dragon.Move
	)

	pv, pvOk := e.PVMove()

	kms := e.killer.Get(e.ply)

	for _, move := range moves {
		if pvOk && move == pv {
			out = append(out, move)
		} else if move == kms[0] || move == kms[1] { // Zero-value is a1a1, an impossible move.
			e.nodeCount.legalKiller++
			killers = append(killers, move)
		} else if move.Promote() == dragon.Queen || IsCheck(e.board, move) {
			checks = append(checks, move)
		} else if move.Promote() != 0 || Occupied(e.board, move.To()) {
			captures = append(captures, move)
		} else {
			others = append(others, move)
		}
	}

	e.mvvLva(captures)

	out = append(out, killers...)
	out = append(out, checks...)
	out = append(out, captures...)
	return append(out, others...)
}

// Most-Valuable Victim/Least-Valuable attacker. Search PxQ, before QxP.
func (e *Engine) mvvLva(captures []dragon.Move) {
	sort.Slice(captures, func(i, j int) bool {
		fa, _ := e.squares.PieceType(captures[i].From())
		ta, _ := e.squares.PieceType(captures[i].To())

		fb, _ := e.squares.PieceType(captures[j].From())
		tb, _ := e.squares.PieceType(captures[j].To())

		return ta-fa > tb-fb
	})
}

// Occupied checks if a square is occupied.
func Occupied(board *dragon.Board, square uint8) bool {
	return (board.Black.All|board.White.All)&uint64(1<<square) >= 1
}

func IsCheck(board *dragon.Board, move dragon.Move) bool {
	unapply := board.Apply(move)
	defer unapply()
	return board.OurKingInCheck()
}

// Check for search extensions. Return the number of plies to extend search by.
func (e *Engine) extensions(move dragon.Move, depth int) int {
	if move.Promote() > 0 {
		return 1
	}
	return 0
}

// Check for search reductions. Return the number of plies to reduce search by.
func (e *Engine) reductions(move dragon.Move, depth int) int {
	return 0
}
