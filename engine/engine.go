package engine

import (
	"context"
	"log"

	"github.com/noahklein/chess/uci"
	"github.com/noahklein/dragon"
)

const (
	name    = "Swindle"
	author  = "Noah Klein"
	version = "1.0"

	depth = 6
)

// The chess engine. Must call NewGame() to initialize, followed by Position().
type Engine struct {
	board *dragon.Board

	ply            int16
	transpositions *Transpositions
	killer         *Killer
	history        *History

	nodeCount NodeCount
	debug     bool // Enables logs/metrics.
	cancel    func()
}

func (e *Engine) About() (string, string, string) {
	return name, author, version
}

func (e *Engine) NewGame() {
	board := dragon.ParseFen(dragon.Startpos)
	e.killer = NewKiller()
	e.board = &board
	e.transpositions = NewTable()
	e.nodeCount = NodeCount{}
	e.history = &History{}
	e.ply = 1
	e.cancel = func() {}
	e.debug = true
}

// Copy an engine for concurrent search.
func (e *Engine) Copy() *Engine {
	board := *e.board

	return &Engine{
		killer:         e.killer,
		transpositions: e.transpositions,
		board:          &board,
		history:        e.history.Copy(),
		ply:            e.ply,
		cancel:         e.cancel,
		debug:          e.debug,
	}
}

func (e *Engine) Position(fen string, moves []string) {
	board := dragon.ParseFen(fen)
	e.board = &board
	e.ply = int16(e.board.Fullmoveno * 2)
	e.history.Push(board.Hash())
	for _, move := range moves {
		m, err := dragon.ParseMove(move)
		if err != nil {
			log.Fatalf("Could not parse move %v: %v", move, err)
		}
		e.Move(m)
	}

	e.Warn("Position set")
}

// Go is the search entry-point, called by the UCI go command.
func (e *Engine) Go(info uci.SearchParams) uci.SearchResults {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	e.cancel = cancel

	if info.Depth == 0 {
		info.Depth = depth
	}
	if info.Infinite {
		info.Depth = 100
	}

	return e.Search(ctx, info)
}

// Make a move on the board. Returns an unmove callback.
func (e *Engine) Move(m dragon.Move) func() {
	unapply := e.board.Apply(m)
	e.history.Push(e.board.Hash())
	e.ply++
	e.nodeCount.Ply(e.ply)

	return func() {
		unapply()
		e.ply--
		e.history.Pop()
	}
}

// Draw checks for threefold repetitions.
func (e *Engine) Draw() bool {
	return e.history.Draw(e.board.Hash(), e.ply, e.board.Halfmoveclock)
}

func (e *Engine) Stop() {
	e.cancel()
}

// IsReady should block until the engine is ready to search.
func (e *Engine) IsReady() {}
