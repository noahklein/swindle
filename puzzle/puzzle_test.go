package puzzle

import (
	"fmt"
	"strings"
	"testing"

	"github.com/noahklein/chess/engine"
	"github.com/noahklein/chess/uci"
)

func TestPuzzles(t *testing.T) {
	var puzzles = PuzzleDB(10, func(p Puzzle) bool {
		return len(p.moves) <= 2 && p.rating < 3000 && !contains(p.themes, "matein1")
	})

	for _, p := range puzzles {
		t.Run(p.id, func(t *testing.T) {
			var e engine.Engine
			e.NewGame()
			e.Position(p.fen, nil)
			e.Debug(false)

			for i := 0; i < len(p.moves); i += 2 {
				e.Position(p.fen, p.moves[:i+1])
				want := p.moves[i+1]
				result := e.Go(uci.SearchParams{
					Depth: 4,
				})

				if result.BestMove != want {
					lichess := fmt.Sprintf("https://lichess.org/analysis/fromPosition/%v", p.fen)
					t.Log(strings.ReplaceAll(lichess, " ", "_"))
					t.Log(p.rating, p.themes)
					t.Fatalf(`Wrong move: got %v, want %v`, result.BestMove, p.moves)
				}
			}
		})
	}
}

func contains(tags []string, t string) bool {
	for _, tag := range tags {
		if t == strings.ToLower(tag) {
			return true
		}
	}
	return false
}
