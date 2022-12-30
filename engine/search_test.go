package engine_test

import (
	"testing"

	"github.com/dylhunn/dragontoothmg"
	"github.com/noahklein/chess/engine"
)

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
