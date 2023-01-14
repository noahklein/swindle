package puzzledb

import (
	"encoding/csv"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

type Puzzle struct {
	ID     string
	Fen    string
	Moves  []string
	Rating int
	Themes []string
}

// Query gets a list of puzzles from the puzzle csv file. Run download.sh before using.
func Query(limit int, predicate func(Puzzle) bool) []Puzzle {
	// PuzzleId,FEN,Moves,Rating,RatingDeviation,Popularity,NbPlays,Themes,GameUrl,OpeningFamily,OpeningVariation
	f, err := os.Open("./lichess_db_puzzle.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)

	if limit == 0 {
		limit = math.MaxInt
	}

	var puzzles []Puzzle
	for len(puzzles) < limit {
		r, err := csvReader.Read()
		if err == io.EOF {
			return puzzles
		}

		rating, err := strconv.Atoi(r[3])
		if err != nil {
			panic(err)
		}
		p := Puzzle{
			ID:     r[0],
			Fen:    r[1],
			Moves:  strings.Split(r[2], " "),
			Rating: rating,
			Themes: strings.Split(r[7], " "),
		}

		if !predicate(p) {
			continue
		}
		puzzles = append(puzzles, p)
	}

	return puzzles
}

const lichess = "https://lichess.org/analysis/fromPosition/"

func LichessUrl(fen string) string {
	return lichess + strings.ReplaceAll(fen, " ", "_")
}
