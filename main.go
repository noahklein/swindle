package main

import (

	// _ "net/http/pprof"

	"github.com/noahklein/chess/engine"
	"github.com/noahklein/chess/uci"
)

func main() {
	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	var e engine.Engine
	uci.Run(&e)
}
