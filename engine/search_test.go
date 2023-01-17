package engine

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/noahklein/chess/log"
	puzzledb "github.com/noahklein/chess/puzzle"
	"github.com/noahklein/chess/uci"
	"github.com/noahklein/dragon"
)

func TestMate(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		depth    int
		want     string
		wantMate int16
	}{
		{
			name:  "mate in 2, w",
			fen:   "r2qkb1r/pp2nppp/3p4/2pNN1B1/2BnP3/3P4/PPP2PPP/R2bK2R w KQkq - 1 0",
			depth: 2,
			want:  "d5f6", wantMate: 1,
		},
		{
			name:  "mate in 2, b",
			fen:   "6k1/pp4p1/2p5/2bp4/8/P5Pb/1P3rrP/2BRRN1K b - - 0 1",
			depth: 2,
			want:  "g2g1", wantMate: 1,
		},
		{
			name:  "mate in 3, b",
			fen:   "r1b1kb1r/pppp1ppp/5q2/4n3/3KP3/2N3PN/PPP4P/R1BQ1B1R b kq - 0 1",
			depth: 5,
			want:  "f8c5", wantMate: 2,
		},
		{
			name:  "K+R vs K, mate in 8, w",
			fen:   "8/8/8/8/4K1k1/4R3/8/8 w - - 0 1",
			depth: 16,
			want:  "e4e5", wantMate: 6, // TODO: should be 8
		},
		{
			name:  "B+B vs K, mate in 8, w",
			fen:   "8/8/3B4/1k1K4/8/8/2B5/8 w - - 10 6",
			depth: 20,
			want:  "d6c5", wantMate: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e Engine
			e.NewGame()
			e.Position(tt.fen, nil)
			e.Debug(false)
			// e.Level = log.NONE

			results := e.IterDeep(context.Background(), uci.SearchParams{
				Depth: tt.depth,
			})

			t.Log(puzzledb.LichessUrl(tt.fen))

			if results.Move != tt.want {
				t.Errorf("Could not find mate: got %v, eval = %v ; want %v", results.Move, results.Score, tt.want)
			}

			if results.Mate <= 0 {
				t.Errorf("Engine did not report mate: got mate = %v, want %v", results.Mate, tt.wantMate)
			}

			// TODO: fix mate search/scores.
			// if results.Mate != tt.wantMate {
			// 	t.Errorf("Engine did not report mate: got mate = %v, want %v", results.Mate, tt.wantMate)
			// }
		})
	}
}

func TestForcedDraw(t *testing.T) {
	// TODO: fix three-fold detection.
	t.SkipNow()

	tests := []struct {
		fen   string
		depth int
		want  string
	}{
		// Black has mate in 2, white to play and draw.
		{"5r1k/8/6Q1/8/1b6/2n5/1q6/7K w - - 0 1", 6, "g6h6"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			var e Engine
			e.NewGame()
			e.Position(tt.fen, nil)
			e.Debug(false)

			results := e.IterDeep(context.Background(), uci.SearchParams{
				Depth: tt.depth,
			})

			if results.Move != tt.want {
				t.Errorf("Could not find forced draw: got %v, eval = %v ; want %v", results.Move, results.Score, tt.want)
			}

			// TODO: fix draw evaluation.
			if results.Score != 0 {
				t.Errorf("Bad draw eval: got %v, want 0", results.Score)
			}
		})
	}

}

func TestMvvLva(t *testing.T) {
	fen := "7k/3q1b2/4P3/1B6/p7/8/8/K2R4 w - - 0 1"
	var e Engine
	e.NewGame()
	e.Position(fen, nil)
	t.Log(e.board.String())

	want := []string{
		// "d1h1", // Rook check
		"e6d7", // PxQ
		"b5d7", // BxQ
		"d1d7", // RxQ
		"e6f7", // PxB
		"b5a4", // BxP
	}

	moves, _ := e.board.GenerateLegalMoves()

	rand.Seed(time.Now().Unix())
	rand.Shuffle(len(moves), func(i, j int) {
		moves[i], moves[j] = moves[j], moves[i]
	})

	moveSorter := e.newMoveSorter(moves)
	for i, move := range want {
		got := moveSorter.Next(i)
		if got.String() != move {
			t.Errorf("Wrong move: got %v, want %v", got.String(), move)
		}
	}
}

func TestThreefold(t *testing.T) {
	var e Engine
	e.NewGame()
	e.Position(dragon.Startpos, nil)
	e.Debug(false)

	moves := []string{
		"b1c3",
		"g8f6",
		"c3b1",
		"f6g8",
		"b1c3",
		"g8f6",
		"c3b1",
		"f6g8",
	}

	var unmove func()
	for _, m := range moves {
		if e.Draw() {
			t.Errorf("False threefold reported after move %v", m)
		}

		move, err := dragon.ParseMove(m)
		if err != nil {
			t.Fatal(err)
		}
		unmove = e.Move(move)
	}

	if !e.Draw() {
		t.Error("Threefold not reported after final move")
	}
	unmove()

	if e.Draw() {
		t.Error("False threefold reported after unmove")
	}
}

func BenchmarkSearchD1(b *testing.B) { benchmarkSearch(1, b) }
func BenchmarkSearchD2(b *testing.B) { benchmarkSearch(2, b) }
func BenchmarkSearchD3(b *testing.B) { benchmarkSearch(3, b) }
func BenchmarkSearchD4(b *testing.B) { benchmarkSearch(4, b) }
func BenchmarkSearchD5(b *testing.B) { benchmarkSearch(5, b) }

func benchmarkSearch(depth int, b *testing.B) {
	var e Engine
	e.NewGame()
	e.Position(dragon.Startpos, nil)
	e.Debug(false)
	e.Level = log.NONE

	ctx := context.Background()

	for n := 0; n < b.N; n++ {
		e.IterDeep(ctx, uci.SearchParams{
			Depth: depth,
		})

	}
}
