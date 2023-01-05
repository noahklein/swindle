package engine

import (
	"context"
	"fmt"
	"log"

	"github.com/fatih/color"
	"github.com/noahklein/chess/uci"
	"github.com/noahklein/dragon"
)

const (
	name    = "Cheese"
	author  = "Noah Klein"
	version = "1.0"

	depth = 5
)

// The chess engine. Must call NewGame() to initialize, followed by Position().
type Engine struct {
	killer  *Killer
	table   *Table
	board   *dragon.Board
	history map[uint64]int
	ply     int
	cancel  func()

	debug bool // Enables logs/metrics.
}

func (e *Engine) About() (string, string, string) {
	return name, author, version
}

func (e *Engine) NewGame() {
	board := dragon.ParseFen(dragon.Startpos)
	e.killer = NewKiller()
	e.board = &board
	e.table = NewTable()
	e.history = map[uint64]int{}
	e.ply = 1
	e.cancel = func() {}
	e.debug = true
}

func (e *Engine) Position(fen string, moves []string) {
	board := dragon.ParseFen(fen)
	e.board = &board
	e.history = map[uint64]int{board.Hash(): 1}
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

// Make a move on the board. Returns an unmove callback.
func (e *Engine) Move(m dragon.Move) func() {
	unapply := e.board.Apply(m)
	e.ply++
	hash := e.board.Hash()
	e.history[hash]++

	return func() {
		unapply()
		e.ply--
		e.history[hash]--
	}
}

func (e *Engine) Threefold() bool {
	return e.history[e.board.Hash()] >= 3
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

func (e *Engine) Error(s string, a ...any) {
	color.Red(s, a...)
}
