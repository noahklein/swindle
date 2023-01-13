package engine

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/noahklein/dragon"
)

func init() {
	fmt.Println("Hash size:", tableSize/MB, "mb")
}

const (
	megabytes = 200
	MB        = 1024 * 1024
	size      = uint64(unsafe.Sizeof(Entry{}))
)

// const tableSize uint64 = 0xFFFFF
var tableSize uint64 = roundPow2(megabytes * MB / size)

type NodeType uint8

const (
	NodeUnknown NodeType = iota
	NodeExact
	NodeAlpha
	NodeBeta
)

type Entry struct {
	key   uint64
	depth int
	flag  NodeType
	value int16
	best  dragon.Move
	age   uint8
}

// Transpositions is a transposition table (TT); used to memoize searched positions.
// TTs add search-instability.
type Transpositions struct {
	sync.Mutex // TODO: this can be done locklessly.
	table      []Entry
	full       uint64
	hits       int
	age        uint8
}

func NewTranspositionTable() *Transpositions {
	return &Transpositions{
		table: make([]Entry, tableSize),
	}
}

func key(hash uint64) uint64 { return hash % tableSize }

// Always replaces previous entries.
func (tt *Transpositions) Add(ply int16, e Entry) {
	// TODO: Adjust mate scores for depth. TT can return misleading results.

	tt.Lock()
	defer tt.Unlock()

	k := key(e.key)
	existing := tt.table[k]

	isEmpty := existing.flag == NodeUnknown

	// Don't replace good entries.
	if !isEmpty && (e.depth < existing.depth || existing.age < e.age) {
		return
	}

	if isEmpty {
		tt.full++
	}

	tt.table[k] = e
}

func (tt *Transpositions) Get(hash uint64) (Entry, bool) {
	tt.Lock()
	defer tt.Unlock()

	e := tt.table[key(hash)]
	ok := e.key == hash

	if ok {
		tt.hits++
	}

	return e, ok
}

func (tt *Transpositions) GetEval(hash uint64, depth int, alpha, beta int16) (int16, NodeType) {
	e, ok := tt.Get(hash)
	if !ok || e.depth < depth {
		return 0, NodeUnknown
	}

	switch {
	case e.flag == NodeExact:
		return e.value, NodeExact
	case e.flag == NodeAlpha && e.value <= alpha:
		return alpha, NodeAlpha
	case e.flag == NodeBeta && e.value >= beta:
		return beta, NodeBeta
	}

	return 0, NodeUnknown
}

func (tt *Transpositions) PermillFull() int {
	tt.Lock()
	defer tt.Unlock()
	return int(1000 * float64(tt.full) / float64(tableSize))
}

func (tt *Transpositions) Hits() int {
	tt.Lock()
	defer tt.Unlock()
	return tt.hits
}

// History is a list of board hashes seen at each ply.
type History struct {
	positions []uint64
}

// Draw checks for threefold repetitions and the fifty-move rule. The halfmove clock is
// reset whenever an irreversible move is made, i.e. pawn moves, captures, castling, and
// moves that lose castling rights.
func (hst *History) Draw(hash uint64, ply int16, halfMoveClock uint8) bool {
	if halfMoveClock >= 100 {
		return true
	}
	if halfMoveClock < 8 {
		// Not enough moves for there to be a repetition.
		return false
	}

	var count uint8
	start := min(0, ply-int16(halfMoveClock+1))
	pos := hst.positions[start:]

	// Repetitions can only occur every 4th position.
	for i := 0; i < len(pos); i += 4 {
		if pos[i] != hash {
			continue
		}
		count++
		if count == 3 {
			return true
		}
	}
	return false
}

func (hst *History) Push(hash uint64) {
	hst.positions = append(hst.positions, hash)
}
func (hst *History) Pop() {
	hst.positions = hst.positions[:len(hst.positions)-1]
}

func (hst *History) Copy() *History {
	var c History
	c.positions = hst.positions
	return &c
}

// Round to the nearest power-of-two.
func roundPow2(n uint64) uint64 {
	pow := uint64(1)
	for pow < n {
		pow *= 2
	}
	return pow
}
