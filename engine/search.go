package engine

import (
	"context"
	"sync"
	"time"

	"github.com/noahklein/chess/uci"
	"github.com/noahklein/dragon"
)

const (
	infinity int16 = 20000
	mateVal  int16 = -15000
	drawVal  int16 = 0
)

// Iterative deepening with aspiration window. After each iteration we use the eval
// as the center of the alpha-beta window, and search again one ply deeper. If the eval
// falls outside of the window, we re-search on the same depth with a wider window.
func (e *Engine) IterDeep(ctx context.Context, params uci.SearchParams) uci.SearchResults {
	// Increase depth with less material. Up to depth + 12 in king and pawn endgames.
	// params.Depth += int(24-materialCount(e.board)) / 2

	const window = pawnVal / 4
	alpha, beta := -infinity, infinity
	exp := 1 // Exponentially increase window on window misses.

	if entry, ok := e.transpositions.Get(e.board.Hash()); ok {
		alpha, beta = entry.value-window, entry.value+window
	}

	var result uci.SearchResults
	for depth := int16(0); ctx.Err() == nil && depth <= int16(params.Depth); {
		start := time.Now()
		result = e.Search(ctx, depth, alpha, beta)
		score := result.Score

		// Eval outside of aspiration window, re-search at same depth with wider window.
		if score <= alpha {
			alpha -= window * (1 << exp)
			exp++
			continue
		} else if score >= beta {
			beta += window * (1 << exp)
			exp++
			continue
		}

		e.UCI(result.Print(start))
		if result.Mate != NotMate {
			e.Warn("Mate found, early return")
			return result
		}

		// Eval inside of window.
		alpha, beta = score-window, score+window
		exp = 1
		depth++
	}

	if ctx.Err() != nil {
		e.Warn("timeout")
	}
	return result
}

// Root search. Runs AlphaBeta search concurrently for each move and collects the results.
func (e *Engine) Search(ctx context.Context, depth, alpha, beta int16) uci.SearchResults {
	moves, _ := e.GenMoves()

	var (
		bestMu    sync.Mutex
		bestScore = -infinity
		bestMove  = moves[0]
		// TODO: investigate mysterious node counts.
		nodes, qNodes int
		maxDepth      int16
		results       uci.SearchResults
	)

	// Search each root-level move in parallel.
	var wg sync.WaitGroup
	for _, move := range moves {
		// Capture variables for goroutine.
		move, e := move, e.Copy()

		wg.Add(1)
		go func() {
			defer wg.Done()
			e.nodeCount.Inc()

			unmove := e.Move(move)
			score := -e.AlphaBeta(-beta, -alpha, int(depth))
			unmove()

			bestMu.Lock()
			defer bestMu.Unlock()
			if score >= bestScore {
				bestScore = score
				bestMove = move

				var pv []string
				for _, m := range e.PrincipalVariation(bestMove, 10) {
					pv = append(pv, m.String())
				}

				nodes += e.nodeCount.nodes
				qNodes += e.nodeCount.qNodes
				maxDepth = max(maxDepth, e.nodeCount.maxPly-e.ply)

				results = uci.SearchResults{
					Move:           bestMove.String(),
					Score:          bestScore,
					Mate:           mateScore(bestScore, e.ply),
					Nodes:          int(nodes),
					Hashfull:       e.transpositions.PermillFull(),
					PV:             pv,
					Depth:          int(depth),
					SelectiveDepth: int(maxDepth),
					TableHits:      e.transpositions.hits,
				}
			}
		}()
	}
	wg.Wait()

	return results
}

// AlphaBeta improves upon the minimax algorithm.
//     Alpha is the lowest score the maximizing player can force
//     Beta is the highest score the minimizing player can force.
// It stops evaluating a move when at least one possibility has been found that
// proves the move to be worse than a previously examined move. In other words,
// you only need one refutation to know a move is bad.
func (e *Engine) AlphaBeta(alpha, beta int16, depth int) int16 {
	e.nodeCount.Inc()

	if e.Draw() {
		return drawVal
	}

	// Mate distance pruning: if we've already found a forced mate on another branch at
	// this ply, then prune this branch.
	//
	// Upper bound, we're mating.
	matingValue := -mateVal - e.ply
	if matingValue < beta {
		beta = matingValue
		if alpha >= matingValue {
			return matingValue
		}
	}
	// Lower bound, we're getting mated.
	matingValue = mateVal + e.ply
	if matingValue > alpha {
		alpha = matingValue
		if beta <= matingValue {
			return matingValue
		}
	}

	moves, inCheck := e.GenMoves()
	if terminalScore, ok := e.terminal(len(moves), inCheck); ok {
		return terminalScore
	}

	// Check transposition table.
	if val, nt := e.transpositions.GetEval(e.board.Hash(), depth, alpha, beta); nt != NodeUnknown {
		return val
	}

	if depth <= 0 {
		return e.Quiesce(alpha, beta)
	}

	if len(moves) == 1 {
		// Only one reply, this ply is free. Extend search.
		// TODO: constrain this.
		depth++
	}

	// Assume this is an alpha node.
	nodeType := NodeAlpha

	var foundPV bool
	var bestMove dragon.Move
	for mNum, move := range e.sortMoves(moves) {
		unmove := e.Move(move)

		// Only search the first 6 sorted moves to full depth.
		lateMoveReduction := 1
		if mNum > 6 && depth >= 3 {
			lateMoveReduction = depth / 3
		}
		// Extend or reduce search depth for this move.
		moveDepth := depth +
			e.extensions(move, depth) - e.reductions(move, depth) -
			lateMoveReduction

		var score int16
		if foundPV {
			// Search tiny window.
			score = -e.AlphaBeta(-alpha-1, -alpha, moveDepth)
			// If we failed, search again with normal window.
			if score > alpha && score < beta {
				score = -e.AlphaBeta(-beta, -alpha, moveDepth)
			}
		} else {
			score = -e.AlphaBeta(-beta, -alpha, moveDepth)
		}
		unmove()

		// Beta-cutoff; better than the previous best move, opponent won't allow this.
		if score >= beta {
			e.killer.Add(e.ply, move)
			e.transpositions.Add(e.ply, Entry{
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

	e.transpositions.Add(e.ply, Entry{
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
// Quiescent search avoids the "horizon effect."
// Note: 50%-90% of nodes searched are here, pruning goes a long way.
func (e *Engine) Quiesce(alpha, beta int16) int16 {
	if e.Draw() {
		return drawVal
	}

	e.nodeCount.Qinc()
	score := Eval(e.board)
	if score >= beta {
		return beta
	}
	// Delta pruning: test if alpha can be improved by greatest material swing. If not,
	// this node is hopeless.
	if score < alpha-queenVal {
		return alpha
	}

	alpha = max(alpha, score)

	// TODO: handle checks specially.
	moves, inCheck := e.GenMoves()
	if terminalScore, ok := e.terminal(len(moves), inCheck); ok {
		return terminalScore
	}

	var loudMoves []dragon.Move
	for _, move := range moves {
		if score, ok := e.terminalMove(move); ok {
			return score
		}

		// Skip non-captures.
		if !Occupied(e.board, move.To()) {
			continue
		}

		// Delta cutoff, this is hopeless.
		victim, _ := dragon.GetPieceType(move.To(), e.board)
		if score+PieceValue[victim]+200 < alpha {
			continue
		}

		attacker, _ := dragon.GetPieceType(move.From(), e.board)
		if badCapture(attacker, victim) {
			continue
		}

		loudMoves = append(loudMoves, move)
	}

	for _, move := range e.sortMoves(loudMoves) {
		unmove := e.Move(move)
		score := -e.Quiesce(-beta, -alpha)
		unmove()

		if score >= beta {
			e.killer.Add(e.ply, move)
			return beta
		}

		alpha = max(score, alpha)
	}

	return alpha
}

// Gets the principal variation by recursively following the best moves in the
// transposition table.
func (e *Engine) PrincipalVariation(bestMove dragon.Move, depth int) []dragon.Move {
	if depth == 0 || !e.legal(bestMove) {
		return nil
	}

	unmove := e.Move(bestMove)
	defer unmove()

	if nextMove, ok := e.PVMove(); ok {
		return append(
			[]dragon.Move{bestMove},
			e.PrincipalVariation(nextMove, depth-1)...,
		)
	}
	return []dragon.Move{bestMove}
}

func (e *Engine) PVMove() (dragon.Move, bool) {
	entry, ok := e.transpositions.Get(e.board.Hash())
	return entry.best, ok
}

// GenMoves generates legal moves and reports whether the side to move is in check.
func (e *Engine) GenMoves() ([]dragon.Move, bool) {
	if e.Draw() {
		return nil, false
	}
	return e.board.GenerateLegalMoves()
}

func (e *Engine) legal(move dragon.Move) bool {
	moves, _ := e.GenMoves()
	for _, m := range moves {
		if m == move {
			return true
		}
	}
	return false
}

// terminalMove checks if a move is terminal and gets the score at terminal nodes.
func (e *Engine) terminalMove(move dragon.Move) (int16, bool) {
	unmove := e.Move(move)
	defer unmove()

	moves, inCheck := e.GenMoves()

	terminalScore, ok := e.terminal(len(moves), inCheck)
	return -terminalScore, ok
}

func (e *Engine) terminal(numMoves int, inCheck bool) (int16, bool) {
	if numMoves > 0 {
		return 0, false
	}

	if inCheck {
		return mateVal + e.ply, true
	}
	return drawVal, true
}

func whiteToMove(board *dragon.Board) int16 {
	if board.Wtomove {
		return 1
	}
	return -1
}
