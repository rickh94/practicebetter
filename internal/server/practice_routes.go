package server

import (
	"net/http"
	"pbgo/internal/pages/practicepages"
)

func (s *Server) randomPractice(w http.ResponseWriter, r *http.Request) {
	s.HxRender(w, r, practicepages.RandomPractice(s))
}

func (s *Server) sequencePractice(w http.ResponseWriter, r *http.Request) {
	s.HxRender(w, r, practicepages.SequencePractice(s))
}

func (s *Server) repeatPractice(w http.ResponseWriter, r *http.Request) {
	s.HxRender(w, r, practicepages.RepeatPractice(s))
}

func (s *Server) startingPointPractice(w http.ResponseWriter, r *http.Request) {
	s.HxRender(w, r, practicepages.StartingPointPractice(s))
}
