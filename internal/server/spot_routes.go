package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"practicebetter/internal/components"
	"practicebetter/internal/db"
	"practicebetter/internal/pages/librarypages"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/mavolin/go-htmx"
	"github.com/nrednav/cuid2"
)

func (s *Server) singleSpot(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	spotID := chi.URLParam(r, "spotID")
	user := r.Context().Value("user").(db.User)
	queries := db.New(s.DB)

	spot, err := queries.GetSpot(r.Context(), db.GetSpotParams{
		SpotID:  spotID,
		UserID:  user.ID,
		PieceID: pieceID,
	})
	if err != nil {
		// TODO: create a pretty 404 handler
		log.Default().Println(err)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Could not find matching spot"))
		return
	}

	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.SingleSpot(s, spot, token))
}

func (s *Server) addSpotPage(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)
	pieceID := chi.URLParam(r, "pieceID")
	user := r.Context().Value("user").(db.User)
	queries := db.New(s.DB)
	spots, err := queries.ListPieceSpots(r.Context(), db.ListPieceSpotsParams{
		PieceID: pieceID,
		UserID:  user.ID,
	})
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	s.HxRender(w, r, librarypages.AddSpotPage(s, token, pieceID, spots))
}

func (s *Server) addSpot(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	pieceID := chi.URLParam(r, "pieceID")
	queries := db.New(s.DB)
	r.ParseForm()
	idx, err := strconv.Atoi(r.FormValue("idx"))
	if err != nil {
		log.Default().Println(err)
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Invalid index",
			Title:    "Error",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}
	currentTempo := sql.NullInt64{Valid: false}
	currentTempoVal := r.FormValue("currentTempo")
	log.Default().Println(currentTempoVal)
	if currentTempoVal != "" && currentTempoVal != "null" {
		currentTempoInt, err := strconv.Atoi(currentTempoVal)
		if err != nil {
			log.Default().Println(err)
			htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
				Message:  "Invalid Current Tempo",
				Title:    "Error",
				Variant:  "error",
				Duration: 3000,
			})
			http.Error(w, "Invalid current tempo", http.StatusBadRequest)
			return
		}
		currentTempo.Int64 = int64(currentTempoInt)
		currentTempo.Valid = true
	}
	measures := sql.NullString{Valid: false}
	measuresVal := r.FormValue("measures")
	if measuresVal != "" && measuresVal != "null" {
		measures.String = measuresVal
		measures.Valid = true
	}
	spot, err := queries.CreateSpot(r.Context(), db.CreateSpotParams{
		UserID:         user.ID,
		PieceID:        pieceID,
		ID:             cuid2.Generate(),
		Name:           r.FormValue("name"),
		Idx:            int64(idx),
		Stage:          r.FormValue("stage"),
		AudioPromptUrl: r.FormValue("audioPromptUrl"),
		ImagePromptUrl: r.FormValue("imagePromptUrl"),
		NotesPrompt:    r.FormValue("notesPrompt"),
		TextPrompt:     r.FormValue("textPrompt"),
		CurrentTempo:   currentTempo,
		Measures:       measures,
	})
	if err != nil {
		log.Default().Println(err)
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not add spot: " + err.Error(),
			Title:    "Error",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	outMeasures := librarypages.SpotMeasuresOrEmpty(spot.Measures)
	htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Added Spot: " + spot.Name,
		Title:    "Spot Added!",
		Variant:  "success",
		Duration: 3000,
	})
	w.WriteHeader(http.StatusCreated)
	components.SmallSpotCard(spot.PieceID, spot.ID, spot.Name, outMeasures, spot.Stage).Render(r.Context(), w)
}

func (s *Server) editSpot(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	spotID := chi.URLParam(r, "spotID")
	user := r.Context().Value("user").(db.User)
	queries := db.New(s.DB)

	spot, err := queries.GetSpot(r.Context(), db.GetSpotParams{
		SpotID:  spotID,
		UserID:  user.ID,
		PieceID: pieceID,
	})
	if err != nil {
		// TODO: create a pretty 404 handler
		log.Default().Println(err)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Could not find matching spot"))
		return
	}
	spotData := makeSpotFormDataFromSpot(spot)
	spotJson, err := json.Marshal(spotData)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.EditSpot(s, spot, string(spotJson), token))
}

func (s *Server) updateSpot(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	spotID := chi.URLParam(r, "spotID")
	user := r.Context().Value("user").(db.User)
	queries := db.New(s.DB)
	r.ParseForm()
	idx, err := strconv.Atoi(r.FormValue("idx"))
	if err != nil {
		log.Default().Println(err)
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Invalid index",
			Title:    "Error",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}
	currentTempo := sql.NullInt64{Valid: false}
	currentTempoVal := r.FormValue("currentTempo")
	log.Default().Println(currentTempoVal)
	if currentTempoVal != "" && currentTempoVal != "null" {
		currentTempoInt, err := strconv.Atoi(currentTempoVal)
		if err != nil {
			log.Default().Println(err)
			htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
				Message:  "Invalid Current Tempo",
				Title:    "Error",
				Variant:  "error",
				Duration: 3000,
			})
			http.Error(w, "Invalid current tempo", http.StatusBadRequest)
			return
		}
		currentTempo.Int64 = int64(currentTempoInt)
		currentTempo.Valid = true
	}
	measures := sql.NullString{Valid: false}
	measuresVal := r.FormValue("measures")
	if measuresVal != "" && measuresVal != "null" {
		measures.String = measuresVal
		measures.Valid = true
	}
	err = queries.UpdateSpot(r.Context(), db.UpdateSpotParams{
		UserID:         user.ID,
		PieceID:        pieceID,
		SpotID:         spotID,
		Name:           r.FormValue("name"),
		Idx:            int64(idx),
		Stage:          r.FormValue("stage"),
		AudioPromptUrl: r.FormValue("audioPromptUrl"),
		ImagePromptUrl: r.FormValue("imagePromptUrl"),
		NotesPrompt:    r.FormValue("notesPrompt"),
		TextPrompt:     r.FormValue("textPrompt"),
		CurrentTempo:   currentTempo,
		Measures:       measures,
	})
	if err != nil {
		log.Default().Println(err)
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not update spot: " + err.Error(),
			Title:    "Error",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	spot, err := queries.GetSpot(r.Context(), db.GetSpotParams{
		SpotID:  spotID,
		UserID:  user.ID,
		PieceID: pieceID,
	})
	if err != nil {
		// TODO: create a pretty 404 handler
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	htmx.PushURL(r, "/library/pieces/"+pieceID+"/spots/"+spotID)
	htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "This spot has been updated with your new values",
		Title:    "Spot Updated!",
		Variant:  "success",
		Duration: 3000,
	})
	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.SingleSpot(s, spot, token))
}

func (s *Server) deleteSpot(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	spotID := chi.URLParam(r, "spotID")
	user := r.Context().Value("user").(db.User)
	queries := db.New(s.DB)
	err := queries.DeleteSpot(r.Context(), db.DeleteSpotParams{
		UserID:  user.ID,
		PieceID: pieceID,
		SpotID:  spotID,
	})
	if err != nil {
		log.Default().Println(err)
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not delete spot: " + err.Error(),
			Title:    "Error",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	htmx.PushURL(r, "/library/pieces/"+pieceID)
	htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "This spot has been deleted",
		Title:    "Spot Deleted",
		Variant:  "success",
		Duration: 3000,
	})
	piece, err := queries.GetPieceByID(r.Context(), db.GetPieceByIDParams{
		PieceID: pieceID,
		UserID:  user.ID,
	})
	if err != nil || len(piece) == 0 {
		// TODO: create a pretty 404 handler
		log.Default().Println(err)
		http.Error(w, "Could not find matching piece", http.StatusNotFound)
		return
	}
	token := csrf.Token(r)
	librarypages.SinglePiece(s, token, piece).Render(r.Context(), w)
}

func (s *Server) repeatPracticeSpot(w http.ResponseWriter, r *http.Request) {
	log.Default().Println("get repeat practice spot")
	pieceID := chi.URLParam(r, "pieceID")
	spotID := chi.URLParam(r, "spotID")
	user := r.Context().Value("user").(db.User)
	queries := db.New(s.DB)

	spot, err := queries.GetSpot(r.Context(), db.GetSpotParams{
		SpotID:  spotID,
		UserID:  user.ID,
		PieceID: pieceID,
	})
	if err != nil {
		// TODO: create a pretty 404 handler
		log.Default().Println(err)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Could not find matching spot"))
		return
	}
	var measures *string
	if spot.Measures.Valid {
		measures = &spot.Measures.String
	}
	var currentTempo *int64
	if spot.CurrentTempo.Valid {
		currentTempo = &spot.CurrentTempo.Int64
	}
	spotData := SpotFormData{
		ID:             &spot.ID,
		Name:           spot.Name,
		Idx:            &spot.Idx,
		Stage:          spot.Stage,
		AudioPromptUrl: spot.AudioPromptUrl,
		ImagePromptUrl: spot.ImagePromptUrl,
		NotesPrompt:    spot.NotesPrompt,
		TextPrompt:     spot.TextPrompt,
		CurrentTempo:   currentTempo,
		Measures:       measures,
	}

	spotJson, err := json.Marshal(spotData)

	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.SpotPracticeRepeatPage(s, spot, token, string(spotJson)))
}

type RepeatPracticeInfo struct {
	DurationMinutes int64
	Success         bool
}

func (s *Server) repeatPracticeSpotFinished(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	spotID := chi.URLParam(r, "spotID")
	user := r.Context().Value("user").(db.User)

	queries := db.New(s.DB)
	tx, err := s.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	qtx := queries.WithTx(tx)

	if err := qtx.RepeatPracticeSpot(r.Context(), db.RepeatPracticeSpotParams{
		SpotID:  spotID,
		UserID:  user.ID,
		PieceID: pieceID,
	}); err != nil {
		// TODO: create a pretty 404 handler
		log.Default().Println(err)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Could not find matching spot"))
		return
	}

	var info RepeatPracticeInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	practiceSessionID := cuid2.Generate()
	if err := qtx.CreatePracticeSession(r.Context(), db.CreatePracticeSessionParams{
		ID:              practiceSessionID,
		UserID:          user.ID,
		DurationMinutes: info.DurationMinutes,
		Date:            time.Now().Unix(),
	}); err != nil {
		log.Default().Println(err)
		http.Error(w, "Could not create practice session", http.StatusInternalServerError)
		return
	}

	if err := qtx.PracticeSpot(r.Context(), db.PracticeSpotParams{
		UserID:            user.ID,
		PieceID:           pieceID,
		SpotID:            spotID,
		PracticeSessionID: practiceSessionID,
	}); err != nil {
		log.Default().Println(err)
		http.Error(w, "Could not practice spot", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Default().Println(err)
		http.Error(w, "Could not commit practice session", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
