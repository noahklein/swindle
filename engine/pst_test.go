// Piece square tables.
package engine

import (
	"fmt"
	"testing"

	"github.com/dylhunn/dragontoothmg"
)

func TestFlip(t *testing.T) {
	tests := []struct {
		square uint8
		want   uint8
	}{
		{16, 40},
		{2, 58},
	}
	for _, tt := range tests {
		if got := Flip(tt.square); got != tt.want {
			t.Errorf("Flip() = %v, want %v", got, tt.want)
		}
	}
}

func TestValues(t *testing.T) {
	tests := []struct {
		square uint8
		piece  int
		color  Color

		wantMG, wantEG int16
	}{
		{
			square: 8, // a2
			piece:  dragontoothmg.Pawn,
			color:  White,
			wantMG: -35,
			wantEG: 13,
		},
		{
			square: 8, // a2
			piece:  dragontoothmg.Pawn,
			color:  Black,
			wantMG: 98,
			wantEG: 178,
		},
		{
			square: 18, // c3
			piece:  dragontoothmg.Knight,
			color:  White,
			wantMG: 12,
			wantEG: -1,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v %v at %v", tt.color, tt.piece, tt.square), func(t *testing.T) {
			pc := pieceColor(tt.piece, tt.color)
			if got := MidGameTable[pc][tt.square]; got != tt.wantMG {
				t.Errorf("Bad midgame value: got %v, want %v", got, tt.wantMG)
			}
			if got := EndGameTable[pc][tt.square]; got != tt.wantEG {
				t.Errorf("Bad endgame value: got %v, want %v", got, tt.wantEG)
			}
		})
	}

}
