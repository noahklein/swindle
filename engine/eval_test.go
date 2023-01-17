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
			name:  "doubled-passed pawns on 2nd and 3rd rank",
			board: dragon.ParseFen("8/8/8/8/8/P7/P7/8 w - - 0 1"),
			want:  217,
		},
		{
			name:  "passed pawns on 2nd rank",
			board: dragon.ParseFen("8/8/8/8/8/8/PP7/8 w - - 0 1"),
			want:  221,
		},
		{
			name:  "2 passed pawns on 4th rank",
			board: dragon.ParseFen("8/8/8/8/PP6/8/8/8 w - - 0 1"),
			want:  302,
		},
		{
			name:  "2 black passed pawns on 5th rank",
			board: dragon.ParseFen("8/8/8/pp6/8/8/8/8 b - - 0 1"),
			want:  251,
		},
		{
			name:  "2 black passed pawns on 4th rank",
			board: dragon.ParseFen("8/8/8/8/pp6/8/8/8 w - - 0 1"),
			want:  329,
		},
		// Messy positions:
		{
			name:  "down a knight and a pawn",
			board: dragon.ParseFen("r1bqkbnr/ppp1pppp/2n5/8/2BP4/5p2/PPP2PPP/RNBQK2R w KQkq - 0 1"),
			want:  -429,
		},
		{
			name:  "down a knight, pawn, and rook; black has 2 queens",
			board: dragon.ParseFen("r1bqkbnr/ppp1pppp/2n5/8/2BP4/8/PPP2P1P/RNBQK2q w Qkq - 0 1"),
			want:  -1849,
		},
	}

	const threshold = 20 // Got should be within threshold of want.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Eval(&tt.board); got < tt.want-threshold || got > tt.want+threshold {
				t.Log(tt.board.String())
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

func BenchmarkEvalKiwipete(b *testing.B) {
	board := dragon.ParseFen("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq -")
	for i := 0; i < b.N; i++ {
		result = Eval(&board)
	}
}
