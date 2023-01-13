package engine

import (
	"sync"

	"github.com/noahklein/dragon"
)

// Killer moves, moves that caused a beta-cuttoff. We store 2 per ply.
type Killer struct {
	sync.Mutex
	moves map[int16][2]dragon.Move
}

func NewKiller() *Killer {
	return &Killer{
		moves: map[int16][2]dragon.Move{},
	}
}

func (k *Killer) Add(ply int16, move dragon.Move) {
	k.Lock()
	defer k.Unlock()

	kms := k.moves[ply]
	if kms[0] == move {
		return
	}
	// Push new move to front.
	kms[0], kms[1] = move, kms[0]
}

// Returns 0000 if empty which translates to a1a1, an impossible move.
func (k *Killer) Get(ply int16) [2]dragon.Move {
	k.Lock()
	defer k.Unlock()

	return k.moves[ply]
}
