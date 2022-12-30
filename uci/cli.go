// Package uci implements the Universal Chess Interface protocol.
// See http://wbec-ridderkerk.nl/html/UCIProtocol.html.
package uci

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

const startingFen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

type Engine interface {
	About() (name string, author string, version string)

	NewGame()
	Position(fen string, moves []string)
	Go(info SearchParams) SearchResults
	Stop()

	// IsReady should block until the engine is ready to search.
	IsReady()
	SetOption(option string, value string)
	Debug(isOn bool)
}

func Run(engine Engine) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		if input == "quit" {
			return
		}

		if err := handle(engine, input); err != nil {
			log.Fatal(err)
		}
	}
}

func handle(engine Engine, input string) error {
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return nil
	}

	cmd, args := fields[0], fields[1:]

	switch cmd {
	case "uci":
		name, author, verison := engine.About()
		fmt.Printf("id name %s %s\n", name, verison)
		fmt.Printf("id author %s\n", author)
		fmt.Println("uciok")
	case "setoption":
		if len(args) < 4 {
			log.Printf("Not enough args to setoption: want 4 args, got %v\n", args)
			return nil
		}
		engine.SetOption(args[1], args[3])
	case "ucinewgame":
		engine.NewGame()
	case "isready":
		engine.IsReady()
		fmt.Println("readyok")
	case "position":
		fen := args[0]
		if fen == "startpos" {
			fen = startingFen
		}
		var moves []string
		if len(args) >= 2 {
			moves = args[2:]
		}
		engine.Position(fen, moves)
	case "go":
		search(engine, args)
	case "stop":
		engine.Stop()
	case "ponderhit":
	case "debug":
	}

	return nil
}
