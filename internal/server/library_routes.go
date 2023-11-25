package server

import (
	"net/http"
	"pbgo/internal/pages/librarypages"
)

func (s *Server) libraryDashboard(w http.ResponseWriter, r *http.Request) {
	s.HxRender(w, r, librarypages.Dashboard())
}
