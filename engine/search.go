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

	// Increase depth in endgame.
	phase := gamePhase(e.board)
	if phase == EndGame {
		params.Depth += 1
	}

	moves := e.board.GenerateLegalMoves()
	bestScore := mateVal
	var bestMove dragontoothmg.Move

	alpha, beta := mateVal, -mateVal

	// Iterative deepening with aspirational window. After each iteration we use the eval
	// as the center of the alpha-beta window, and search again one ply deeper. If the eval
	// falls outside of the window, we re-search on the same depth with a wider window.
	for depth := 0; depth <= params.Depth; {
		for _, move := range moves {
			if ctx.Err() != nil {
				break
			}

			nodes++
			unmove := e.Move(move)
			score := -e.AlphaBeta(-beta, -alpha, depth)
			unmove()

			if score >= bestScore {
				e.Print("move %s: %v", move.String(), score)
				bestScore = score
				bestMove = move
			}
		}

		if bestScore < alpha || bestScore > beta {
			e.Print("Eval was outside of aspirational window. Re-search at same depth, %v.", depth)
			alpha = -mateVal
			beta = mateVal
			continue
		}
		alpha = bestScore - pawnVal/4
		beta = bestScore + pawnVal/4
		depth++
		e.Print("Eval was inside aspirational window! New window: %v, %v.", alpha, beta)
	}

	return uci.SearchResults{
		BestMove: bestMove.String(),
		Score:    bestScore / 100,
		Nodes:    nodes,
		Depth:    params.Depth,
	}
}

// AlphaBeta improves upon the minimax algorithm.
//     Alpha is the lowest score the maximizing player can force
//     Beta is the highest score the minimizing player can force.
// It stops evaluating a move when at least one possibility has been found that
// proves the move to be worse than a previously examined move. In other words,
// you only need one refutation to know a move is bad.
func (e *Engine) AlphaBeta(alpha, beta int16, depth int) int16 {
	nodes++
	moves := e.board.GenerateLegalMoves()

	// Checkmate
	if len(moves) == 0 && e.board.OurKingInCheck() {
		return mateVal
	}
	// Draw
	if len(moves) == 0 {
		// fmt.Println("alphabeta draw hit")
		return 0
	}

	if depth == 0 {
		// return Eval(e.board)
		return e.Quiesce(alpha, beta)
	}

	for _, move := range e.sortMoves(moves) {
		unmove := e.Move(move)
		score := -e.AlphaBeta(-beta, -alpha, depth-1)
		unmove()

		// Beta-cutoff; better than the best move.
		if score >= beta {
			e.killer.Add(int(e.ply), move)
			return beta
		}
		if score > alpha {
			alpha = score
		}
	}

	return alpha
}

func (e *Engine) Quiesce(alpha, beta int16) int16 {
	moves := e.board.GenerateLegalMoves()
	if len(moves) == 0 && e.board.OurKingInCheck() {
		return mateVal
	} else if len(moves) == 0 {
		// fmt.Println("quiescence draw hit")
		return 0
	}

	nodes++
	score := Eval(e.board)
	if score >= beta {
		return beta
	}
	if alpha < score {
		alpha = score
	}

	for _, move := range moves {
		// Skip non-captures.
		if !Occupied(e.board, move.To()) {
			continue
		}

		unmove := e.Move(move)
		score := -e.Quiesce(-beta, -alpha)
		unmove()

		if score >= beta {
			e.killer.Add(e.ply, move)
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
		killers, checks, captures, others []dragontoothmg.Move
	)

	kms := e.killer.Get(e.ply)

	for _, move := range moves {
		// The zero-value for moves is a1a1, an impossible move.
		if move == kms[0] || move == kms[1] {
			killers = append(killers, move)
		}
		if IsCheck(e.board, move) {
			checks = append(checks, move)

		} else if Occupied(e.board, move.To()) {
			captures = append(captures, move)
		} else {
			others = append(others, move)
		}
	}

	out := append(killers, checks...)
	out = append(out, captures...)
	return append(out, others...)
}

// Occupied checks if a square is occupied.
func Occupied(board *dragontoothmg.Board, square uint8) bool {
	return (board.Black.All|board.White.All)&uint64(1<<square) >= 1
}

func IsCheck(board *dragontoothmg.Board, move dragontoothmg.Move) bool {
	unapply := board.Apply(move)
	defer unapply()
	return board.OurKingInCheck()
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
