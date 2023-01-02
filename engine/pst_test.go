// Piece square tables.
package engine

import (
	"fmt"
	"testing"

	"github.com/noahklein/dragon"
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
		square  uint8
		piece   int
		isWhite bool

		wantMG, wantEG int16
	}{
		{
			square:  8, // a2
			isWhite: true, piece: dragon.Pawn,
			wantMG: -35, wantEG: 13,
		},
		{
			square:  8, // a2
			isWhite: false, piece: dragon.Pawn,
			wantMG: 98, wantEG: 178,
		},
		{
			square:  18, // c3
			isWhite: true, piece: dragon.Knight,
			wantMG: 12, wantEG: -1,
		},
		{
			square:  18, // c3
			isWhite: false, piece: dragon.Knight,
			wantMG: 37, wantEG: 10,
		},
		{
			square:  9, // b2
			isWhite: true, piece: dragon.Bishop,
			wantMG: 15, wantEG: -18,
		},
		{
			square:  9, // b2
			isWhite: false, piece: dragon.Bishop,
			wantMG: 16, wantEG: -4,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v %v at %v", tt.isWhite, tt.piece, tt.square), func(t *testing.T) {
			pc := pieceColor(tt.piece, tt.isWhite)
			if got := MidGameTable[pc][tt.square]; got != tt.wantMG {
				t.Errorf("Bad midgame value: got %v, want %v", got, tt.wantMG)
			}
			if got := EndGameTable[pc][tt.square]; got != tt.wantEG {
				t.Errorf("Bad endgame value: got %v, want %v", got, tt.wantEG)
			}
		})
	}

}
