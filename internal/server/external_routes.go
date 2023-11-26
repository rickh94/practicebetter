package server

import (
	"net/http"
	"practicebetter/internal/pages"
)

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	s.HxRender(w, r, pages.IndexPage())
}

func (s *Server) about(w http.ResponseWriter, r *http.Request) {
	s.HxRender(w, r, pages.AboutPage(s))
}
