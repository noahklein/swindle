package engine

import (
	"context"
	"fmt"
	"log"

	"github.com/noahklein/chess/uci"
	"github.com/noahklein/dragon"
)

const (
	name    = "Cheese"
	author  = "Noah Klein"
	version = "1.0"

	depth = 4
)

// The chess engine. Must call NewGame() to initialize, followed by Position().
type Engine struct {
	killer *Killer
	board  *dragon.Board
	ply    int
	cancel func()

	debug bool // Enables logs/metrics.
}

func (e *Engine) About() (string, string, string) {
	return name, author, version
}

func (e *Engine) NewGame() {
	board := dragon.ParseFen(dragon.Startpos)
	e.killer = NewKiller()
	e.board = &board
	e.ply = 1
	e.cancel = func() {}
	e.debug = true
}

func (e *Engine) Position(fen string, moves []string) {
	board := dragon.ParseFen(fen)
	e.board = &board
	for _, move := range moves {
		m, err := dragon.ParseMove(move)
		if err != nil {
			log.Fatalf("Could not parse move %v: %v", move, err)
		}
		e.Move(m)
	}
}

func (e *Engine) Go(info uci.SearchParams) uci.SearchResults {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	e.cancel = cancel

	if info.Depth == 0 {
		info.Depth = depth
	}

	return e.Search(ctx, info)
}

func (e *Engine) Move(m dragon.Move) func() {
	unapply := e.board.Apply(m)
	e.ply++

	return func() {
		unapply()
		e.ply--
	}
}

func (e *Engine) Stop() {
	e.cancel()
}

// IsReady should block until the engine is ready to search.
func (e *Engine) IsReady() {}

func (e *Engine) SetOption(option string, value string) {
	// TODO: Implement.
}

// Debug enables logging and metric reporting.
func (e *Engine) Debug(isOn bool) {
	e.debug = isOn
}

func (e *Engine) Print(s string, a ...any) {
	if e.debug {
		fmt.Printf("info string "+s+"\n", a...)
	}
}
