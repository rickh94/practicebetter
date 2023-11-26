package server

import (
	"net/http"
	"practicebetter/internal/pages/librarypages"
)

func (s *Server) libraryDashboard(w http.ResponseWriter, r *http.Request) {
	s.HxRender(w, r, librarypages.Dashboard())
}
