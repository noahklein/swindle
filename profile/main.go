package main

import (
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
	maxDepth = flag.Int("depth", 3, "Max depth to search")
	profile  = flag.Bool("profile", false, "Enables pprof")
)

func init() {
	flag.Parse()
}

func main() {
	if *profile {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	var (
		headerColor   = color.New(color.FgGreen, color.Underline).SprintfFunc()
		firstColColor = color.New(color.FgYellow).SprintfFunc()
	)
	tbl := table.New("Depth", "Score", "Time", "Nodes", "NPS")
	tbl.WithHeaderFormatter(headerColor)
	tbl.WithFirstColumnFormatter(firstColColor)

	for depth := 1; depth <= *maxDepth; depth++ {
		start := time.Now()
		results := playGame(dragon.Startpos, depth)

		tbl.AddRow(
			depth, results.Score, fmtDuration(time.Since(start)),
			results.Nodes,
			fmt.Sprintf("%v/s", results.Nodes/int(time.Since(start)/time.Second)),
		)
		fmt.Printf("Finished depth %v in %v\n", depth, time.Since(start))
		time.Sleep(1 * time.Second)
	}

	fmt.Print("==== Finished ====\n\n")
	tbl.Print()
}

func playGame(fen string, depth int) uci.SearchResults {
	e := &engine.Engine{}
	e.NewGame()
	e.Position(fen, nil)
	e.Debug(false)

	var finalResults uci.SearchResults
	for len(e.GenMoves()) > 0 {
		results := e.Go(uci.SearchParams{
			Depth: depth,
		})

		move, err := dragon.ParseMove(results.BestMove)
		if err != nil {
			panic(err)
		}
		e.Move(move)

		results.Nodes += finalResults.Nodes
		finalResults = results
	}
	return finalResults
}

func fmtDuration(d time.Duration) string {
	scale := 10 * time.Second
	// Look for the max scale that is smaller than d.
	for scale > d {
		scale /= 10
	}
	return d.Round(scale / 100).String()
}
