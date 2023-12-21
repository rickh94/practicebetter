package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"practicebetter/internal/db"
	"practicebetter/internal/pages/librarypages"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/mavolin/go-htmx"
	"github.com/nrednav/cuid2"
)

func (s *Server) createPieceForm(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.CreatePiecePage(s, token), "Create Piece")
}

func (s *Server) createPiece(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	user := r.Context().Value("user").(db.User)
	tx, err := s.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	queries := db.New(s.DB)
	qtx := queries.WithTx(tx)

	m, err := strconv.Atoi(r.Form.Get("measures"))
	measures := sql.NullInt64{Int64: int64(m), Valid: true}
	if err != nil {
		measures = sql.NullInt64{Valid: false}
	}
	b, err := strconv.Atoi(r.Form.Get("beatsPerMeasure"))
	beatsPerMeasure := sql.NullInt64{Int64: int64(b), Valid: true}
	if err != nil {
		beatsPerMeasure = sql.NullInt64{Valid: false}
	}
	g, err := strconv.Atoi(r.Form.Get("goalTempo"))
	goalTempo := sql.NullInt64{Int64: int64(g), Valid: true}
	if err != nil {
		goalTempo = sql.NullInt64{Valid: false}
	}

	pieceID := cuid2.Generate()

	_, err = qtx.CreatePiece(r.Context(), db.CreatePieceParams{
		ID:              pieceID,
		Title:           r.Form.Get("title"),
		Description:     sql.NullString{String: r.Form.Get("description"), Valid: true},
		Composer:        sql.NullString{String: r.Form.Get("composer"), Valid: true},
		Measures:        measures,
		BeatsPerMeasure: beatsPerMeasure,
		GoalTempo:       goalTempo,
		UserID:          user.ID,
	})
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Failed to create piece", http.StatusBadRequest)
		return
	}
	for _, s := range r.Form["spots"] {
		var spot SpotFormData
		err = json.Unmarshal([]byte(s), &spot)
		if err != nil {
			log.Default().Println(err)
			continue
		}
		currentTempo := sql.NullInt64{Valid: false}
		if spot.CurrentTempo != nil && *spot.CurrentTempo > 0 {
			currentTempo = sql.NullInt64{Int64: *spot.CurrentTempo, Valid: true}
		}
		measures := sql.NullString{Valid: false}
		if spot.Measures != nil {
			measures = sql.NullString{String: *spot.Measures, Valid: true}
		}
		newSpotID := cuid2.Generate()
		_, err := qtx.CreateSpot(r.Context(), db.CreateSpotParams{
			UserID:         user.ID,
			PieceID:        pieceID,
			ID:             newSpotID,
			Name:           spot.Name,
			Idx:            *spot.Idx,
			Stage:          spot.Stage,
			AudioPromptUrl: spot.AudioPromptUrl,
			ImagePromptUrl: spot.ImagePromptUrl,
			NotesPrompt:    spot.NotesPrompt,
			TextPrompt:     spot.TextPrompt,
			CurrentTempo:   currentTempo,
			Measures:       measures,
		})
		if err != nil {
			log.Default().Println(err)
			http.Error(w, "Failed to create spot", http.StatusInternalServerError)
			return
		}
	}

	piece, err := qtx.GetPieceByID(r.Context(), db.GetPieceByIDParams{
		PieceID: pieceID,
		UserID:  user.ID,
	})
	if err != nil || len(piece) == 0 {
		// TODO: create a pretty 404 handler
		log.Default().Println(err)
		http.Error(w, "Could not find matching piece", http.StatusNotFound)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Database operation failed", http.StatusInternalServerError)
		return
	}
	token := csrf.Token(r)

	hxRequest := htmx.Request(r)
	if hxRequest == nil {
		s.Redirect(w, r, "/library/pieces/"+pieceID)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	htmx.PushURL(r, "/library/pieces/"+pieceID)
	htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Successfully added piece: " + piece[0].Title,
		Title:    "Piece Created!",
		Variant:  "success",
		Duration: 3000,
	})

	w.WriteHeader(http.StatusCreated)
	component := librarypages.SinglePiece(s, token, piece)
	component.Render(r.Context(), w)
	return
}

const piecesPerPage = 20

func (s *Server) pieces(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}
	pageNum, err := strconv.Atoi(page)
	if err != nil {
		pageNum = 1
	}
	queries := db.New(s.DB)
	pieces, err := queries.ListPaginatedUserPieces(r.Context(), db.ListPaginatedUserPiecesParams{
		UserID: user.ID,
		Limit:  piecesPerPage,
		Offset: int64((pageNum - 1) * piecesPerPage),
	})
	totalPieces, err := queries.CountUserPieces(r.Context(), user.ID)
	if err != nil {
		log.Default().Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	totalPages := int(math.Ceil(float64(totalPieces) / float64(piecesPerPage)))
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	s.HxRender(w, r, librarypages.PieceList(pieces, pageNum, totalPages), "Pieces")
}

func (s *Server) singlePiece(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	user := r.Context().Value("user").(db.User)
	queries := db.New(s.DB)
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
	sessions1, err := queries.ListRecentPracticeSessionsForPiece(r.Context(), db.ListRecentPracticeSessionsForPieceParams{
		UserID:  user.ID,
		PieceID: piece[0].ID,
	})
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	sessions2, err := queries.ListRecentPracticeSessionsForPieceSpots(r.Context(), db.ListRecentPracticeSessionsForPieceSpotsParams{
		UserID:  user.ID,
		PieceID: piece[0].ID,
	})

	for _, s := range sessions1 {
		log.Default().Println(s)
	}
	for _, s := range sessions2 {
		log.Default().Println(s)
	}

	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	s.HxRender(w, r, librarypages.SinglePiece(s, token, piece), piece[0].Title)
}

func (s *Server) deletePiece(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	user := r.Context().Value("user").(db.User)
	queries := db.New(s.DB)
	err := queries.DeletePiece(r.Context(), db.DeletePieceParams{
		ID:     pieceID,
		UserID: user.ID,
	})
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Could not delete piece", http.StatusBadRequest)
		return
	}

	pieces, err := queries.ListPaginatedUserPieces(r.Context(), db.ListPaginatedUserPiecesParams{
		UserID: user.ID,
		Limit:  piecesPerPage,
		Offset: 0,
	})
	totalPieces, err := queries.CountUserPieces(r.Context(), user.ID)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	totalPages := int(math.Ceil(float64(totalPieces) / float64(piecesPerPage)))
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	htmx.PushURL(r, "/library/pieces")
	htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Successfully deleted piece",
		Title:    "Piece Deleted!",
		Variant:  "success",
		Duration: 3000,
	})
	w.WriteHeader(http.StatusOK)
	s.HxRender(w, r, librarypages.PieceList(pieces, 1, totalPages), "Pieces")
}

func (s *Server) editPiece(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	user := r.Context().Value("user").(db.User)
	queries := db.New(s.DB)
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

	var pieceFormData PieceFormData
	pieceFormData.ID = piece[0].ID
	pieceFormData.Title = piece[0].Title
	if piece[0].Description.Valid {
		pieceFormData.Description = &piece[0].Description.String
	}
	if piece[0].Composer.Valid {
		pieceFormData.Composer = &piece[0].Composer.String
	}
	if piece[0].Measures.Valid && piece[0].Measures.Int64 > 0 {
		pieceFormData.Measures = &piece[0].Measures.Int64
	}
	if piece[0].BeatsPerMeasure.Valid && piece[0].BeatsPerMeasure.Int64 > 0 {
		pieceFormData.BeatsPerMeasure = &piece[0].BeatsPerMeasure.Int64
	}
	if piece[0].GoalTempo.Valid && piece[0].GoalTempo.Int64 > 0 {
		pieceFormData.GoalTempo = &piece[0].GoalTempo.Int64
	}

	var spotsFormData []SpotFormData

	for _, row := range piece {
		if row.SpotID.Valid {
			spotsFormData = append(spotsFormData, makeSpotFormDataFromRow(row))
		}
	}
	pieceJson, err := json.Marshal(pieceFormData)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	spotsJson, err := json.Marshal(spotsFormData)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.EditPiecePage(s, token, pieceFormData.Title, pieceID, string(pieceJson), string(spotsJson)), pieceFormData.Title)
}

func (s *Server) updatePiece(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	user := r.Context().Value("user").(db.User)
	tx, err := s.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	queries := db.New(s.DB)
	qtx := queries.WithTx(tx)

	m, err := strconv.Atoi(r.Form.Get("measures"))
	measures := sql.NullInt64{Int64: int64(m), Valid: true}
	log.Default().Println(m)
	log.Default().Println(measures)
	if err != nil {
		measures = sql.NullInt64{Valid: false}
	}
	b, err := strconv.Atoi(r.Form.Get("beatsPerMeasure"))
	beatsPerMeasure := sql.NullInt64{Int64: int64(b), Valid: true}
	if err != nil {
		beatsPerMeasure = sql.NullInt64{Valid: false}
	}
	// TODO: refactor some of this (and in create piece) to maybe avoid saving weird values in the database
	g, err := strconv.Atoi(r.Form.Get("goalTempo"))
	goalTempo := sql.NullInt64{Int64: int64(g), Valid: true}
	if err != nil {
		goalTempo = sql.NullInt64{Valid: false}
	}

	pieceID := chi.URLParam(r, "pieceID")
	_, err = qtx.UpdatePiece(r.Context(), db.UpdatePieceParams{
		ID:              pieceID,
		Title:           r.Form.Get("title"),
		Description:     sql.NullString{String: r.Form.Get("description"), Valid: true},
		Composer:        sql.NullString{String: r.Form.Get("composer"), Valid: true},
		Measures:        measures,
		BeatsPerMeasure: beatsPerMeasure,
		GoalTempo:       goalTempo,
		UserID:          user.ID,
		Stage:           r.Form.Get("stage"),
	})
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Could not update piece", http.StatusInternalServerError)
		return
	}
	var keepSpotIDs []string
	for _, s := range r.Form["spots"] {
		var spot SpotFormData
		err = json.Unmarshal([]byte(s), &spot)
		if err != nil {
			log.Default().Println(err)
			continue
		}
		currentTempo := sql.NullInt64{Valid: false}
		if spot.CurrentTempo != nil && *spot.CurrentTempo > 0 {
			currentTempo = sql.NullInt64{Int64: *spot.CurrentTempo, Valid: true}
		}
		measures := sql.NullString{Valid: false}
		if spot.Measures != nil {
			measures = sql.NullString{String: *spot.Measures, Valid: true}
		}
		if spot.ID != nil {
			keepSpotIDs = append(keepSpotIDs, *spot.ID)
			err := qtx.UpdateSpot(r.Context(), db.UpdateSpotParams{
				Name:           spot.Name,
				Idx:            *spot.Idx,
				Stage:          spot.Stage,
				AudioPromptUrl: spot.AudioPromptUrl,
				ImagePromptUrl: spot.ImagePromptUrl,
				NotesPrompt:    spot.NotesPrompt,
				TextPrompt:     spot.TextPrompt,
				CurrentTempo:   currentTempo,
				SpotID:         *spot.ID,
				UserID:         user.ID,
				PieceID:        pieceID,
				Measures:       measures,
			})
			if err != nil {
				log.Default().Println(err)
				http.Error(w, "Failed to update spot", http.StatusInternalServerError)
				return
			}
		} else {
			newSpotID := cuid2.Generate()
			_, err := qtx.CreateSpot(r.Context(), db.CreateSpotParams{
				UserID:         user.ID,
				PieceID:        pieceID,
				ID:             newSpotID,
				Name:           spot.Name,
				Idx:            *spot.Idx,
				Stage:          spot.Stage,
				AudioPromptUrl: spot.AudioPromptUrl,
				ImagePromptUrl: spot.ImagePromptUrl,
				NotesPrompt:    spot.NotesPrompt,
				TextPrompt:     spot.TextPrompt,
				CurrentTempo:   currentTempo,
				Measures:       measures,
			})
			if err != nil {
				log.Default().Println(err)
				http.Error(w, "Failed to create spot", http.StatusInternalServerError)
				return
			}
			keepSpotIDs = append(keepSpotIDs, newSpotID)
		}

	}
	err = qtx.DeleteSpotsExcept(r.Context(), db.DeleteSpotsExceptParams{
		SpotIDs: keepSpotIDs,
		UserID:  user.ID,
		PieceID: pieceID,
	})
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Failed to delete spots", http.StatusInternalServerError)
		return
	}

	piece, err := qtx.GetPieceByID(r.Context(), db.GetPieceByIDParams{
		PieceID: pieceID,
		UserID:  user.ID,
	})
	if err != nil || len(piece) == 0 {
		// TODO: create a pretty 404 handler
		log.Default().Println(err)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Could not find matching piece"))
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Default().Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to commit transaction"))
		return
	}
	token := csrf.Token(r)

	hxRequest := htmx.Request(r)
	if hxRequest == nil {
		s.Redirect(w, r, "/library/pieces/"+pieceID)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	htmx.PushURL(r, "/library/pieces/"+pieceID)
	htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Successfully updated piece: " + piece[0].Title,
		Title:    "Piece Updated!",
		Variant:  "success",
		Duration: 3000,
	})

	w.WriteHeader(http.StatusCreated)
	component := librarypages.SinglePiece(s, token, piece)
	component.Render(r.Context(), w)
	return
}

func (s *Server) piecePracticeRandomSpotsPage(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	user := r.Context().Value("user").(db.User)
	queries := db.New(s.DB)
	piece, err := queries.GetPieceWithRandomSpots(r.Context(), db.GetPieceWithRandomSpotsParams{
		PieceID: pieceID,
		UserID:  user.ID,
	})
	if err != nil || len(piece) == 0 {
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
		s.HxRender(w, r, librarypages.PiecePracticeNoSpotsPage(piece[0].Title, piece[0].ID), piece[0].Title)

		return
	}
	// TODO: unfuck this and get rid of the goddamn pointers
	var spots []SpotFormData
	for _, row := range piece {
		var measures *string
		if row.SpotMeasures.Valid {
			// Copy the value
			m := row.SpotMeasures.String
			measures = &m
		} else {
			measures = nil
		}
		var currentTempo *int64
		if row.SpotCurrentTempo.Valid {
			t := row.SpotCurrentTempo.Int64
			currentTempo = &t
		} else {
			currentTempo = nil
		}
		// row is a moving pointer, directly referencing underlying data is unreliable when
		// the pointer moves (the spots all ended up with the last spot's id). Need to make copies of the data to point to
		spotID := row.SpotID
		spotIdx := row.SpotIdx
		spots = append(spots, SpotFormData{
			ID:             &spotID,
			Name:           row.SpotName,
			Idx:            &spotIdx,
			Stage:          row.SpotStage,
			AudioPromptUrl: row.SpotAudioPromptUrl,
			ImagePromptUrl: row.SpotImagePromptUrl,
			NotesPrompt:    row.SpotNotesPrompt,
			TextPrompt:     row.SpotTextPrompt,
			CurrentTempo:   currentTempo,
			Measures:       measures,
		})
	}

	spotsData, err := json.Marshal(spots)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.PiecePracticeRandomSpotsPage(s, token, piece, string(spotsData)), piece[0].Title)
}

type PieceSpotsPracticeInfo struct {
	DurationMinutes int64    `json:"durationMinutes"`
	SpotIDs         []string `json:"spotIDs"`
}

func (s *Server) finishPracticePieceSpots(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	pieceID := chi.URLParam(r, "pieceID")
	var info PieceSpotsPracticeInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	queries := db.New(s.DB)
	tx, err := s.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	qtx := queries.WithTx(tx)

	var practiceSessionID string
	activePracticePlanID, err := s.GetActivePracticePlanID(r.Context())
	if err != nil {
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
	} else {
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
	}

	if activePracticePlanID != "" {
		if err := qtx.CompletePracticePlanPiece(r.Context(), db.CompletePracticePlanPieceParams{
			UserID:       user.ID,
			PlanID:       activePracticePlanID,
			PieceID:      pieceID,
			PracticeType: "random_spots",
		}); err != nil {
			log.Default().Println(err)
			http.Error(w, "Could not complete practice plan piece", http.StatusInternalServerError)
			return
		}

	}

	for _, spotID := range info.SpotIDs {
		if err := qtx.PracticeSpot(r.Context(), db.PracticeSpotParams{
			UserID:            user.ID,
			SpotID:            spotID,
			PracticeSessionID: practiceSessionID,
		}); err != nil {
			log.Default().Println(err)
			http.Error(w, "Could not practice spot", http.StatusInternalServerError)
			return
		}

		if err := qtx.UpdateSpotPracticed(r.Context(), db.UpdateSpotPracticedParams{
			SpotID: spotID,
			UserID: user.ID,
		}); err != nil {
			log.Default().Println(err)
			http.Error(w, "Could not update spot", http.StatusInternalServerError)
			return
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

func (s *Server) piecePracticeRandomSequencePage(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	user := r.Context().Value("user").(db.User)
	queries := db.New(s.DB)
	piece, err := queries.GetPieceWithRandomSpots(r.Context(), db.GetPieceWithRandomSpotsParams{
		PieceID: pieceID,
		UserID:  user.ID,
	})
	if err != nil || len(piece) == 0 {
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
		s.HxRender(w, r, librarypages.PiecePracticeNoSpotsPage(piece[0].Title, piece[0].ID), piece[0].Title)

		return
	}
	var spots []SpotFormData
	for _, row := range piece {
		var measures *string
		if row.SpotMeasures.Valid {
			measures = &row.SpotMeasures.String
		}
		var currentTempo *int64
		if row.SpotCurrentTempo.Valid {
			currentTempo = &row.SpotCurrentTempo.Int64
		}
		// row is a moving pointer, directly referencing underlying data is unreliable when
		// the pointer moves (the spots all ended up with the last spot's id). Need to make copies of the data to point to
		spotID := row.SpotID
		spotIdx := row.SpotIdx
		spots = append(spots, SpotFormData{
			ID:             &spotID,
			Name:           row.SpotName,
			Idx:            &spotIdx,
			Stage:          row.SpotStage,
			AudioPromptUrl: row.SpotAudioPromptUrl,
			ImagePromptUrl: row.SpotImagePromptUrl,
			NotesPrompt:    row.SpotNotesPrompt,
			TextPrompt:     row.SpotTextPrompt,
			CurrentTempo:   currentTempo,
			Measures:       measures,
		})
	}

	spotData, err := json.Marshal(spots)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.PiecePracticeRandomSequencePage(s, token, piece, string(spotData)), piece[0].Title)
}

func (s *Server) piecePracticeStartingPointPage(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	user := r.Context().Value("user").(db.User)
	queries := db.New(s.DB)
	piece, err := queries.GetPieceWithoutSpots(r.Context(), db.GetPieceWithoutSpotsParams{
		ID:     pieceID,
		UserID: user.ID,
	})
	if err != nil {
		// TODO: create a pretty 404 handler
		log.Default().Println(err)
		http.Error(w, "Could not find matching piece", http.StatusNotFound)
		return
	}
	if !piece.Measures.Valid || !piece.BeatsPerMeasure.Valid {
		s.HxRender(w, r, librarypages.PiecePracticeMissingMeasureInfoPage(piece.Title, piece.ID), piece.Title)
		return
	}

	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.PiecePracticeStartingPointPage(s, token, piece), piece.Title)
}

type PiecePracticeInfo struct {
	DurationMinutes   int64  `json:"durationMinutes"`
	MeasuresPracticed string `json:"measuresPracticed"`
}

func (s *Server) piecePracticeStartingPointFinished(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	user := r.Context().Value("user").(db.User)
	var info PiecePracticeInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	queries := db.New(s.DB)
	tx, err := s.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	qtx := queries.WithTx(tx)

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
	if err := qtx.PracticePiece(r.Context(), db.PracticePieceParams{
		UserID:            user.ID,
		PieceID:           pieceID,
		PracticeSessionID: practiceSessionID,
		Measures:          info.MeasuresPracticed,
	}); err != nil {
		log.Default().Println(err)
		http.Error(w, "Could not practice piece", http.StatusInternalServerError)
		return
	}
	if err := qtx.UpdatePiecePracticed(r.Context(), db.UpdatePiecePracticedParams{
		UserID:  user.ID,
		PieceID: pieceID,
	}); err != nil {
		log.Default().Println(err)
		http.Error(w, "Could not practice piece", http.StatusInternalServerError)
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

func (s *Server) piecePracticeRepeatPage(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	user := r.Context().Value("user").(db.User)
	queries := db.New(s.DB)
	piece, err := queries.GetPieceWithIncompleteSpots(r.Context(), db.GetPieceWithIncompleteSpotsParams{
		PieceID: pieceID,
		UserID:  user.ID,
	})
	if err != nil {
		// TODO: create a pretty 404 handler
		log.Default().Println(err)
		http.Error(w, "Could not find matching piece", http.StatusNotFound)
		return
	}

	s.HxRender(w, r, librarypages.PiecePracticeRepeatPage(piece), piece[0].Title)
}
