package engine

import (
	"strconv"

	"github.com/noahklein/chess/log"
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
func (e *Engine) SetOption(option string, value string) error {
	switch option {
	case "nullmove":
		e.disableNullMove = false
		if value == "false" {
			e.disableNullMove = true
		}
	case "clear hash":
		e.ClearTT()
	case "hash":
		i, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		e.hashSizeMB = i
		e.UCI("Hash size set to %v mb", e.hashSizeMB)

	default:
		e.Warn("Unsupported option: %v", option)
	}

	return nil
}

func (e *Engine) PrintOptions() {
	e.UCI("option name Nullmove type check default true")
	e.UCI("option name Clear Hash type button")
	e.UCI("option name Hash type spin default 16 min 1 max 1024")
}

// Debug enables logging and metric reporting.
func (e *Engine) Debug(isOn bool) {
	e.debug = isOn

	if isOn {
		e.Logger.Level = log.WARN
	}
}
