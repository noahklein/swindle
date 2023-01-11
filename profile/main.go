// Plays games at multiple depths and reports stats. To get a profile, use the --profile
// flag and run: go tool pprof -top http://localhost:6060/debug/pprof/profile
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/noahklein/chess/engine"
	"github.com/noahklein/chess/uci"
	"github.com/noahklein/dragon"

	"github.com/fatih/color"
	"github.com/rodaine/table"

	_ "net/http/pprof"
)

var (
	maxDepth  = flag.Int("depth", 3, "Max depth to search")
	thinkTime = flag.Duration("think", 1*time.Second, "How long to think each move")
	profile   = flag.Bool("profile", false, "Enables pprof")
)

func init() {
	flag.Parse()
}

func main() {
	if *profile {
		fmt.Println(`To get a profile run:
    go tool pprof -top http://localhost:6060/debug/pprof/profile
	`)
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	fmt.Println("Playing full games at increasing depths:")

	var (
		headerColor   = color.New(color.FgGreen, color.Underline).SprintfFunc()
		firstColColor = color.New(color.FgYellow).SprintfFunc()
	)
	tbl := table.New("Depth", "Score", "Moves", "Time", "Hashfull", "Nodes", "NPS")
	tbl.WithHeaderFormatter(headerColor)
	tbl.WithFirstColumnFormatter(firstColColor)

	// Play a whole game at a certain depth and log the results in a table.
	for depth := 1; depth <= *maxDepth; depth++ {
		start := time.Now()
		results, moveCount := playGame(dragon.Startpos, depth, *thinkTime)

		nps := results.Nodes * int(time.Second) / int(time.Since(start))

		tbl.AddRow(
			depth, results.Score, moveCount,
			fmtDuration(time.Since(start)),
			results.Hashfull,
			results.Nodes, fmt.Sprintf("%v/s", nps),
		)
		color.Green("Finished depth %v in %v", depth, time.Since(start))

		time.Sleep(1 * time.Second)
	}

	fmt.Print("==== Finished ====\n\n")
	tbl.Print()
}

func playGame(fen string, depth int, thinkTime time.Duration) (uci.SearchResults, int) {
	e := &engine.Engine{}
	e.NewGame()
	e.Position(fen, nil)
	e.Debug(false)

	var moveCount int
	var finalResults uci.SearchResults
	for moves, _ := e.GenMoves(); len(moves) > 0; moves, _ = e.GenMoves() {
		moveCount++

		ctx, cancel := context.WithTimeout(context.Background(), thinkTime)
		results := e.IterDeep(ctx, uci.SearchParams{
			Depth: depth,
		})
		cancel()

		move, err := dragon.ParseMove(results.Move)
		if err != nil {
			panic(err)
		}
		e.Move(move)

		results.Nodes += finalResults.Nodes
		finalResults = results

		if moveCount > 30 {
			break
		}
	}
	return finalResults, moveCount
}

func fmtDuration(d time.Duration) string {
	scale := 10 * time.Second
	// Look for the max scale that is smaller than d.
	for scale > d {
		scale /= 10
	}
	return d.Round(scale / 100).String()
}
