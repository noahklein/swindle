// Test the engine against the Lichess puzzle db.
package main

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/noahklein/chess/elo"
	"github.com/noahklein/chess/engine"
	"github.com/noahklein/chess/log"
	puzzledb "github.com/noahklein/chess/puzzle"
	"github.com/noahklein/chess/uci"
	"github.com/noahklein/dragon"
)

var (
	tag    = flag.String("tag", "", "tag filter")
	id     = flag.String("id", "", "comma-seperated list of ids")
	limit  = flag.Int("limit", 10, "limit the number of puzzles; 0 for all (not recommended)")
	depth  = flag.Int("depth", 30, "depth to search puzzles at")
	think  = flag.Duration("think", 5*time.Second, "how long to think")
	rating = flag.Int("rating", 0, "min rating of puzzles")
	length = flag.Int("length", 0, "puzzle length filter, must be even; 0 for all")

	tsearch = flag.String("tsearch", "", "search for tags")
	verbose = flag.Int("v", 0, "log level, -1 to disable logging")
)

func main() {
	flag.Parse()

	if *tsearch != "" {
		tagSearch(*tsearch)
		return
	}

	var ids []string
	if len(*id) > 0 {
		ids = strings.Split(*id, ",")
	}
	var puzzles = puzzledb.Query(*limit, func(p puzzledb.Puzzle) bool {
		if len(ids) > 0 && !contains(ids, p.ID) {
			return false
		}
		if *tag != "" && !contains(p.Themes, *tag) {
			return false
		}
		if *length != 0 && len(p.Moves) != *length {
			return false
		}
		if p.Rating < *rating {
			return false
		}
		return true
	})

	fmt.Printf("Attempting %v puzzles\n", len(puzzles))

	correct := len(puzzles)
	start := time.Now()

	rating := elo.Default
	var failedIDs []string

	var e engine.Engine

	for pNum, p := range puzzles {
		e.NewGame()
		e.Position(p.Fen, nil)
		e.Level = log.Level(*verbose)

		var failed bool
		var movesCompleted string
		for i := 0; i < len(p.Moves); i += 2 {
			e.Position(p.Fen, p.Moves[:i+1])
			want := p.Moves[i+1]

			ctx, cancel := context.WithTimeout(context.Background(), *think)
			result := e.IterDeep(ctx, uci.SearchParams{
				Depth: *depth,
			})
			cancel()

			if result.Move != want {
				failed = true

				// Check mate (alternate solution.)
				m, err := dragon.ParseMove(result.Move)
				if err != nil {
					panic(err)
				}
				e.Move(m)

				moves, inCheck := e.GenMoves()
				if inCheck && len(moves) == 0 {
					movesCompleted += "."
					log.Yellow("%6d) Passed %s %s (alternate solution)", pNum+1, p.ID, movesCompleted)
					break
				}

				correct--
				movesCompleted += "x"
				log.Red("%6d) Failed %s %s", pNum+1, p.ID, movesCompleted)
				log.Red(puzzledb.LichessUrl(p.Fen))
				log.Red("%v %v", p.Rating, strings.Join(p.Themes, ", "))
				log.Red("Wrong move: got %v, want %v, %v", result.Move, want, p.Moves)
				log.Red(result.Print(start))

				failed = true
				failedIDs = append(failedIDs, p.ID)
				break
			}
			movesCompleted += "."
		}
		score := 0
		if !failed {
			log.Green("%6d) Passed %s %s", pNum+1, p.ID, movesCompleted)
			score = 1
		}
		rating, _ = elo.Rating(rating, p.Rating, float64(score))
	}

	fmt.Println()
	fmt.Println("========================")
	fmt.Printf("Answered %v/%v puzzles correctly\n", correct, len(puzzles))
	fmt.Println("Elapsed:", time.Since(start))
	fmt.Println("Rating:", rating)

	if len(failedIDs) != 0 {
		fmt.Println("Failed:", strings.Join(failedIDs, ","))
	}
}

func tagSearch(tag string) {
	var found = map[string]struct{}{}

	puzzledb.Query(0, func(p puzzledb.Puzzle) bool {
		for _, theme := range p.Themes {
			if tag == "*" || strings.Contains(theme, tag) {
				found[theme] = struct{}{}
			}
		}
		return false
	})

	var themes []string
	for t := range found {
		themes = append(themes, t)
	}

	sort.Strings(themes)
	for _, t := range themes {
		fmt.Println(t)
	}
}

func contains(tags []string, t string) bool {
	for _, tag := range tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}

	return false
}
