package puzzle

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"strings"
)

type Puzzle struct {
	id     string
	fen    string
	moves  []string
	rating int
	themes []string
}

func PuzzleDB(howMany int, predicate func(Puzzle) bool) []Puzzle {
	// PuzzleId,FEN,Moves,Rating,RatingDeviation,Popularity,NbPlays,Themes,GameUrl,OpeningFamily,OpeningVariation
	f, err := os.Open("./lichess_db_puzzle.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)

	var puzzles []Puzzle
	for len(puzzles) < howMany {
		r, err := csvReader.Read()
		if err == io.EOF {
			break
		}

		rating, err := strconv.Atoi(r[3])
		if err != nil {
			panic(err)
		}
		p := Puzzle{
			id:     r[0],
			fen:    r[1],
			moves:  strings.Split(r[2], " "),
			rating: rating,
			themes: strings.Split(r[7], " "),
		}

		if !predicate(p) {
			continue
		}
		puzzles = append(puzzles, p)
	}

	return puzzles
}
