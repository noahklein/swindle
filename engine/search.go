package engine

import (
	"context"

	"github.com/dylhunn/dragontoothmg"
	"github.com/noahklein/chess/uci"
)

var nodes int

// Root search.
func (e *Engine) Search(ctx context.Context, params uci.SearchParams) uci.SearchResults {
	nodes = 0 // reset node count.

	moves := e.board.GenerateLegalMoves()

	maxScore := mateVal
	var bestMove dragontoothmg.Move
	for _, move := range moves {
		if ctx.Err() != nil {
			break
		}

		nodes++
		unapply := e.board.Apply(move)
		score := -e.AlphaBeta(mateVal, -mateVal, params.Depth)
		if score >= maxScore {
			e.Print("move %s score %v", move.String(), score)
			maxScore = score
			bestMove = move
		}
		unapply()
	}

	return uci.SearchResults{
		BestMove: bestMove.String(),
		Score:    maxScore / 100,
		Nodes:    nodes,
		Depth:    params.Depth,
	}
}

// AlphaBeta improves upon the minimax algorithm.
//     Alpha is the lowest score the maximizing player can force
//     Beta is the highest score the minimizing player can force.
// It stops evaluating a move when at least one possibility has been found that
// proves the move to be worse than a previously examined move. In other words,
// you only need one refutation to a move to know it's bad.
func (e *Engine) AlphaBeta(alpha, beta int16, depth int) int16 {
	nodes++
	moves := e.sortMoves(e.board.GenerateLegalMoves())

	// Checkmate
	if len(moves) == 0 && e.board.OurKingInCheck() {
		return whiteToMove(e.board) * mateVal
	}
	// Draw
	if len(moves) == 0 {
		return 0
	}

	if depth == 0 {
		// return Eval(e.board)
		return e.Quiesce(alpha, beta)
	}

	for _, move := range moves {
		unapply := e.board.Apply(move)
		score := -e.AlphaBeta(-beta, -alpha, depth-1)
		unapply()

		if score >= beta {
			return beta
		}
		if score > alpha {
			alpha = score
		}
	}

	return alpha
}

func (e *Engine) Quiesce(alpha, beta int16) int16 {
	score := Eval(e.board)
	if score >= beta {
		return beta
	}
	if alpha < score {
		alpha = score
	}

	for _, move := range e.board.GenerateLegalMoves() {
		// Skip non-captures.
		if !Occupied(e.board, move.To()) {
			continue
		}

		unapply := e.board.Apply(move)
		score := -e.Quiesce(-beta, -alpha)
		unapply()

		if score >= beta {
			return beta
		}
		if score > alpha {
			alpha = score
		}
	}

	return alpha
}

// Sort moves using heuristics, e.g. search captures and promotions before other moves.
// Searching better moves first helps us prune nodes with beta cutoffs.
func (e *Engine) sortMoves(moves []dragontoothmg.Move) []dragontoothmg.Move {
	var (
		captures, others []dragontoothmg.Move
	)

	for _, move := range moves {
		if Occupied(e.board, move.To()) {
			captures = append(captures, move)
		} else {
			others = append(others, move)
		}
	}

	return append(captures, others...)
}

// Occupied checks if a square is occupied.
func Occupied(board *dragontoothmg.Board, square uint8) bool {
	return (board.Black.All|board.White.All)&uint64(1<<square) >= 1
}

func contains(bitset uint64, square uint8) bool {
	return bitset&(1<<square) >= 1
}

func whiteToMove(board *dragontoothmg.Board) int16 {
	if board.Wtomove {
		return 1
	}
	return -1
}
