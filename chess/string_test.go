package chess

import (
	"testing"
)

func TestParseFen(t *testing.T) {
	tests := []struct {
		name string
		fen  string
	}{
		{
			name: "startpos",
			fen:  StartPos,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFen(tt.fen)
			if err != nil {
				t.Error("ParseFen() failed", err)
			}
			t.Logf("%064b", got.Knights)
			t.Log(got.String())
		})
	}
}
