package engine

import (
	"fmt"

	"github.com/fatih/color"
)

// Not thread-safe.
type NodeCount struct{ nodes, qNodes int }

func (nc *NodeCount) Inc()   { nc.nodes++ }
func (nc *NodeCount) Qinc()  { nc.qNodes++; nc.nodes++ }
func (nc *NodeCount) Reset() { nc.nodes = 0; nc.qNodes = 0 }

// TODO: Implement.
func (e *Engine) SetOption(option string, value string) {}

// Debug enables logging and metric reporting.
func (e *Engine) Debug(isOn bool) {
	e.debug = isOn
}

func (e *Engine) Print(s string, a ...any) {
	if e.debug {
		fmt.Printf("info string "+s+"\n", a...)
	}
}

func (e *Engine) Error(s string, a ...any) {
	color.Red(s, a...)
}
