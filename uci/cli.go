// Package uci implements the Universal Chess Interface protocol.
// See http://wbec-ridderkerk.nl/html/UCIProtocol.html.
package uci

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/noahklein/chess/log"
)

const startingFen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

var errExit = errors.New("exit command received")

type Engine interface {
	About() (name string, author string, version string)

	NewGame()
	Position(fen string, moves []string)
	Go(info SearchParams) SearchResults
	Stop()

	// IsReady should block until the engine is ready to search.
	IsReady()
	SetOption(option string, value string) error
	Debug(isOn bool)
	ClearTT()
}

func Run(engine Engine) {
	name, author, version := engine.About()
	log.Green(`%s v%s by %s`, name, version, author)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		if input == "quit" {
			return
		}

		if err := handle(engine, input); err == errExit {
			return
		} else if err != nil {
			panic(err)
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
			fmt.Println("Not enough args to setoption: want 4 args, got", args)
			return nil
		}
		if err := engine.SetOption(args[1], args[3]); err != nil {
			fmt.Println("Failed to set option:", err)
		}
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
	case "exit":
		return errExit

		// Custom commands for debugging.
	case "start":
		handle(engine, "uci")
		handle(engine, "ucinewgame")
		handle(engine, "position startpos moves e2e4")
	case "cleartt":
		engine.ClearTT()
	case "help":
		printHelp(engine)
	}

	return nil
}

const helpString = `This is a UCI-compatible chess engine.
For a full list of comamnds, read the UCI protocol:
	http://wbec-ridderkerk.nl/html/UCIProtocol.html

Set-up a position:
	ucinewgame
	position startpos moves e2e4 e7e5

To search a position at depth 10:
	go depth 10

Infinite search:
	go infinite
	stop

Search can be cancelled with the stop command.
`

func printHelp(e Engine) {
	fmt.Print(helpString)
}
