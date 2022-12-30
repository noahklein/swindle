package engine

import (
	"testing"

	"github.com/dylhunn/dragontoothmg"
)

func TestEval(t *testing.T) {
	tests := []struct {
		board dragontoothmg.Board
		want  int16
	}{
		{
			board: dragontoothmg.ParseFen(dragontoothmg.Startpos),
			want:  0,
		},
	}
	for _, tt := range tests {
		if got := Eval(&tt.board); got != tt.want {
			t.Errorf("Eval() = %v, want %v", got, tt.want)
		}
	}
}

var result int16

func BenchmarkEval(b *testing.B) {
	board := dragontoothmg.ParseFen(dragontoothmg.Startpos)
	for i := 0; i < b.N; i++ {
		result = Eval(&board)
	}
}
