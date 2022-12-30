package main

import (
	"github.com/noahklein/chess/engine"
	"github.com/noahklein/chess/uci"
)

func main() {
	e := &engine.Engine{}

	uci.Run(e)

}
