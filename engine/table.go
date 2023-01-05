package engine

import (
	"unsafe"

	"github.com/noahklein/dragon"
)

const MB = 1 << 20
const size = unsafe.Sizeof(Entry{})

const tableSize = 100 * MB / size
const tSize uint64 = uint64(tableSize)

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

// Table is a transposition table; used to memoize searched positions.
type Table struct {
	table [tableSize]Entry
}

func NewTable() *Table {
	return &Table{
		table: [tableSize]Entry{},
	}
}

func (t *Table) Get(hash uint64, depth int, alpha, beta int16) (int16, NodeType) {
	// TODO: Fix table.
	return 0, NodeUnknown

	e := t.table[hash%tSize]
	if e.key != hash || e.depth < depth {
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
	key := e.key % tSize
	t.table[key] = e
}
