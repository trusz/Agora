package server

import (
	"os"
)

func Run() {

	argsWithoutProg := os.Args[1:]

	port := "54324"
	if len(argsWithoutProg) > 0 {
		port = argsWithoutProg[0]
	}

	dbpath := "tmp/agora_local.db"
	if len(argsWithoutProg) > 1 {
		dbpath = argsWithoutProg[1]
	}

	srv := NewServer("localhost", port, dbpath)
	srv.Start()
	srv.WaitTilRunning()
}
