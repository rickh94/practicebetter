package server

import (
	"encoding/json"
	"log"
	"net/http"
	"practicebetter/internal/db"
	"practicebetter/internal/pages/planpages"
	"strconv"
)

func (s *Server) listPracticeSessions(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}
	pageNum, err := strconv.Atoi(page)
	log.Default().Println(pageNum)

	queries := db.New(s.DB)
	ps, err := queries.ListPracticeSessions(r.Context(), db.ListPracticeSessionsParams{
		UserID: user.ID,
		Page:   pageNum,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	hasNext, err := queries.HasMorePracticeSessions(r.Context(), db.HasMorePracticeSessionsParams{
		UserID: user.ID,
		Page:   pageNum,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	psData, err := json.Marshal(&ps)
	s.HxRender(w, r, planpages.PSList(s, string(psData), pageNum, hasNext), "Practice Sessions")
}
