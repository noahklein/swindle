// Handy bitboards and operations.
package engine

import (
	"github.com/noahklein/dragon"
)

func up(b uint64) uint64    { return b << 8 }
func down(b uint64) uint64  { return b >> 8 }
func right(b uint64) uint64 { return b & ^dragon.FileMasks[7] << 1 }
func left(b uint64) uint64  { return (b & ^dragon.FileMasks[0]) >> 1 }

func upRight(b uint64) uint64   { return up(right(b)) }
func upLeft(b uint64) uint64    { return up(left(b)) }
func downRight(b uint64) uint64 { return down(right(b)) }
func downLeft(b uint64) uint64  { return down(left(b)) }
