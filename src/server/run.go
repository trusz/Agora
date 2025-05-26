package server

func Run() {

	srv := NewServer("localhost", "54324")
	srv.Start()
	srv.WaitTilRunning()
}
