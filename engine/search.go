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
	const window = pawnVal / 4
	alpha, beta := -infinity, infinity
	exp := 1 // Exponentially increase window on window misses.

	if entry, ok := e.transpositions.Get(e.board.Hash(), e.ply); ok {
		alpha, beta = entry.value-window, entry.value+window
	}

	moves, _ := e.GenMoves()
	if len(moves) == 0 {
		panic("IterDeep called with no moves")
	}

	var nodes int
	start := time.Now()
	// Default best move is first move in case of timeout before first iteration.
	bestResult := uci.SearchResults{Move: moves[0].String()}
	for depth := int16(1); ctx.Err() == nil && depth <= int16(params.Depth); {
		e.Warn("depth=%v, ab: %v, %v", depth, alpha, beta)
		result := e.Search(ctx, depth, alpha, beta)
		score := result.Score
		nodes += result.Nodes
		result.Nodes = nodes

		if ctx.Err() != nil {
			e.Warn("timeout")
			bestResult.Nodes = nodes
			return bestResult
		}

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
		bestResult = result
	}

	return bestResult
}

// Root search. Runs AlphaBeta search concurrently for each move and collects the results.
func (e *Engine) Search(ctx context.Context, depth, alpha, beta int16) uci.SearchResults {
	moves, _ := e.GenMoves()

	if e.threads == 0 || e.threads > len(moves) {
		e.threads = len(moves)
	}
	e.threads = len(moves)

	moveCh := make(chan dragon.Move, len(moves))
	for _, move := range moves {
		moveCh <- move
	}
	close(moveCh)

	resultCh := make(chan uci.SearchResults, len(moves))

	var wg sync.WaitGroup
	for t := 0; t < e.threads; t++ {
		e := e.Copy()
		wg.Add(1)
		go func() {
			defer wg.Done()

			for move := range moveCh {
				e.nodeCount.Inc()

				// Search this move.
				unmove := e.Move(move)
				score := -e.AlphaBeta(ctx, -beta, -alpha, int(depth))
				unmove()

				var pv []string
				for _, m := range e.PrincipalVariation(move, 10) {
					pv = append(pv, m.String())
				}

				resultCh <- uci.SearchResults{
					Move:           move.String(),
					Score:          score,
					Mate:           mateScore(score, e.ply),
					Nodes:          e.nodeCount.nodes,
					PV:             pv,
					SelectiveDepth: int(e.nodeCount.maxPly - e.ply),
				}
			}
		}()
	}

	wg.Wait()
	close(resultCh)

	best := uci.SearchResults{
		Score: -infinity, Mate: NotMate,
		Move: moves[0].String(), // Init to first move in case of immediate cancellation.

		Depth:     int(depth),
		Hashfull:  e.transpositions.PermillFull(),
		TableHits: e.transpositions.hits,
	}

	for result := range resultCh {
		best.Nodes += result.Nodes
		best.SelectiveDepth = int(max(int16(best.SelectiveDepth), int16(result.SelectiveDepth)))

		if result.Score >= best.Score {
			best.Score = result.Score
			best.Mate = result.Mate
			best.Move = result.Move
			best.PV = result.PV
		}
	}

	return best
}

// AlphaBeta improves upon the minimax algorithm.
//     Alpha is the lowest score the maximizing player can force
//     Beta is the highest score the minimizing player can force.
// It stops evaluating a move when at least one possibility has been found that
// proves the move to be worse than a previously examined move. In other words,
// you only need one refutation to know a move is bad.
func (e *Engine) AlphaBeta(ctx context.Context, alpha, beta int16, depth int) int16 {
	e.nodeCount.Inc()
	// Only do forward-pruning techniques in zero-window search.
	pvNode := alpha != beta-1

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

	if len(moves) == 1 || inCheck {
		// Only one reply, this ply is free. Extend search.
		// TODO: constrain this.
		depth++
	}

	// Check transposition table.
	if val, nt := e.transpositions.GetEval(e.board.Hash(), depth, alpha, beta, e.ply); nt != NodeUnknown {
		return val
	}

	if depth <= 0 || ctx.Err() != nil {
		return e.Quiesce(alpha, beta)
	}

	// Null-move pruning.
	// TODO: skip this in endgames.
	if !e.disableNullMove && !pvNode && !inCheck && depth >= 3 {
		if score := e.searchNullMove(ctx, -beta, depth); score >= beta {
			return beta
		}
	}
	// Futility pruning, if the current eval is hopelessly lower than alpha and
	// we're close to the horizon, skip to quiescence search.
	if !pvNode && !inCheck && depth < 3 {
		eval := Eval(e.board)
		if depth == 1 && eval+knightVal < alpha {
			return e.Quiesce(alpha, beta)
		}
		if depth == 2 && eval+rookVal < alpha {
			return e.Quiesce(alpha, beta)
		}
	}

	// e.sortMoves(moves)
	moveSorter := e.newMoveSorter(moves)

	// Assume this is an alpha node.
	nodeType := NodeAlpha
	var bestMove dragon.Move
	for mNum := range moves {
		move := moveSorter.Next(mNum)
		unmove := e.Move(move)

		// Only search the first 6 sorted moves to full depth.
		lateMoveReduction := 1
		if mNum > 6 && depth >= 3 {
			lateMoveReduction = depth / 3
		}
		// Extend or reduce search depth for this move.
		moveDepth := depth - lateMoveReduction

		var score int16
		if nodeType == NodeExact {
			// Zero-window search.
			score = -e.AlphaBeta(ctx, -alpha-1, -alpha, moveDepth)
			// If we failed, search again with normal window.
			if score > alpha && score < beta {
				score = -e.AlphaBeta(ctx, -beta, -alpha, moveDepth)
			}
		} else {
			score = -e.AlphaBeta(ctx, -beta, -alpha, moveDepth)
		}
		unmove()

		// Beta-cutoff; better than the previous best move, opponent won't allow this.
		if score >= beta {
			e.killer.Add(e.ply, move)
			e.transpositions.Add(e.ply, Entry{
				key:   e.board.Hash(),
				depth: moveDepth,
				flag:  NodeBeta,
				value: beta,
				best:  move,
			})
			return beta
		}
		if score > alpha {
			alpha = score
			bestMove = move
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

		victim, _ := e.squares.PieceType(move.To())
		// Skip non-captures.
		if victim == dragon.Nothing {
			continue
		}
		// Delta cutoff, this is hopeless.
		if score+PieceValue[victim]+200 < alpha {
			continue
		}

		attacker, _ := e.squares.PieceType(move.From())
		// Skip captures that lose material.
		if badCapture(attacker, victim) {
			continue
		}

		loudMoves = append(loudMoves, move)
	}

	moveSorter := e.newMoveSorter(loudMoves)
	for mNum := range loudMoves {
		move := moveSorter.Next(mNum)
		unmove := e.Move(move)
		score := -e.Quiesce(-beta, -alpha)
		unmove()

		if score >= beta {
			// e.killer.Add(e.ply, move)
			return beta
		}

		alpha = max(score, alpha)
	}

	return alpha
}

// Pass turn and do a zero-window search at reduced depth, if opponent still has no
// good moves, prune this node. Note: this causes issues if we're in zugzwang.
func (e *Engine) searchNullMove(ctx context.Context, beta int16, depth int) int16 {
	r := 2
	if depth >= 6 {
		r = 4 + depth/6
	}

	undoNull := e.board.NullMove()
	score := -e.AlphaBeta(ctx, -beta, -beta+1, depth-r)
	undoNull()

	return score
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
	entry, ok := e.transpositions.Get(e.board.Hash(), e.ply)
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
