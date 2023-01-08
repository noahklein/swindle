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
	Move     string
	Score    float32
	Mate     int16    // Moves till mate, negative if engine is losing.
	PV       []string // Principal variation.
	Hashfull int      // The hash is x permill full.
	Nodes    int

	Depth, SelectiveDepth int
}

func (sr SearchResults) Print(duration time.Duration) string {
	var b strings.Builder
	b.WriteString("info ")

	add := func(str string, a ...any) {
		b.WriteString(fmt.Sprintf(str+" ", a...))
	}

	add("depth %d", sr.Depth)
	add("seldepth %d", sr.SelectiveDepth)
	if sr.Mate != 0 {
		add("score mate %v", sr.Mate)
	} else {
		add("score cp %v", sr.Score)
	}
	add("hashfull %d", sr.Hashfull)
	add("time %d", duration/time.Millisecond)
	add("nodes %d", sr.Nodes)

	if len(sr.PV) > 0 {
		add("pv %v", strings.Join(sr.PV, " "))
	}

	return b.String()
}

func search(engine Engine, args []string) {
	params := parseParams(args)
	go func() {
		start := time.Now()
		result := engine.Go(params)
		duration := time.Since(start)

		fmt.Println(result.Print(duration))
		fmt.Printf("bestmove %v\n", result.Move)
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
