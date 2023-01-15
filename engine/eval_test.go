package engine

import (
	"testing"

	"github.com/noahklein/dragon"
)

func TestEval(t *testing.T) {
	tests := []struct {
		name  string
		board dragon.Board
		want  int16
	}{
		{
			name:  "startpos is equal",
			board: dragon.ParseFen(dragon.Startpos),
			want:  0,
		},
		{
			name:  "down a knight and a pawn",
			board: dragon.ParseFen("r1bqkbnr/ppp1pppp/2n5/8/2BP4/5p2/PPP2PPP/RNBQK2R w KQkq - 0 1"),
			want:  -440,
		},
		{
			name:  "down a knight, pawn, and rook; black has 2 queens",
			board: dragon.ParseFen("r1bqkbnr/ppp1pppp/2n5/8/2BP4/8/PPP2P1P/RNBQK2q w Qkq - 0 1"),
			want:  -1840,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Eval(&tt.board); got < tt.want-100 || got > tt.want+100 {
				t.Errorf("Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMateScore(t *testing.T) {
	tests := []struct {
		score, ply, want int16
	}{
		{-(mateVal + 40), 30, 5},
		{-(mateVal + 41), 30, 5},
		{(mateVal + 20), 15, -2},
		{(mateVal + 20), 16, -2},
	}

	for _, tt := range tests {
		if got := mateScore(tt.score, tt.ply); got != tt.want {
			t.Errorf("Wrong score: mateScore(%v, %v) = %v, want %v", tt.score, tt.ply, got, tt.want)
		}
	}
}

var result int16

func BenchmarkEval(b *testing.B) {
	board := dragon.ParseFen(dragon.Startpos)
	for i := 0; i < b.N; i++ {
		result = Eval(&board)
	}
}
