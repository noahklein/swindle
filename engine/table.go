package engine

import (
	"sync"

	"github.com/noahklein/dragon"
)

// const MB = 1 << 20
// const size = unsafe.Sizeof(Entry{})

// const tableSize uint64 = uint64(100 * MB / size)
const tableSize uint64 = 0xFFFFF

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
}

// Transpositions is a transposition table (TT); used to memoize searched positions. TTs add
// search-instability.
type Transpositions struct {
	sync.Mutex // TODO: this can be done locklessly.
	table      [tableSize]Entry
	size       uint64
}

func NewTable() *Transpositions {
	return &Transpositions{
		table: [tableSize]Entry{},
	}
}

func key(hash uint64) uint64 { return hash % tableSize }

func (t *Transpositions) Add(ply int16, e Entry) {
	e.value = min(max(mateVal, e.value), -mateVal)

	t.Lock()
	defer t.Unlock()

	t.size++
	t.table[key(e.key)] = e
}

func (t *Transpositions) Get(hash uint64) (Entry, bool) {
	t.Lock()
	defer t.Unlock()

	e := t.table[key(hash)]
	return e, e.key == hash
}

func (t *Transpositions) GetEval(hash uint64, depth int, alpha, beta int16) (int16, NodeType) {
	// TODO: Need to detect repetitions before enabling tt.
	// return 0, NodeUnknown

	e, ok := t.Get(hash)
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

func (t *Transpositions) PermillFull() int {
	return int(1000 * float64(t.size) / float64(tableSize))
}

// History is a list of board hashes seen at each ply.
type History struct {
	positions []uint64
}

// Draw checks for threefold repetitions and the fifty-move rule. The halfmove clock is
// reset whenever an irreversible move is made, i.e. pawn moves, captures, castling, and
// moves that lose castling rights.
func (hst *History) Draw(hash uint64, ply int16, halfMoveClock uint8) bool {
	if halfMoveClock >= 50 {
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
