package engine

import (
	"context"
	"fmt"
	"log"

	"github.com/dylhunn/dragontoothmg"
	"github.com/noahklein/chess/uci"
)

type Engine struct {
	board  dragontoothmg.Board
	cancel func()
}

func (e *Engine) About() (string, string, string) {
	return "Cheese", "Noah Klein", "1.0"
}

func (e *Engine) NewGame() {
	e.board = dragontoothmg.ParseFen(dragontoothmg.Startpos)
	e.cancel = func() {}
}

func (e *Engine) Position(fen string, moves []string) {
	e.board = dragontoothmg.ParseFen(fen)
	for _, move := range moves {
		m, err := dragontoothmg.ParseMove(move)
		if err != nil {
			log.Fatalf("Could not parse move %v: %v", move, err)
		}
		e.board.Apply(m)
	}
}

func (e *Engine) Go(info uci.SearchParams) uci.SearchResults {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if info.Depth == 0 {
		info.Depth = 5
	}

	return e.Search(ctx, info)
}

func (e *Engine) Stop() {
	e.cancel()
}

// IsReady should block until the engine is ready to search.
func (e *Engine) IsReady() {}

func (e *Engine) SetOption(option string, value string) {
	// TODO: Implement.
}

func (e *Engine) Debug(isOn bool) {
	// TODO: Implement
}

func (e *Engine) Print(s string, a ...any) {
	fmt.Printf("info string "+s+"\n", a...)
}
