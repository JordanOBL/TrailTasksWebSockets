package main

import (
	"github.com/jordanOBL/TrailTasksWebSockets/internal/server"
)

func main() {
	//create new server
	s := server.NewServer("", 8080)
	//start server
	err := s.Start()
	if err != nil {
		panic(err)
	}
	select {}

}
