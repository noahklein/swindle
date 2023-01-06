package main

import (
	"flag"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/noahklein/chess/engine"
	"github.com/noahklein/chess/puzzle"
	"github.com/noahklein/chess/uci"
	"github.com/noahklein/dragon"

	"github.com/fatih/color"
)

var (
	tag   = flag.String("tag", "", "tag filter")
	id    = flag.String("id", "", "id filter")
	limit = flag.Int("limit", 10, "limit the number of puzzles; 0 for all (not recommended)")
	depth = flag.Int("depth", 4, "depth to search puzzles at")

	tsearch = flag.String("tsearch", "", "search for tags")
)

func main() {
	flag.Parse()

	if *tsearch != "" {
		tagSearch(*tsearch)
		return
	}

	if *id != "" {
		*limit = 1
	}

	var puzzles = puzzle.PuzzleDB(*limit, func(p puzzle.Puzzle) bool {
		if *id != "" {
			return p.ID == *id
		}
		if *tag != "" && !contains(p.Themes, *tag) {
			return false
		}
		return true
	})

	color.White("Attempting %v puzzles", len(puzzles))

	correct := len(puzzles)
	start := time.Now()

	var e engine.Engine
	for pNum, p := range puzzles {
		e.NewGame()
		e.Position(p.Fen, nil)
		e.Debug(false)

		var failed bool
		var movesCompleted string
		for i := 0; i < len(p.Moves); i += 2 {
			e.Position(p.Fen, p.Moves[:i+1])
			want := p.Moves[i+1]
			result := e.Go(uci.SearchParams{
				Depth: *depth,
			})

			if result.BestMove != want {
				failed = true

				// Check mate (alternate solution.)
				m, err := dragon.ParseMove(result.BestMove)
				if err != nil {
					panic(err)
				}
				e.Move(m)

				moves, inCheck := e.GenMoves()
				if inCheck && len(moves) == 0 {
					movesCompleted += "."
					color.Yellow("%3d) Passed %s %s (alternate solution)", pNum+1, p.ID, movesCompleted)
					break
				}

				correct--
				movesCompleted += "x"
				color.Red("%3d) Failed %s %s", pNum+1, p.ID, movesCompleted)
				lichess := fmt.Sprintf("https://lichess.org/analysis/fromPosition/%v", p.Fen)
				color.Red(strings.ReplaceAll(lichess, " ", "_"))
				color.Red("%v %v", p.Rating, strings.Join(p.Themes, ", "))
				color.Red(`Wrong move: got %v, want %v, %v`, result.BestMove, want, p.Moves)

				failed = true
				break
			}
			movesCompleted += "."
		}
		if !failed {
			color.Green("%3d) Passed %s %s", pNum+1, p.ID, movesCompleted)
		}
	}

	fmt.Println()
	fmt.Println("========================")
	fmt.Printf("Answered %v/%v puzzles correctly\n", correct, len(puzzles))
	fmt.Printf("Elapsed: %v\n", time.Since(start))

}

func tagSearch(tag string) {
	var found = map[string]struct{}{}

	puzzle.PuzzleDB(0, func(p puzzle.Puzzle) bool {
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
