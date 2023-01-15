// Plays games at multiple depths and reports stats. To get a profile, use the --profile
// flag and run: go tool pprof -top http://localhost:6060/debug/pprof/profile
package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/noahklein/chess/engine"
	"github.com/noahklein/chess/log"
	puzzledb "github.com/noahklein/chess/puzzle"
	"github.com/noahklein/chess/uci"
	"github.com/noahklein/dragon"

	"github.com/rodaine/table"

	_ "net/http/pprof"
)

var (
	maxDepth  = flag.Int("depth", 3, "Max depth to search")
	thinkTime = flag.Duration("think", 30*time.Second, "How long to think each move")
	ttd       = flag.Int("ttd", 0, "Time-till-depth; report how long it takes to reach a given depth in non-mate puzzles")
	profile   = flag.Bool("profile", false, "Enables pprof")
	v         = flag.Int("v", -1, "verbose")
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
			fmt.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	if *ttd > 0 {
		puzzleDepth(*ttd)
		return
	}

	fmt.Println("Playing full games at increasing depths:")

	var ()
	tbl := table.New("Depth", "Score", "Moves", "Time", "Hashfull", "KNodes", "NPS")

	// Play a whole game at a certain depth and log the results in a table.
	for depth := *maxDepth; depth <= *maxDepth; depth++ {
		start := time.Now()
		results, moveCount := playGame(dragon.Startpos, depth, *thinkTime)

		nps := results.Nodes * int(time.Second) / int(time.Since(start))

		tbl.AddRow(
			depth, results.Score, moveCount,
			fmtDuration(time.Since(start)),
			results.Hashfull,
			results.Nodes/1000, fmt.Sprintf("%vk/s", nps/1000),
		)
		log.Green("Finished depth %v in %v", depth, time.Since(start))

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
	e.SetOption("hash", "128")
	e.Level = log.Level(*v)

	fmt.Print("Depth: ", depth, "")

	var moveCount int
	var finalResults uci.SearchResults
	for moves, _ := e.GenMoves(); len(moves) > 0; moves, _ = e.GenMoves() {
		moveCount++
		fmt.Print(".")

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

		if moveCount == 30 {
			break
		}
	}
	fmt.Println()
	return finalResults, moveCount
}

func puzzleDepth(depth int) {
	rand.Seed(time.Now().Unix())
	puzzles := puzzledb.Query(10, func(p puzzledb.Puzzle) bool {
		for _, v := range p.Themes {
			if v == "mate" {
				return false
			}
		}
		return true
	})

	var elapsed time.Duration
	for _, p := range puzzles {
		elapsed += timeTillDepth(p.Fen, depth)
	}

	fmt.Println()
	log.Green("Done")
	fmt.Println("\tTotal:", elapsed)
	fmt.Println("\tAvg:  ", elapsed/time.Duration(len(puzzles)))
}

func timeTillDepth(fen string, depth int) time.Duration {
	var e engine.Engine
	e.NewGame()
	e.Position(fen, nil)
	e.Level = log.Level(*v)

	start := time.Now()

	ctx := context.Background() // No time limit.
	results := e.IterDeep(ctx, uci.SearchParams{
		Depth: depth,
	})
	fmt.Println(results.Print(start))
	elapsed := time.Since(start)

	colorFn := log.Green
	if elapsed > 20*time.Second {
		colorFn = log.Red
	}
	colorFn("Elapsed: %v", time.Since(start))
	fmt.Println()

	return elapsed
}

func fmtDuration(d time.Duration) string {
	scale := 10 * time.Second
	// Look for the max scale that is smaller than d.
	for scale > d {
		scale /= 10
	}
	return d.Round(scale / 100).String()
}
