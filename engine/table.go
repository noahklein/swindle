package engine

import (
	"unsafe"

	"github.com/noahklein/dragon"
)

const MB = 1 << 20
const size = unsafe.Sizeof(Entry{})

const tableSize uint64 = uint64(100 * MB / size)

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

// Table is a transposition table (TT); used to memoize searched positions. TTs add
// search-instability.
// TODO: Make goroutine-safe. This can be done locklessly, but we'll have to verify the
// result after reading; potentially dangerous.
type Table struct {
	table [tableSize]Entry
}

func NewTable() *Table {
	return &Table{
		table: [tableSize]Entry{},
	}
}

func (t *Table) Get(hash uint64) (Entry, bool) {
	e := t.table[hash%tableSize]
	return e, e.key == hash
}

func (t *Table) GetEval(hash uint64, depth int, alpha, beta int16) (int16, NodeType) {
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

func (t *Table) Add(e Entry) {
	key := e.key % tableSize
	t.table[key] = e
}
