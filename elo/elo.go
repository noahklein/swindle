package elo

import "math"

const (
	Default = 1400
	K       = 32
)

func r(n int) float64 { return math.Pow10(n / 400) }

// Calculates the rating delta after a game.
func delta(ra, rb int, score float64) float64 {
	return K * (score - expected(ra, rb))
}

func expected(ra, rb int) float64 {
	return r(ra) / (r(ra) + r(rb))
}

// Score should be 0 if black won, 0.5 for a draw, and 1 if white won.
func Rating(white, black int, score float64) (int, int) {
	d := delta(white, black, score)

	w, b := float64(white)+d, float64(black)-d
	return int(w), int(b)
}
