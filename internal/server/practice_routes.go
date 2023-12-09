package server

import (
	"net/http"
	"practicebetter/internal/pages/practicepages"
)

func (s *Server) randomPractice(w http.ResponseWriter, r *http.Request) {
	s.HxRender(w, r, practicepages.RandomPractice(s), "Random Practice")
}

func (s *Server) sequencePractice(w http.ResponseWriter, r *http.Request) {
	s.HxRender(w, r, practicepages.SequencePractice(s), "Sequence Practice")
}

func (s *Server) repeatPractice(w http.ResponseWriter, r *http.Request) {
	s.HxRender(w, r, practicepages.RepeatPractice(s), "Repeat Practice")
}

func (s *Server) startingPointPractice(w http.ResponseWriter, r *http.Request) {
	s.HxRender(w, r, practicepages.StartingPointPractice(s), "Starting Point")
}
