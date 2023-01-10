package engine

import (
	"strings"

	"github.com/noahklein/chess/log"
	"github.com/noahklein/dragon"
)

// NodeCount tracks useful stats for reporting. Not thread-safe.
type NodeCount struct {
	nodes, qNodes int
	maxPly        int16 // For reporting max depth.

	legalKiller int
}

func (nc *NodeCount) Inc()          { nc.nodes++ }
func (nc *NodeCount) Qinc()         { nc.qNodes++; nc.nodes++ }
func (nc *NodeCount) Ply(ply int16) { nc.maxPly = max(ply, nc.maxPly) }

func (nc *NodeCount) Reset() {
	nc.nodes = 0
	nc.qNodes = 0
	nc.maxPly = 0
}

// TODO: Implement.
func (e *Engine) SetOption(option string, value string) {}

// Debug enables logging and metric reporting.
func (e *Engine) Debug(isOn bool) {
	e.debug = isOn

	if isOn {
		e.Logger.Level = log.WARN
	}
}

func stringMoves(ms []dragon.Move) string {
	var b strings.Builder
	for _, m := range ms {
		b.WriteString(m.String() + " ")
	}
	return b.String()
}
