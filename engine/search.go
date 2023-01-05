package engine

import (
	"context"
	"sort"
	"time"

	"github.com/noahklein/chess/uci"
	"github.com/noahklein/dragon"
)

// Counts nodes visited in a search.
// TODO: Make goroutine-safe.
var nodes int

const (
	mateVal      int16 = -15000
	drawVal      int16 = 0
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

	// TODO: Smarter time management; look at remaining clock.
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	moves := e.GenMoves()
	if len(moves) == 0 {
		e.Error("Search() called on game that has already ended.")
	}

	bestScore := initialAlpha
	bestMove := moves[0]

	for _, move := range e.GenMoves() {
		nodes++
		unmove := e.Move(move)
		score := e.IterDeep(ctx, params.Depth)
		unmove()

		if score >= bestScore {
			e.Print("move %s: %v", move.String(), score)
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

	var pv []string
	for _, move := range e.PrincipalVariation(bestMove, 2) {
		pv = append(pv, move.String())
	}

	return uci.SearchResults{
		BestMove: bestMove.String(),
		Score:    bestScore / 100,
		Mate:     mate,
		Nodes:    nodes,
		PV:       pv,
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
		// Eval outside of aspiration window, re-search at same depth with wider window.
		if score <= alpha || score >= beta {
			alpha, beta = initialAlpha, initialBeta
			continue
		}
		// Eval inside of window.
		alpha, beta = score-window, score+window
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

	if e.Threefold() {
		return drawVal
	}

	moves := e.GenMoves()
	// Checkmate
	if len(moves) == 0 && e.board.OurKingInCheck() {
		return mateVal + int16(e.ply)
	}
	if len(moves) == 0 {
		return drawVal
	}

	// Check transposition table.
	if val, nt := e.table.GetEval(e.board.Hash(), depth, alpha, beta); nt != NodeUnknown {
		return val
	}

	if depth <= 0 {
		// return Eval(e.board)
		return e.Quiesce(alpha, beta)
	}

	// Assume this is an alpha node.
	nodeType := NodeAlpha

	var foundPV bool
	var bestMove dragon.Move
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

		// Beta-cutoff; better than the previous best move, opponent won't allow this.
		if score >= beta {
			e.killer.Add(e.ply, move)
			e.table.Add(Entry{
				key:   e.board.Hash(),
				depth: depth,
				flag:  NodeBeta,
				value: beta,
				best:  move,
			})
			return beta
		}
		if score > alpha {
			alpha = score
			bestMove = move
			foundPV = true
			nodeType = NodeExact
		}
	}

	e.table.Add(Entry{
		key:   e.board.Hash(),
		depth: depth,
		flag:  nodeType,
		value: alpha,
		best:  bestMove,
	})
	return alpha
}

// Quiesce runs a limited search on checks and captures until it reaches a quiet position.
// Eval() is unreliable in "loud" positions as there might be a queen hanging or worse.
// Quiescent search avoids the "horizon effect".
// Note: 50%-90% of nodes searched are here, pruning goes a long way.
func (e *Engine) Quiesce(alpha, beta int16) int16 {
	if e.Threefold() {
		return drawVal
	}

	moves := e.GenMoves()
	if len(moves) == 0 && e.board.OurKingInCheck() {
		return mateVal + int16(e.ply)
	} else if len(moves) == 0 {
		return drawVal
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

// Gets the principal variation by recursively following the best moves in the
// transposition table.
func (e *Engine) PrincipalVariation(bestMove dragon.Move, depth int) []dragon.Move {
	if depth == 0 {
		return nil
	}

	unmove := e.board.Apply(bestMove)
	defer unmove()

	if entry, ok := e.table.Get(e.board.Hash()); ok {
		return append([]dragon.Move{bestMove}, e.PrincipalVariation(entry.best, depth-1)...)
	}
	return []dragon.Move{bestMove}
}

func (e *Engine) PVMove() (dragon.Move, bool) {
	entry, ok := e.table.Get(e.board.Hash())
	return entry.best, ok

}

func (e *Engine) GenMoves() []dragon.Move {
	if e.Threefold() {
		return nil
	}
	return e.board.GenerateLegalMoves()
}

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
			killers = append(killers, move)
		} else if IsCheck(e.board, move) {
			checks = append(checks, move)
		} else if Occupied(e.board, move.To()) {
			captures = append(captures, move)
		} else {
			others = append(others, move)
		}
	}

	// Most-Valuable Victim/Least-Valuable attacker. Search PxQ, before QxP.
	sort.Slice(captures, func(i, j int) bool {
		f1, _ := dragon.GetPieceType(captures[i].From(), e.board)
		f2, _ := dragon.GetPieceType(captures[j].From(), e.board)
		t1, _ := dragon.GetPieceType(captures[i].To(), e.board)
		t2, _ := dragon.GetPieceType(captures[j].To(), e.board)
		return t1-f1 > t2-f2
	})

	out = append(out, killers...)
	out = append(out, checks...)
	out = append(out, captures...)
	return append(out, others...)
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

func whiteToMove(board *dragon.Board) int16 {
	if board.Wtomove {
		return 1
	}
	return -1
}
