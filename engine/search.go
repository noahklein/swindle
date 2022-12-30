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
		nodes++
		unapply := e.board.Apply(move)
		// e.Print("alpha = %v, beta = %v", wtm*mateVal, wtm*-mateVal)
		score := e.AlphaBeta(mateVal, -mateVal, params.Depth)
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

func (e *Engine) AlphaBeta(alpha, beta int16, depth int) int16 {
	nodes++
	wToMove := whiteToMove(e.board)
	moves := e.sortMoves(e.board.GenerateLegalMoves())
	// Checkmate
	if len(moves) == 0 {
		return wToMove * mateVal
	}
	if depth == 0 {
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

	return score
}

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
func Occupied(board dragontoothmg.Board, square uint8) bool {
	return (board.Black.All|board.White.All)&uint64(1<<square) >= 1
}

func whiteToMove(board dragontoothmg.Board) int16 {
	if board.Wtomove {
		return 1
	}
	return -1
}
