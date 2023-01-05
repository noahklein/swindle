package uci

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type SearchParams struct {
	// Clock info.
	WhiteTime time.Duration
	BlackTime time.Duration
	WhiteInc  time.Duration
	BlackInc  time.Duration

	Depth    int
	Infinite bool
	moveTime time.Duration
}

type SearchResults struct {
	BestMove string
	Score    int16
	Mate     int16    // Moves till mate, negative if engine is losing.
	PV       []string // Principal variation.
	Nodes    int
	Depth    int
}

func search(engine Engine, args []string) {
	params := parseParams(args)
	go func() {
		start := time.Now()
		result := engine.Go(params)
		duration := time.Since(start)
		// TODO: report other search results.
		if result.Mate != 0 {
			fmt.Printf("info depth %v score mate %v time %d nodes %v pv %v\n", result.Depth, result.Mate, duration/time.Millisecond, result.Nodes, strings.Join(result.PV, " "))
		}
		fmt.Printf("info depth %v score cp %v time %d nodes %v pv %v\n", result.Depth, result.Score, duration/time.Millisecond, result.Nodes, strings.Join(result.PV, " "))
		fmt.Printf("bestmove %v\n", result.BestMove)
	}()
}

func parseParams(args []string) SearchParams {
	var sp SearchParams
	for i := 0; i < len(args); i++ {
		param := args[i]
		switch param {
		case "wtime":
			sp.WhiteTime = parseMs(args[i+1])
			i++
		case "btime":
			sp.BlackTime = parseMs(args[i+1])
			i++
		case "winc":
			sp.WhiteInc = parseMs(args[i+1])
			i++
		case "binc":
			sp.BlackInc = parseMs(args[i+1])
			i++
		case "movetime":
			sp.moveTime = parseMs(args[i+1])
			i++
		case "depth":
			sp.Depth, _ = strconv.Atoi(args[i+1])
			i++
		case "infinite":
			sp.Infinite = true
		}
	}

	return sp
}

// Converts a string number, e.g. "3000", into a duration in milliseconds.
func parseMs(s string) time.Duration {
	ms, err := time.ParseDuration(s + "ms")
	if err != nil {
		log.Printf("Failed to parse uci duration: %v\n", err)
	}
	return ms
}
