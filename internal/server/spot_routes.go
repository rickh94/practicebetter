package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"practicebetter/internal/ck"
	"practicebetter/internal/components"
	"practicebetter/internal/config"
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
	user := r.Context().Value(ck.UserKey).(db.User)
	queries := db.New(s.DB)

	spot, err := queries.GetSpot(r.Context(), db.GetSpotParams{
		SpotID:  spotID,
		UserID:  user.ID,
		PieceID: pieceID,
	})
	if err != nil {
		// TODO: create a pretty 404 handler
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not find matching spot",
			Title:    "Error",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.SingleSpot(s, spot, token), spot.Name+" - "+spot.PieceTitle)
}

func (s *Server) addSpotsFromPDFPage(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)
	pieceID := chi.URLParam(r, "pieceID")
	user := r.Context().Value(ck.UserKey).(db.User)
	queries := db.New(s.DB)
	piece, err := queries.GetPieceWithoutSpots(r.Context(), db.GetPieceWithoutSpotsParams{
		ID:     pieceID,
		UserID: user.ID,
	})
	if err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not find matching piece",
			Title:    "Not Found",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Could not find piece", http.StatusNotFound)
		return
	}
	s.HxRender(w, r, librarypages.AddSpotsFromPDFPage(s, token, piece.ID, piece.Title), "Add Spots from PDF")
}

func (s *Server) addSingleSpotPage(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)
	pieceID := chi.URLParam(r, "pieceID")
	user := r.Context().Value(ck.UserKey).(db.User)
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
	user := r.Context().Value(ck.UserKey).(db.User)
	pieceID := chi.URLParam(r, "pieceID")
	queries := db.New(s.DB)
	if err := r.ParseForm(); err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not get information from the form",
			Title:    "Invalid Form",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Invalid Form", http.StatusBadRequest)
		return
	}
	currentTempo := sql.NullInt64{Valid: false}
	currentTempoVal := r.FormValue("currentTempo")
	if currentTempoVal != "" && currentTempoVal != "null" {
		currentTempoInt, err := strconv.Atoi(currentTempoVal)
		if err != nil {
			log.Default().Println(err)
			if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
				Message:  "Invalid Current Tempo",
				Title:    "Error",
				Variant:  "error",
				Duration: 3000,
			}); err != nil {
				log.Default().Println(err)
			}
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
		Stage:          r.FormValue("stage"),
		AudioPromptUrl: r.FormValue("audioPromptUrl"),
		ImagePromptUrl: r.FormValue("imagePromptUrl"),
		NotesPrompt:    r.FormValue("notesPrompt"),
		TextPrompt:     r.FormValue("textPrompt"),
		CurrentTempo:   currentTempo,
		Measures:       measures,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not add spot")
		return
	}
	outMeasures := librarypages.SpotMeasuresOrEmpty(spot.Measures)
	if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Added Spot: " + spot.Name,
		Title:    "Spot Added!",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}
	w.WriteHeader(http.StatusCreated)
	if err := components.SmallSpotCard(spot.PieceID, spot.ID, spot.Name, outMeasures, spot.Stage).Render(r.Context(), w); err != nil {
		log.Default().Println(err)
		http.Error(w, "Render Error", http.StatusInternalServerError)
	}
}

const MAX_SPOTS_AT_ONCE = 100

func (s *Server) addSpotsFromPDF(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	pieceID := chi.URLParam(r, "pieceID")
	r.Body = http.MaxBytesReader(w, r.Body, config.MAX_UPLOAD_SIZE*MAX_SPOTS_AT_ONCE+1024)
	if err := r.ParseMultipartForm(config.MAX_UPLOAD_SIZE*MAX_SPOTS_AT_ONCE + 1024); err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Too many large spot images to process at once.",
			Title:    "Max Upload Exceeded",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	queries := db.New(s.DB)
	tx, err := s.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Default().Println(err)
		}
	}()

	qtx := queries.WithTx(tx)

	numSpots, err := strconv.Atoi(r.FormValue("numSpots"))
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Invalid number of spots", http.StatusBadRequest)
		return
	}

	hasErr := false
	for i := 0; i < numSpots; i++ {
		spotNameField := fmt.Sprintf("spots.%d.name", i)
		spotImageField := fmt.Sprintf("spots.%d.image", i)
		spotName := r.FormValue(spotNameField)
		if spotName == "" {
			continue
		}
		spotImageFile, spotImageHeader, err := r.FormFile(spotImageField)
		if err != nil {
			log.Default().Println(err)
			hasErr = true
			continue
		}
		spotImageHeader.Filename = fmt.Sprintf("%s-spot%d.png", pieceID, i)
		_, spotImageURL, err := s.saveImage(spotImageFile, spotImageHeader, user.ID)
		if err != nil {
			log.Default().Println(err)
			hasErr = true
			continue
		}
		spotID := cuid2.Generate()
		_, err = qtx.CreateSpot(r.Context(), db.CreateSpotParams{
			UserID:         user.ID,
			PieceID:        pieceID,
			ID:             spotID,
			Name:           spotName,
			Stage:          "repeat",
			AudioPromptUrl: "",
			ImagePromptUrl: spotImageURL,
			NotesPrompt:    "",
			TextPrompt:     "",
			CurrentTempo:   sql.NullInt64{},
			Measures:       sql.NullString{},
		})

		if err != nil {
			hasErr = true
			log.Default().Println(err)
			continue
		}
	}
	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit to database", http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(map[string]bool{"error": hasErr})
	if err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")

	if _, err = w.Write(body); err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func makeSpotFormDataFromSpot(row db.GetSpotRow) SpotFormData {
	var spot SpotFormData
	spot.ID = &row.ID
	spot.Name = row.Name
	spot.Stage = row.Stage
	spot.TextPrompt = row.TextPrompt
	spot.AudioPromptUrl = row.AudioPromptUrl
	spot.ImagePromptUrl = row.ImagePromptUrl
	spot.NotesPrompt = row.NotesPrompt
	if row.CurrentTempo.Valid && row.CurrentTempo.Int64 > 0 {
		spot.CurrentTempo = &row.CurrentTempo.Int64
	}
	if row.Measures.Valid {
		spot.Measures = &row.Measures.String
	}
	return spot
}

func (s *Server) editSpot(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	spotID := chi.URLParam(r, "spotID")
	user := r.Context().Value(ck.UserKey).(db.User)
	queries := db.New(s.DB)

	spot, err := queries.GetSpot(r.Context(), db.GetSpotParams{
		SpotID:  spotID,
		UserID:  user.ID,
		PieceID: pieceID,
	})
	if err != nil {
		// TODO: create a pretty 404 handler
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not find matching spot",
			Title:    "Error",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Not Found", http.StatusNotFound)
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
	user := r.Context().Value(ck.UserKey).(db.User)
	queries := db.New(s.DB)
	if err := r.ParseForm(); err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Invalid Form",
			Title:    "Invalid Input",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	currentTempo := sql.NullInt64{Valid: false}
	currentTempoVal := r.FormValue("currentTempo")
	if currentTempoVal != "" && currentTempoVal != "null" {
		currentTempoInt, err := strconv.Atoi(currentTempoVal)
		if err != nil {
			log.Default().Println(err)
			if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
				Message:  "Invalid Current Tempo",
				Title:    "Error",
				Variant:  "error",
				Duration: 3000,
			}); err != nil {
				log.Default().Println(err)
			}
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
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not update spot: " + err.Error(),
			Title:    "Error",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
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
		s.DatabaseError(w, r, err, "Could not update spot")
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
	if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "This spot has been updated with your new values",
		Title:    "Spot Updated!",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}
	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.SingleSpot(s, spot, token), spot.Name+" - "+spot.PieceTitle)
}

type UpdatedSpot struct {
	Name           string `json:"name"`
	Stage          string `json:"stage"`
	StageStarted   *int64 `json:"stageStarted,omitempty"`
	AudioPromptUrl string `json:"audioPromptUrl"`
	ImagePromptUrl string `json:"imagePromptUrl"`
	NotesPrompt    string `json:"notesPrompt"`
	TextPrompt     string `json:"textPrompt"`
	CurrentTempo   *int64 `json:"currentTempo,omitempty"`
	Measures       string `json:"measures,omitempty"`
	ID             string `json:"id"`
	PieceID        string `json:"pieceId"`
}

func (s *Server) updatePartialSpot(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	spotID := chi.URLParam(r, "spotID")
	user := r.Context().Value(ck.UserKey).(db.User)
	queries := db.New(s.DB)

	if err := r.ParseForm(); err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Invalid Form",
			Title:    "Invalid Input",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	currentTempo := sql.NullInt64{Valid: false}
	currentTempoVal := r.FormValue("currentTempo")
	if currentTempoVal != "" && currentTempoVal != "null" {
		currentTempoInt, err := strconv.Atoi(currentTempoVal)
		if err != nil {
			log.Default().Println(err)
			if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
				Message:  "Invalid Current Tempo",
				Title:    "Error",
				Variant:  "error",
				Duration: 3000,
			}); err != nil {
				log.Default().Println(err)
			}
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

	updatedSpot, err := queries.UpdatePartialSpot(r.Context(), db.UpdatePartialSpotParams{
		Name:         r.FormValue("name"),
		CurrentTempo: currentTempo,
		Measures:     measures,
		SpotID:       spotID,
		UserID:       user.ID,
		PieceID:      pieceID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not update spot")
		return
	}

	updatedSpotInfo := UpdatedSpot{
		Name:           updatedSpot.Name,
		Stage:          updatedSpot.Stage,
		AudioPromptUrl: updatedSpot.AudioPromptUrl,
		ImagePromptUrl: updatedSpot.ImagePromptUrl,
		NotesPrompt:    updatedSpot.NotesPrompt,
		TextPrompt:     updatedSpot.TextPrompt,
		ID:             updatedSpot.ID,
		PieceID:        updatedSpot.PieceID,
	}
	if updatedSpot.CurrentTempo.Valid {
		updatedSpotInfo.CurrentTempo = &updatedSpot.CurrentTempo.Int64
	}
	if updatedSpot.Measures.Valid {
		updatedSpotInfo.Measures = updatedSpot.Measures.String
	}
	if updatedSpot.StageStarted.Valid {
		updatedSpotInfo.StageStarted = &updatedSpot.StageStarted.Int64
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedSpotInfo); err != nil {
		log.Default().Println(err)
	}
}

func (s *Server) deleteSpot(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	spotID := chi.URLParam(r, "spotID")
	user := r.Context().Value(ck.UserKey).(db.User)
	queries := db.New(s.DB)
	err := queries.DeleteSpot(r.Context(), db.DeleteSpotParams{
		UserID:  user.ID,
		PieceID: pieceID,
		SpotID:  spotID,
	})
	if err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not delete spot.",
			Title:    "Delete Failed",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Delete Failed", http.StatusBadRequest)
		return
	}
	htmx.PushURL(r, "/library/pieces/"+pieceID)
	if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "This spot has been deleted",
		Title:    "Spot Deleted",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}

	s.renderPiece(w, r, pieceID, user.ID)
}

func (s *Server) repeatPracticeSpot(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	spotID := chi.URLParam(r, "spotID")
	user := r.Context().Value(ck.UserKey).(db.User)
	queries := db.New(s.DB)

	spot, err := queries.GetSpot(r.Context(), db.GetSpotParams{
		SpotID:  spotID,
		UserID:  user.ID,
		PieceID: pieceID,
	})
	if err != nil {
		// TODO: create a pretty 404 handler
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not find matching spot",
			Title:    "Not Found",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Not Found", http.StatusNotFound)
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
		Stage:          spot.Stage,
		AudioPromptUrl: spot.AudioPromptUrl,
		ImagePromptUrl: spot.ImagePromptUrl,
		NotesPrompt:    spot.NotesPrompt,
		TextPrompt:     spot.TextPrompt,
		CurrentTempo:   currentTempo,
		Measures:       measures,
	}

	spotJson, err := json.Marshal(spotData)
	if err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not get spot: " + err.Error(),
			Title:    "Error",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
	user := r.Context().Value(ck.UserKey).(db.User)

	queries := db.New(s.DB)
	tx, err := s.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Default().Println(err)
		}
	}()

	qtx := queries.WithTx(tx)

	var info RepeatPracticeInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: better error handling
	activePracticePlanID, ok := s.GetActivePracticePlanID(r.Context())

	if ok && activePracticePlanID != "" {
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
			if err := qtx.UpdatePiecePracticed(r.Context(), db.UpdatePiecePracticedParams{
				UserID:  user.ID,
				PieceID: pieceID,
			}); err != nil {
				s.DatabaseError(w, r, err, "Could not update piece practiced")
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
			if err := qtx.UpdatePiecePracticed(r.Context(), db.UpdatePiecePracticedParams{
				UserID:  user.ID,
				PieceID: pieceID,
			}); err != nil {
				s.DatabaseError(w, r, err, "Could not update piece practiced")
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
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Default().Println(err)
	}
}

func (s *Server) updateReminders(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	spotID := chi.URLParam(r, "spotID")
	user := r.Context().Value(ck.UserKey).(db.User)
	queries := db.New(s.DB)

	newText := r.FormValue("text")
	spot, err := queries.UpdateTextPrompt(r.Context(), db.UpdateTextPromptParams{
		SpotID:     spotID,
		UserID:     user.ID,
		PieceID:    pieceID,
		TextPrompt: newText,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not update reminders")
		log.Default().Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(spot); err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) updateSpotImage(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	spotID := chi.URLParam(r, "spotID")
	if pieceID == "" || spotID == "" {
		http.Error(w, "Missing pieceID or spotID", http.StatusBadRequest)
		return
	}
	user := r.Context().Value(ck.UserKey).(db.User)
	r.Body = http.MaxBytesReader(w, r.Body, config.MAX_UPLOAD_SIZE)
	if err := r.ParseMultipartForm(config.MAX_UPLOAD_SIZE); err != nil {
		log.Default().Println(err)
		http.Error(w, "The uploaded file is too big. Please choose an file that's less than 1MB in size", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "File Error", http.StatusBadRequest)
		if _, err := w.Write([]byte("Could not load file")); err != nil {
			log.Default().Println(err)
		}
		return
	}
	newFileName, newFileUrl, err := s.saveImage(file, fileHeader, user.ID)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "File Save Error", http.StatusInternalServerError)
		if _, err := w.Write([]byte("Could not save file")); err != nil {
			log.Default().Println(err)
		}
		return
	}

	queries := db.New(s.DB)

	if err := queries.UpdateImagePrompt(r.Context(), db.UpdateImagePromptParams{
		SpotID:         spotID,
		UserID:         user.ID,
		PieceID:        pieceID,
		ImagePromptUrl: newFileUrl,
	}); err != nil {
		s.DatabaseError(w, r, err, "Could not update image prompt")
		log.Default().Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	data := UploadedFileInfo{
		newFileName,
		newFileUrl,
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (s *Server) updateSpotAudio(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	spotID := chi.URLParam(r, "spotID")
	if pieceID == "" || spotID == "" {
		http.Error(w, "Missing pieceID or spotID", http.StatusBadRequest)
		return
	}
	user := r.Context().Value(ck.UserKey).(db.User)
	r.Body = http.MaxBytesReader(w, r.Body, config.MAX_UPLOAD_SIZE)
	if err := r.ParseMultipartForm(config.MAX_UPLOAD_SIZE); err != nil {
		log.Default().Println(err)
		http.Error(w, "The uploaded file is too big. Please choose an file that's less than 1MB in size", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "File Error", http.StatusBadRequest)
		if _, err := w.Write([]byte("Could not load file")); err != nil {
			log.Default().Println(err)
		}
		return
	}
	newFileName, newFileUrl, err := s.saveAudio(file, fileHeader, user.ID)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "File Save Error", http.StatusInternalServerError)
		if _, err := w.Write([]byte("Could not save file")); err != nil {
			log.Default().Println(err)
		}
		return
	}

	queries := db.New(s.DB)

	if err := queries.UpdateAudioPrompt(r.Context(), db.UpdateAudioPromptParams{
		SpotID:         spotID,
		UserID:         user.ID,
		PieceID:        pieceID,
		AudioPromptUrl: newFileUrl,
	}); err != nil {
		s.DatabaseError(w, r, err, "Could not update audio prompt")
		log.Default().Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	data := UploadedFileInfo{
		newFileName,
		newFileUrl,
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (s *Server) getPracticeSpotDisplay(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
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
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not find matching spot",
			Title:    "Not Found",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	spotJSON, err := json.Marshal(spot)
	if err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Invalid Spot",
			Title:    "Error",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := librarypages.PracticeSpotDisplay(string(spotJSON), spot.PieceID, spot.PieceTitle).Render(r.Context(), w); err != nil {
		log.Default().Println(err)
		http.Error(w, "Render Error", http.StatusInternalServerError)
	}
}
