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
}

func NewTable() *Transpositions {
	return &Transpositions{
		table: [tableSize]Entry{},
	}
}

func key(hash uint64) uint64 { return hash % tableSize }

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

func (t *Transpositions) Add(ply int16, e Entry) {
	e.value = min(max(mateVal, e.value), -mateVal)
	t.Lock()
	defer t.Unlock()

	t.table[key(e.key)] = e
}

// History is a list of board hashes seen at each ply.
type History struct {
	positions []uint64
}

// Threefold checks for threefold repetitions.
func (hst *History) Threefold(hash uint64, ply int16, halfMoveClock uint8) bool {
	if halfMoveClock < 8 {
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

func (hst *History) Add(hash uint64) {
	hst.positions = append(hst.positions, hash)
}
func (hst *History) Remove() {
	hst.positions = hst.positions[:len(hst.positions)-1]
}

func (hst *History) Copy() *History {
	var c History
	c.positions = hst.positions
	return &c
}
