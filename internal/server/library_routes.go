package server

import (
	"log"
	"net/http"
	"practicebetter/internal/pages/librarypages"

	"github.com/gorilla/csrf"
)

func (s *Server) libraryDashboard(w http.ResponseWriter, r *http.Request) {
	s.HxRender(w, r, librarypages.Dashboard([]string{"piece 1", "piece 2", "piece 3"}))
}

func (s *Server) createPieceForm(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.CreatePiecePage(s, token))
}

func (s *Server) createPiece(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	log.Default().Println(r.Form.Get("title"))
	log.Default().Println(r.Form.Get("spots"))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	// s.HxRender(w, r, librarypages.CreatePiecePage(s))
}
