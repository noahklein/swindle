package engine_test

import (
	"context"
	"math"
	"testing"

	"github.com/dylhunn/dragontoothmg"
	"github.com/noahklein/chess/engine"
	"github.com/noahklein/chess/uci"
)

func TestMate(t *testing.T) {
	tests := []struct {
		fen   string
		depth int
		want  string
	}{
		// Mate in 2
		{"r2qkb1r/pp2nppp/3p4/2pNN1B1/2BnP3/3P4/PPP2PPP/R2bK2R w KQkq - 1 0", 2, "d5f6"},
		{"6k1/pp4p1/2p5/2bp4/8/P5Pb/1P3rrP/2BRRN1K b - - 0 1", 2, "g2g1"},
		// Mate in 3
		{"r1b1kb1r/pppp1ppp/5q2/4n3/3KP3/2N3PN/PPP4P/R1BQ1B1R b kq - 0 1", 4, "f8c5"},
	}

	for _, tt := range tests {
		t.Run(tt.fen, func(t *testing.T) {
			var e engine.Engine
			e.Position(tt.fen, nil)

			results := e.Search(context.Background(), uci.SearchParams{
				Depth: tt.depth,
			})

			if results.BestMove != tt.want {
				t.Errorf("Could not find mate: got %v, eval = %v ; want %v", results.BestMove, results.Score, tt.want)
			}

			mateValThreshold := (math.Abs(math.MinInt16/2) - 200) / 100
			if math.Abs(float64(results.Score)) < mateValThreshold {
				t.Errorf("Bad mate eval: got %v, want > %v", results.Score, mateValThreshold)
			}
		})
	}
}

func TestForcedDraw(t *testing.T) {
	tests := []struct {
		fen   string
		depth int
		want  string
	}{
		// Black has mate in 2, white to play and draw.
		{"5r1k/8/6Q1/8/1b6/2n5/1q6/7K w - - 0 1", 3, "g6h6"},
	}

	for _, tt := range tests {
		t.Run(tt.fen, func(t *testing.T) {
			var e engine.Engine
			e.Position(tt.fen, nil)

			results := e.Search(context.Background(), uci.SearchParams{
				Depth: tt.depth,
			})

			if results.BestMove != tt.want {
				t.Errorf("Could not find forced draw: got %v, eval = %v ; want %v", results.BestMove, results.Score, tt.want)
			}

			// TODO: fix draw evaluation.
			// if results.Score != 0 {
			// 	t.Errorf("Bad draw eval: got %v, want 0", results.Score)
			// }
		})
	}

}

func TestOccupied(t *testing.T) {
	board := dragontoothmg.ParseFen(dragontoothmg.Startpos)
	tests := []struct {
		square uint8
		want   bool
	}{
		{0, true},
		{1, true},
		{14, true},
		{15, true},
		{16, false},
		{17, false},
		{32, false},
		{33, false},
		{58, true},
		{63, true},
	}
	for _, tt := range tests {
		if got := engine.Occupied(&board, tt.square); got != tt.want {
			t.Errorf("Occupied(%v) = %v, want %v", tt.square, got, tt.want)
			t.Logf("%64b", uint64(1)<<tt.square)
			t.Logf("%b", board.White.All|board.Black.All)
		}
	}
}
