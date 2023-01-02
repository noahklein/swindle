package engine

import (
	"context"
	"sort"
	"time"

	"github.com/dylhunn/dragontoothmg"
	"github.com/noahklein/chess/uci"
)

var nodes int

const (
	mateVal      int16 = -15000
	initialAlpha int16 = -20000
	initialBeta  int16 = 20000
)

// Root search.
func (e *Engine) Search(ctx context.Context, params uci.SearchParams) uci.SearchResults {
	nodes = 0 // reset node count.

	// Increase depth in endgame.
	phase := gamePhase(e.board)
	if phase == EndGame {
		params.Depth += 1
	}

	// TODO: Smarter time management; look at remaining clock?
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	bestScore := initialAlpha
	var bestMove dragontoothmg.Move

	for _, move := range e.board.GenerateLegalMoves() {
		nodes++
		unmove := e.Move(move)
		score := e.IterDeep(ctx, params.Depth)
		unmove()

		if score >= bestScore {
			e.Print("move %s: %v; d=%v", move.String(), score, depth)
			bestScore = score
			bestMove = move
		}
	}

	var mate int16
	plyTillMate := abs(mateVal) - abs(bestScore)
	if plyTillMate < 600 {
		mate = plyTillMate / 2

		if bestScore < 0 {
			mate *= -1
		}
	}

	return uci.SearchResults{
		BestMove: bestMove.String(),
		Score:    bestScore / 100,
		Mate:     mate,
		Nodes:    nodes,
		Depth:    params.Depth,
	}
}

// Iterative deepening with aspiration window. After each iteration we use the eval
// as the center of the alpha-beta window, and search again one ply deeper. If the eval
// falls outside of the window, we re-search on the same depth with a wider window.
func (e *Engine) IterDeep(ctx context.Context, maxDepth int) int16 {
	var score int16
	alpha, beta := initialAlpha, initialBeta
	const window = pawnVal / 2

	for depth := 0; depth <= maxDepth; {
		if ctx.Err() != nil {
			return score
		}

		score = -e.AlphaBeta(-beta, -alpha, depth)
		if score <= alpha || score >= beta {
			e.Print("Eval was outside of aspirational window. Re-search at same depth, %v.", depth)
			alpha, beta = initialAlpha, initialBeta
			continue
		}
		alpha, beta = score-window, score+window
		e.Print("Eval was inside aspirational window! New window: %v, %v.", alpha, beta)
		depth++
	}

	return score
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
		return mateVal + int16(e.ply)
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

	var foundPV bool
	for _, move := range e.sortMoves(moves) {
		unmove := e.Move(move)

		var score int16
		if foundPV {
			// Search tiny window.
			score = -e.AlphaBeta(-alpha-1, -alpha, depth-1)
			// If we failed, search again with normal window.
			if score > alpha && score < beta {
				score = -e.AlphaBeta(-beta, -alpha, depth-1)
			}
		} else {
			score = -e.AlphaBeta(-beta, -alpha, depth-1)
		}
		unmove()

		// Beta-cutoff; better than the best move.
		if score >= beta {
			e.killer.Add(int(e.ply), move)
			return beta
		}
		if score > alpha {
			alpha = score
			foundPV = true
		}
	}

	return alpha
}

// Quiesce runs a limited search on checks and captures until it reaches a quiet position.
// Eval() is unreliable in "loud" positions as there might be a queen hanging or worse.
// Quiescent search avoids the "horizon effect".
// Note: 50%-90% of nodes searched are here, pruning goes a long way.
func (e *Engine) Quiesce(alpha, beta int16) int16 {

	// Checks are extra noisy. Search one move deeper.
	if e.board.OurKingInCheck() {
		return e.AlphaBeta(alpha, beta, 1)
	}

	nodes++
	score := Eval(e.board)
	if score >= beta {
		return beta
	}
	if alpha < score {
		alpha = score
	}

	for _, move := range e.board.GenerateLegalMoves() {
		// Skip non-captures.
		// TODO: also skip bad captures, e.g. QxP.
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

	// Most-Valuable Victim/Least-Valuable attacker. Search PxQ, before QxP.
	sort.Slice(captures, func(i, j int) bool {
		_, f1 := At(e.board, captures[i].From())
		_, f2 := At(e.board, captures[j].From())
		_, t1 := At(e.board, captures[i].To())
		_, t2 := At(e.board, captures[j].To())
		return t1-f1 > t2-f2
	})

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
