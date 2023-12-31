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
	s.HxRender(w, r, librarypages.SingleSpot(s, spot, token), spot.Name+" - "+spot.PieceTitle)
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
	var pieceTitle string
	if len(spots) > 0 {
		pieceTitle = spots[0].PieceTitle
	} else {
		piece, err := queries.GetPieceWithoutSpots(r.Context(), db.GetPieceWithoutSpotsParams{
			ID:     pieceID,
			UserID: user.ID,
		})
		if err != nil {
			log.Default().Println(err)
			http.Error(w, "Could not find matching piece", http.StatusNotFound)
			return
		}
		pieceTitle = piece.Title
	}

	s.HxRender(w, r, librarypages.AddSpotPage(s, token, pieceID, pieceTitle, spots), pieceTitle)
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
	s.HxRender(w, r, librarypages.EditSpot(s, spot, string(spotJson), token), spot.PieceTitle)
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
	var stageStarted int64
	spotStageInfo, err := queries.GetSpotStageStarted(r.Context(), db.GetSpotStageStartedParams{
		SpotID:  spotID,
		UserID:  user.ID,
		PieceID: pieceID,
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
	if spotStageInfo.Stage == r.FormValue("stage") && spotStageInfo.StageStarted.Valid {
		stageStarted = spotStageInfo.StageStarted.Int64
	} else {
		stageStarted = time.Now().Unix()
	}
	err = queries.UpdateSpot(r.Context(), db.UpdateSpotParams{
		Name:           r.FormValue("name"),
		Idx:            int64(idx),
		Stage:          r.FormValue("stage"),
		StageStarted:   sql.NullInt64{Int64: stageStarted, Valid: true},
		AudioPromptUrl: r.FormValue("audioPromptUrl"),
		ImagePromptUrl: r.FormValue("imagePromptUrl"),
		NotesPrompt:    r.FormValue("notesPrompt"),
		TextPrompt:     r.FormValue("textPrompt"),
		CurrentTempo:   currentTempo,
		Measures:       measures,
		SpotID:         spotID,
		UserID:         user.ID,
		PieceID:        pieceID,
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
	s.HxRender(w, r, librarypages.SingleSpot(s, spot, token), spot.Name+" - "+spot.PieceTitle)
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
	// TODO: refactor to get and render piece function
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

	breakdown := getSpotBreakdown(piece)
	token := csrf.Token(r)
	librarypages.SinglePiece(s, token, piece, breakdown).Render(r.Context(), w)
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
	s.HxRender(w, r, librarypages.SpotPracticeRepeatPage(s, spot, token, string(spotJson)), spot.Name+" - "+spot.PieceTitle)
}

type RepeatPracticeInfo struct {
	DurationMinutes int64
	Success         bool
	ToStage         string
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

	var info RepeatPracticeInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: better error handling
	var practiceSessionID string
	activePracticePlanID, ok := s.GetActivePracticePlanID(r.Context())
	if ok {
		activePracticePlan, err := qtx.GetPracticePlan(r.Context(), db.GetPracticePlanParams{
			ID:     activePracticePlanID,
			UserID: user.ID,
		})
		if err != nil {
			log.Default().Println(err)
			http.Error(w, "Could not get active practice plan", http.StatusInternalServerError)
			return
		}
		if !activePracticePlan.PracticeSessionID.Valid {
			practiceSessionID = cuid2.Generate()
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
		}
		practiceSessionID = activePracticePlan.PracticeSessionID.String
		if err := qtx.ExtendPracticeSessionToNow(r.Context(), db.ExtendPracticeSessionToNowParams{
			ID:     practiceSessionID,
			UserID: user.ID,
		}); err != nil {
			log.Default().Println(err)
			http.Error(w, "Could not extend practice session", http.StatusInternalServerError)
			return
		}
	} else {
		log.Default().Println(err)
		practiceSessionID = cuid2.Generate()
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
	}

	if err := qtx.CreatePracticeSpot(r.Context(), db.CreatePracticeSpotParams{
		UserID:            user.ID,
		SpotID:            spotID,
		PracticeSessionID: practiceSessionID,
	}); err != nil {
		if err := qtx.AddRepToPracticeSpot(r.Context(), db.AddRepToPracticeSpotParams{
			UserID:            user.ID,
			SpotID:            spotID,
			PracticeSessionID: practiceSessionID,
		}); err != nil {
			log.Default().Println(err)
			http.Error(w, "Could not practice spot", http.StatusInternalServerError)
			return
		}
	}

	// TODO: consider more about whether this should require success
	if activePracticePlanID != "" {
		if err := qtx.CompletePracticePlanSpot(r.Context(), db.CompletePracticePlanSpotParams{
			UserID: user.ID,
			SpotID: spotID,
			PlanID: activePracticePlanID,
		}); err != nil {
			log.Default().Println(err)
			http.Error(w, "Could not complete practice plan spot", http.StatusInternalServerError)
			return
		}
	}

	if info.Success {
		switch info.ToStage {
		case "random":
			if err := qtx.PromoteSpotToRandom(r.Context(), db.PromoteSpotToRandomParams{
				UserID:  user.ID,
				PieceID: pieceID,
				SpotID:  spotID,
			}); err != nil {
				log.Default().Println(err)
				http.Error(w, "Could not promote to random", http.StatusInternalServerError)
				return
			}
		case "extra_repeat":
			if err := qtx.PromoteSpotToExtraRepeat(r.Context(), db.PromoteSpotToExtraRepeatParams{
				UserID:  user.ID,
				PieceID: pieceID,
				SpotID:  spotID,
			}); err != nil {
				log.Default().Println(err)
				http.Error(w, "Could not promote to more repeat", http.StatusInternalServerError)
				return
			}
		}
	}

	if err := tx.Commit(); err != nil {
		log.Default().Println(err)
		http.Error(w, "Could not commit practice session", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) getEditRemindersForm(w http.ResponseWriter, r *http.Request) {
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
	librarypages.EditRemindersSummary(spot.TextPrompt, spot.PieceID, spot.ID, token, "").Render(r.Context(), w)
}

func (s *Server) updateReminders(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	spotID := chi.URLParam(r, "spotID")
	user := r.Context().Value("user").(db.User)
	queries := db.New(s.DB)

	newText := r.FormValue("text")
	spot, err := queries.UpdateTextPrompt(r.Context(), db.UpdateTextPromptParams{
		SpotID:     spotID,
		UserID:     user.ID,
		PieceID:    pieceID,
		TextPrompt: newText,
	})
	if err != nil {
		log.Default().Println(err)
		htmx.TriggerAfterSettle(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not update reminders",
			Title:    "Database Error",
			Variant:  "error",
			Duration: 3000,
		})
		librarypages.EditRemindersSummary(spot.TextPrompt, spot.PieceID, spot.ID, csrf.Token(r), "Failed to Update").Render(r.Context(), w)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	htmx.Trigger(r, "UpdateSpotRemindersField", map[string]string{
		"id":   spot.ID,
		"text": newText,
	})
	librarypages.RemindersSummary(spot.TextPrompt, spot.PieceID, spot.ID).Render(r.Context(), w)
}

func (s *Server) getReminders(w http.ResponseWriter, r *http.Request) {
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

	librarypages.RemindersSummary(spot.TextPrompt, spot.PieceID, spot.ID).Render(r.Context(), w)
}

func (s *Server) getPracticeSpotDisplay(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	spotID := chi.URLParam(r, "spotID")
	pieceID := chi.URLParam(r, "pieceID")
	queries := db.New(s.DB)

	spot, err := queries.GetSpot(r.Context(), db.GetSpotParams{
		SpotID:  spotID,
		UserID:  user.ID,
		PieceID: pieceID,
	})
	if err != nil {
		// TODO: create a pretty 404 handler
		log.Default().Println(err)
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not find matching spot",
			Title:    "Error",
			Variant:  "error",
			Duration: 3000,
		})
		w.WriteHeader(http.StatusNotFound)
		return
	}

	spotJSON, err := json.Marshal(spot)
	if err != nil {
		log.Default().Println(err)
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Invalid Spot",
			Title:    "Error",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	librarypages.PracticeSpotDisplay(string(spotJSON), spot.PieceID, spot.PieceTitle).Render(r.Context(), w)
}
