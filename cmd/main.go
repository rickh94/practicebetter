package main

import (
	"encoding/gob"
	"practicebetter/internal/server"
)

func main() {
	gob.Register(server.PlanInterleaveSpotInfo{})
	gob.Register([]server.PlanInterleaveSpotInfo{})
	gob.Register(server.PracticeBreak{})

	server := server.NewServer()

	err := server.ListenAndServe()
	if err != nil {
		panic("cannot start server")
	}
}
