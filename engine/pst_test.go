// Piece square tables.
package engine

import "testing"

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
