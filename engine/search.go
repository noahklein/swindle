package engine

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/noahklein/chess/uci"
	"github.com/noahklein/dragon"
)

const (
	mateVal      int16 = -15000
	drawVal      int16 = 0
	initialAlpha int16 = -20000
	initialBeta  int16 = 20000
)

// Root search.
func (e *Engine) Search(ctx context.Context, params uci.SearchParams) uci.SearchResults {
	e.nodeCount.Reset()

	// TODO: Smarter time management; look at remaining clock.
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	moves, _ := e.GenMoves()
	if len(moves) == 0 {
		e.Error("Search() called on game that has already ended.")
		return uci.SearchResults{}
	}

	var (
		bestMu    sync.Mutex
		bestScore = initialAlpha
		bestMove  = moves[0]
		// TODO: investigate mysterious node counts.
		nodes, qNodes int
		maxDepth      int16
	)

	var wg sync.WaitGroup
	for _, move := range moves {
		// Capture variables for goroutine.
		move, e := move, e.Copy()

		wg.Add(1)
		go func() {
			defer wg.Done()
			e.nodeCount.Inc()

			unmove := e.Move(move)
			defer unmove()
			score := e.IterDeep(ctx, params.Depth)

			bestMu.Lock()
			defer bestMu.Unlock()
			if score >= bestScore {
				e.Print("move %s: %v", move.String(), score)
				bestScore = score
				bestMove = move

				nodes += e.nodeCount.nodes
				qNodes += e.nodeCount.qNodes
				maxDepth = max(maxDepth, e.nodeCount.maxPly-e.ply)
			}
		}()
	}
	wg.Wait()

	var pv []string
	for _, move := range e.PrincipalVariation(bestMove, 5) {
		pv = append(pv, move.String())
	}

	return uci.SearchResults{
		BestMove: bestMove.String(),
		Score:    float32(bestScore) / 100,
		Mate:     e.mateScore(bestScore),
		Nodes:    int(nodes),
		PV:       pv,
		Depth:    int(maxDepth),
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
	e.nodeCount.Inc()

	if e.Threefold() {
		return drawVal
	}

	moves, inCheck := e.GenMoves()
	if inCheck && len(moves) == 0 {
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
		return e.Quiesce(alpha, beta)
	}

	if len(moves) == 1 {
		// TODO: constrain this.
		// Only one reply, this ply is free. Extend search.
		// depth++
	}

	// Assume this is an alpha node.
	nodeType := NodeAlpha

	var foundPV bool
	var bestMove dragon.Move
	for mNum, move := range e.sortMoves(moves) {
		unmove := e.Move(move)

		reduction := 1
		if mNum > 10 {
			reduction = 2
		}

		var score int16
		if foundPV {
			// Search tiny window.
			score = -e.AlphaBeta(-alpha-1, -alpha, depth-reduction)
			// If we failed, search again with normal window.
			if score > alpha && score < beta {
				score = -e.AlphaBeta(-beta, -alpha, depth-reduction)
			}
		} else {
			score = -e.AlphaBeta(-beta, -alpha, depth-reduction)
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
// Quiescent search avoids the "horizon effect."
// Note: 50%-90% of nodes searched are here, pruning goes a long way.
func (e *Engine) Quiesce(alpha, beta int16) int16 {
	if e.Threefold() {
		return drawVal
	}

	e.nodeCount.Qinc()
	score := Eval(e.board)
	if score >= beta {
		return beta
	}
	if alpha < score {
		alpha = score
	}

	// TODO: handle checks specially.
	moves, _ := e.GenMoves()
	for _, move := range moves {
		if score, ok := e.Terminal(move); ok {
			return -score
		}

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
	if depth == 0 || !e.legal(bestMove) {
		return nil
	}

	unmove := e.Move(bestMove)
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

// GenMoves generates legal moves and reports whether the side to move is in check.
func (e *Engine) GenMoves() ([]dragon.Move, bool) {
	if e.Threefold() {
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

// Terminal gets the score at terminal nodes.
func (e *Engine) Terminal(move dragon.Move) (int16, bool) {
	unmove := e.Move(move)
	defer unmove()

	moves, inCheck := e.GenMoves()
	if len(moves) > 0 {
		return 0, false
	}

	if inCheck {
		return -mateVal - e.ply, true
	}
	return drawVal, true
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

const maxMate = 400

// Converts an eval score into ply till mate. Returns 0 if not mating.
func (e *Engine) mateScore(score int16) int16 {
	var mate int16
	plyTillMate := abs(mateVal) - abs(score) - e.ply
	if plyTillMate < maxMate {
		mate = plyTillMate / 2

		if score < 0 {
			mate *= -1
		}
	}

	return mate
}
