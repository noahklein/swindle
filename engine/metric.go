package engine

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/noahklein/dragon"
)

// NodeCount tracks useful stats for reporting. Not thread-safe.
type NodeCount struct {
	nodes, qNodes int
	maxPly        int16 // For reporting max depth.
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
}

func (e *Engine) Logf(s string, a ...any) {
	if e.debug {
		fmt.Printf(s+"\n", a...)
	}
}

func (e *Engine) Warn(s string, a ...any) {
	if e.debug {
		color.Yellow(s, a...)

	}
}

func (e *Engine) Error(s string, a ...any) {
	color.Red(s, a...)
}

func stringMoves(ms []dragon.Move) string {
	var b strings.Builder
	for _, m := range ms {
		b.WriteString(m.String() + " ")
	}
	return b.String()
}
