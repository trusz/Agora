package app

import "goazuread/src/db"

func Run() {

	dbLocation := "tmp/agora.db"
	agoraDB, _ := db.Open(dbLocation)
	SetupDB(agoraDB)

	srv := NewServer("localhost", "54324")
	srv.Start()
	srv.WaitTilRunning()
}
