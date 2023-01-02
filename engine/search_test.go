package engine

import (
	"context"
	"reflect"
	"testing"

	"github.com/dylhunn/dragontoothmg"
	"github.com/noahklein/chess/uci"
)

func TestMate(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		depth    int
		want     string
		wantMate int16
	}{
		// Mate in 2
		{
			name:  "mate in 2, w",
			fen:   "r2qkb1r/pp2nppp/3p4/2pNN1B1/2BnP3/3P4/PPP2PPP/R2bK2R w KQkq - 1 0",
			depth: 2,
			want:  "d5f6", wantMate: 2,
		},
		{
			name:  "mate in 2, b",
			fen:   "6k1/pp4p1/2p5/2bp4/8/P5Pb/1P3rrP/2BRRN1K b - - 0 1",
			depth: 2,
			want:  "g2g1", wantMate: 2,
		},
		// Mate in 3
		{
			name:  "mate in 3, b",
			fen:   "r1b1kb1r/pppp1ppp/5q2/4n3/3KP3/2N3PN/PPP4P/R1BQ1B1R b kq - 0 1",
			depth: 5,
			want:  "f8c5", wantMate: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e Engine
			e.NewGame()
			e.Position(tt.fen, nil)
			e.Debug(false)

			results := e.Search(context.Background(), uci.SearchParams{
				Depth: tt.depth,
			})

			if results.BestMove != tt.want {
				t.Errorf("Could not find mate: got %v, eval = %v ; want %v", results.BestMove, results.Score, tt.want)
			}

			mateValThreshold := (abs(mateVal) - 200) / 100
			if abs(results.Score) < mateValThreshold {
				t.Errorf("Bad mate eval: got %v, want > %v", results.Score, mateValThreshold)
			}

			if results.Mate != tt.wantMate {
				t.Errorf("Engine did not report mate: got mate = %v, want %v", results.Mate, tt.wantMate)
			}
		})
	}
}

// TODO: fix three-fold detection.
func TestForcedDraw(t *testing.T) {
	tests := []struct {
		fen   string
		depth int
		want  string
	}{
		// Black has mate in 2, white to play and draw.
		// {"5r1k/8/6Q1/8/1b6/2n5/1q6/7K w - - 0 1", 5, "g6h6"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			var e Engine
			e.NewGame()
			e.Position(tt.fen, nil)
			e.Debug(false)

			results := e.Search(context.Background(), uci.SearchParams{
				Depth: tt.depth,
			})

			if results.BestMove != tt.want {
				t.Errorf("Could not find forced draw: got %v, eval = %v ; want %v", results.BestMove, results.Score, tt.want)
			}

			// TODO: fix draw evaluation.
			if results.Score != 0 {
				t.Errorf("Bad draw eval: got %v, want 0", results.Score)
			}
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
		if got := Occupied(&board, tt.square); got != tt.want {
			t.Errorf("Occupied(%v) = %v, want %v", tt.square, got, tt.want)
			t.Logf("%64b", uint64(1)<<tt.square)
			t.Logf("%b", board.White.All|board.Black.All)
		}
	}
}

func TestMvvLva(t *testing.T) {
	fen := "7k/3q4/4P3/8/B7/8/8/K2R4 w - - 0 1"
	var e Engine
	e.NewGame()
	e.Position(fen, nil)

	moves := e.sortMoves(e.board.GenerateLegalMoves())[0:4]

	var got []string
	for _, m := range moves {
		got = append(got, m.String())
	}

	want := []string{
		"d1h1", // Rook check
		"e6d7", // PxQ
		"a4d7", // BxQ
		"d1d7", // RxQ
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Wrong sort order: got %v, want %v", got, want)
	}
}
