package engine

import (
	"testing"

	"github.com/noahklein/dragon"
)

func TestEval(t *testing.T) {
	tests := []struct {
		board dragon.Board
		want  int16
	}{
		{
			board: dragon.ParseFen(dragon.Startpos),
			want:  0,
		},
	}
	for _, tt := range tests {
		if got := Eval(&tt.board); got != tt.want {
			t.Errorf("Eval() = %v, want %v", got, tt.want)
		}
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
