package main

import (
	"log"
	"net/http"

	"github.com/noahklein/chess/engine"
	"github.com/noahklein/chess/uci"

	_ "net/http/pprof"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	e := &engine.Engine{}

	uci.Run(e)

}
