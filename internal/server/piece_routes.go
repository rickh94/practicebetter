package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"practicebetter/internal/ck"
	"practicebetter/internal/config"
	"practicebetter/internal/db"
	"practicebetter/internal/pages/librarypages"
	"strconv"
	"strings"

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
	user := r.Context().Value(ck.UserKey).(db.User)
	queries := db.New(s.DB)

	if err := r.ParseForm(); err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Variant:  "error",
			Title:    "Invalid Data",
			Message:  "Your submission was invalid",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m, err := strconv.Atoi(r.Form.Get("measures"))
	measures := sql.NullInt64{Int64: int64(m), Valid: true}
	if err != nil {
		measures = sql.NullInt64{Valid: false}
	}
	b, err := strconv.Atoi(r.Form.Get("beats"))
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

	piece, err := queries.CreatePiece(r.Context(), db.CreatePieceParams{
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
		s.DatabaseError(w, r, err, "Failed to create piece")
		return
	}

	token := csrf.Token(r)

	hxRequest := htmx.Request(r)
	if hxRequest == nil {
		s.Redirect(w, r, "/library/pieces/"+pieceID+"/spots/add")
		return
	}

	w.Header().Set("Content-Type", "text/html")
	htmx.PushURL(r, "/library/pieces/"+pieceID+"/spots/add")
	if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Successfully added piece: " + piece.Title + ". Now add some spots and start practicing!",
		Title:    "Piece Created",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}

	w.WriteHeader(http.StatusCreated)
	s.HxRender(w, r, librarypages.AddSpotsFromPDFPage(s, token, pieceID, piece.Title), piece.Title)
}

func (s *Server) pieces(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
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
		Limit:  config.ItemsPerPage,
		Offset: int64((pageNum - 1) * config.ItemsPerPage),
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not load pieces")
		return
	}
	totalPieces, err := queries.CountUserPieces(r.Context(), user.ID)
	if err != nil {
		log.Default().Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	totalPages := int(math.Ceil(float64(totalPieces) / float64(config.ItemsPerPage)))
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
	user := r.Context().Value(ck.UserKey).(db.User)

	s.renderPiece(w, r, pieceID, user.ID)
}

func (s *Server) renderPiece(w http.ResponseWriter, r *http.Request, pieceID string, userID string) {
	queries := db.New(s.DB)
	piece, err := queries.GetPieceByID(r.Context(), db.GetPieceByIDParams{
		PieceID: pieceID,
		UserID:  userID,
	})
	if err != nil || len(piece) == 0 {
		// TODO: create a pretty 404 handler
		log.Default().Println(err)
		http.Error(w, "Could not find matching piece", http.StatusNotFound)
		return
	}
	pieceInfo := librarypages.SinglePieceInfo{
		Title:           piece[0].Title,
		ID:              pieceID,
		Composer:        piece[0].Composer,
		Measures:        piece[0].Measures,
		BeatsPerMeasure: piece[0].BeatsPerMeasure,
		GoalTempo:       piece[0].GoalTempo,
		LastPracticed:   piece[0].LastPracticed,
		Stage:           piece[0].Stage,
		SpotBreakdown: librarypages.PieceSpotsBreakdown{
			Repeat:      0,
			ExtraRepeat: 0,
			Random:      0,
			Interleave:  0,
			Infrequent:  0,
			Completed:   0,
		},
		Spots: make([]librarypages.PiecePageSpot, 0, len(piece)),
	}
	for _, row := range piece {
		if !row.SpotID.Valid || !row.SpotStage.Valid || !row.SpotName.Valid {
			continue
		}
		spotInfo := librarypages.PiecePageSpot{
			ID:       row.SpotID.String,
			Name:     row.SpotName.String,
			Measures: "",
			Stage:    row.SpotStage.String,
		}
		if row.SpotMeasures.Valid {
			spotInfo.Measures = row.SpotMeasures.String
		}
		switch row.SpotStage.String {
		case "repeat":
			pieceInfo.SpotBreakdown.Repeat++
		case "extra_repeat":
			pieceInfo.SpotBreakdown.ExtraRepeat++
		case "random":
			pieceInfo.SpotBreakdown.Random++
		case "interleave":
			pieceInfo.SpotBreakdown.Interleave++
		case "interleave_days":
			pieceInfo.SpotBreakdown.Infrequent++
		case "completed":
			pieceInfo.SpotBreakdown.Completed++
		}
		if row.SpotLastPracticed.Valid &&
			(!pieceInfo.LastPracticed.Valid ||
				row.SpotLastPracticed.Int64 > pieceInfo.LastPracticed.Int64) {
			pieceInfo.LastPracticed = sql.NullInt64{
				Int64: row.SpotLastPracticed.Int64,
				Valid: true,
			}
		}
		pieceInfo.Spots = append(pieceInfo.Spots, spotInfo)

	}
	log.Default().Println(pieceInfo.LastPracticed.Int64)
	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.SinglePiece(s, pieceInfo, token), pieceInfo.Title)
}

func getSpotBreakdown(piece []db.GetPieceByIDRow) librarypages.PieceSpotsBreakdown {
	breakdown := librarypages.PieceSpotsBreakdown{
		Repeat:      0,
		ExtraRepeat: 0,
		Random:      0,
		Interleave:  0,
		Infrequent:  0,
		Completed:   0,
	}
	for _, row := range piece {
		if row.SpotID.Valid && row.SpotStage.Valid {
			switch row.SpotStage.String {
			case "repeat":
				breakdown.Repeat++
			case "extra_repeat":
				breakdown.ExtraRepeat++
			case "random":
				breakdown.Random++
			case "interleave":
				breakdown.Interleave++
			case "interleave_days":
				breakdown.Infrequent++
			case "completed":
				breakdown.Completed++
			}
		}
	}
	return breakdown
}

func (s *Server) pieceSpots(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	user := r.Context().Value(ck.UserKey).(db.User)
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
	info := librarypages.ListSpotsInfo{
		PieceTitle:          piece[0].Title,
		PieceID:             piece[0].ID,
		RepeatSpots:         make([]librarypages.ListSpotsSpot, 0, len(piece)/2),
		ExtraRepeatSpots:    make([]librarypages.ListSpotsSpot, 0, len(piece)/2),
		RandomSpots:         make([]librarypages.ListSpotsSpot, 0, len(piece)/2),
		InterleaveSpots:     make([]librarypages.ListSpotsSpot, 0, len(piece)/2),
		InterleaveDaysSpots: make([]librarypages.ListSpotsSpot, 0, len(piece)/2),
		CompletedSpots:      make([]librarypages.ListSpotsSpot, 0, len(piece)/2),
	}
	for _, row := range piece {
		if !row.SpotID.Valid {
			continue
		}
		if !row.SpotStage.Valid {
			continue
		}
		if !row.SpotName.Valid {
			continue
		}
		var measures string
		if row.SpotMeasures.Valid {
			measures = row.SpotMeasures.String
		} else {
			measures = ""
		}
		switch row.SpotStage.String {
		case "repeat":
			info.RepeatSpots = append(info.RepeatSpots, librarypages.ListSpotsSpot{
				ID:       row.SpotID.String,
				Name:     row.SpotName.String,
				Measures: measures,
				Stage:    row.SpotStage.String,
			})
		case "extra_repeat":
			info.ExtraRepeatSpots = append(info.ExtraRepeatSpots, librarypages.ListSpotsSpot{
				ID:       row.SpotID.String,
				Name:     row.SpotName.String,
				Measures: measures,
				Stage:    row.SpotStage.String,
			})
		case "random":
			info.RandomSpots = append(info.RandomSpots, librarypages.ListSpotsSpot{
				ID:       row.SpotID.String,
				Name:     row.SpotName.String,
				Measures: measures,
				Stage:    row.SpotStage.String,
			})
		case "interleave":
			info.InterleaveSpots = append(info.InterleaveSpots, librarypages.ListSpotsSpot{
				ID:       row.SpotID.String,
				Name:     row.SpotName.String,
				Measures: measures,
				Stage:    row.SpotStage.String,
			})
		case "interleave_days":
			info.InterleaveDaysSpots = append(info.InterleaveDaysSpots, librarypages.ListSpotsSpot{
				ID:       row.SpotID.String,
				Name:     row.SpotName.String,
				Measures: measures,
				Stage:    row.SpotStage.String,
			})
		case "completed":
			info.CompletedSpots = append(info.CompletedSpots, librarypages.ListSpotsSpot{
				ID:       row.SpotID.String,
				Name:     row.SpotName.String,
				Measures: measures,
				Stage:    row.SpotStage.String,
			})
		}
	}
	s.HxRender(w, r, librarypages.ListSpots(s, token, info), info.PieceTitle)
}

func (s *Server) deletePiece(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	user := r.Context().Value(ck.UserKey).(db.User)
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
		Limit:  config.ItemsPerPage,
		Offset: 0,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not find matching piece")
		return
	}
	totalPieces, err := queries.CountUserPieces(r.Context(), user.ID)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	totalPages := int(math.Ceil(float64(totalPieces) / float64(config.ItemsPerPage)))
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	htmx.PushURL(r, "/library/pieces")
	if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Successfully deleted piece",
		Title:    "Piece Deleted!",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}
	w.WriteHeader(http.StatusOK)
	s.HxRender(w, r, librarypages.PieceList(pieces, 1, totalPages), "Pieces")
}

func (s *Server) editPiece(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	user := r.Context().Value(ck.UserKey).(db.User)
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

	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.EditPiecePage(s, token, piece), piece.Title)
}

func (s *Server) updatePiece(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)

	if err := r.ParseForm(); err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Invalid data",
			Title:    "Form Error",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	queries := db.New(s.DB)

	m, err := strconv.Atoi(r.Form.Get("measures"))
	measures := sql.NullInt64{Int64: int64(m), Valid: true}
	if err != nil {
		measures = sql.NullInt64{Valid: false}
	}
	b, err := strconv.Atoi(r.Form.Get("beats"))
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
	_, err = queries.UpdatePiece(r.Context(), db.UpdatePieceParams{
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
		s.DatabaseError(w, r, err, "Could not update piece")
		return
	}

	w.Header().Set("Content-Type", "text/html")
	htmx.PushURL(r, "/library/pieces/"+pieceID)
	if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Successfully updated piece: " + r.Form.Get("title"),
		Title:    "Piece Updated!",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}

	s.renderPiece(w, r, pieceID, user.ID)
}

func (s *Server) piecePracticeRandomSpotsPage(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	user := r.Context().Value(ck.UserKey).(db.User)
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
		var stageStarted *int64
		if row.SpotStageStarted.Valid {
			t := row.SpotStageStarted.Int64
			stageStarted = &t
		} else {
			stageStarted = nil
		}
		// row is a moving pointer, directly referencing underlying data is unreliable when
		// the pointer moves (the spots all ended up with the last spot's id). Need to make copies of the data to point to
		spotID := row.SpotID
		spots = append(spots, SpotFormData{
			ID:             &spotID,
			Name:           row.SpotName,
			Stage:          row.SpotStage,
			AudioPromptUrl: row.SpotAudioPromptUrl,
			ImagePromptUrl: row.SpotImagePromptUrl,
			NotesPrompt:    row.SpotNotesPrompt,
			TextPrompt:     row.SpotTextPrompt,
			CurrentTempo:   currentTempo,
			Measures:       measures,
			StageStarted:   stageStarted,
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

type PracticeSpot struct {
	ID      string `json:"id"`
	Promote bool   `json:"promote"`
	Demote  bool   `json:"demote"`
}

type PieceSpotsPracticeInfo struct {
	DurationMinutes int64          `json:"durationMinutes"`
	Spots           []PracticeSpot `json:"Spots"`
}

func (s *Server) finishPracticePieceSpots(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
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
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Default().Println(err)
		}
	}()

	qtx := queries.WithTx(tx)

	activePracticePlanID, ok := s.GetActivePracticePlanID(r.Context())
	if ok && activePracticePlanID != "" {
		log.Default().Println("Completing practice plan piece")
		if err := qtx.CompletePracticePlanPiece(r.Context(), db.CompletePracticePlanPieceParams{
			UserID:       user.ID,
			PlanID:       activePracticePlanID,
			PieceID:      pieceID,
			PracticeType: "random_spots",
		}); err != nil {
			s.DatabaseError(w, r, err, "Could not complete practice plan piece")
			return
		}
		if err := qtx.UpdatePlanLastPracticed(r.Context(), db.UpdatePlanLastPracticedParams{
			ID:     activePracticePlanID,
			UserID: user.ID,
		}); err != nil {
			s.DatabaseError(w, r, err, "Could not update plan last practiced")
			return
		}
	}

	for _, spot := range info.Spots {
		if spot.Promote {
			if err := qtx.PromoteSpotToInterleave(r.Context(), db.PromoteSpotToInterleaveParams{
				SpotID:  spot.ID,
				UserID:  user.ID,
				PieceID: pieceID,
			}); err != nil {
				log.Default().Println(err)
				http.Error(w, "Could not promote spot", http.StatusInternalServerError)
				return
			}
		} else if spot.Demote {
			if err := qtx.DemoteSpotToExtraRepeat(r.Context(), db.DemoteSpotToExtraRepeatParams{
				SpotID:  spot.ID,
				UserID:  user.ID,
				PieceID: pieceID,
			}); err != nil {
				log.Default().Println(err)
				http.Error(w, "Could not demote spot", http.StatusInternalServerError)
				return
			}
		} else {
			if err := qtx.UpdateSpotPracticed(r.Context(), db.UpdateSpotPracticedParams{
				SpotID: spot.ID,
				UserID: user.ID,
			}); err != nil {
				log.Default().Println(err)
				http.Error(w, "Could not update spot", http.StatusInternalServerError)
				return
			}
		}

	}
	if err := tx.Commit(); err != nil {
		s.DatabaseError(w, r, err, "Could not save changes")
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Default().Println(err)
		http.Error(w, "Could not write response", http.StatusInternalServerError)
	}
}

func (s *Server) piecePracticeStartingPointPage(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	user := r.Context().Value(ck.UserKey).(db.User)
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
	user := r.Context().Value(ck.UserKey).(db.User)
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
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Default().Println(err)
		}
	}()

	qtx := queries.WithTx(tx)
	activePracticePlanID, ok := s.GetActivePracticePlanID(r.Context())
	if ok && activePracticePlanID != "" {
		if err := qtx.CompletePracticePlanPiece(r.Context(), db.CompletePracticePlanPieceParams{
			UserID:       user.ID,
			PlanID:       activePracticePlanID,
			PieceID:      pieceID,
			PracticeType: "starting_point",
		}); err != nil {
			s.DatabaseError(w, r, err, "Could not complete practice plan piece")
			return
		}

		if err := qtx.UpdatePlanLastPracticed(r.Context(), db.UpdatePlanLastPracticedParams{
			ID:     activePracticePlanID,
			UserID: user.ID,
		}); err != nil {
			s.DatabaseError(w, r, err, "Could not update plan last practiced")
			return
		}

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
		s.DatabaseError(w, r, err, "Could not save changes")
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Default().Println(err)
		http.Error(w, "Could not write response", http.StatusInternalServerError)
	}
}

func (s *Server) piecePracticeRepeatPage(w http.ResponseWriter, r *http.Request) {
	pieceID := chi.URLParam(r, "pieceID")
	user := r.Context().Value(ck.UserKey).(db.User)
	queries := db.New(s.DB)
	piece, err := queries.GetPieceWithIncompleteSpots(r.Context(), db.GetPieceWithIncompleteSpotsParams{
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
		s.HxRender(w, r, librarypages.PieceRepeatPracticeNoSpotsPage(piece[0].Title, piece[0].ID), piece[0].Title)

		return
	}
	if err != nil {
		// TODO: create a pretty 404 handler
		log.Default().Println(err)
		http.Error(w, "Could not find matching piece", http.StatusNotFound)
		return
	}

	s.HxRender(w, r, librarypages.PiecePracticeRepeatPage(piece), piece[0].Title)
}

type ImportExportSpot struct {
	Name           string `json:"name"`
	Stage          string `json:"stage"`
	Measures       string `json:"measures,omitempty"`
	AudioPromptUrl string `json:"audioPromptUrl,omitempty"`
	ImagePromptUrl string `json:"imagePromptUrl,omitempty"`
	NotesPrompt    string `json:"notesPrompt,omitempty"`
	TextPrompt     string `json:"textPrompt,omitempty"`
	CurrentTempo   int64  `json:"currentTempo"`
	Priority       int64  `json:"priority"`
}

type ImportExportPiece struct {
	Title           string             `json:"title"`
	Description     string             `json:"description,omitempty"`
	Composer        string             `json:"composer,omitempty"`
	Measures        int64              `json:"measures"`
	BeatsPerMeasure int64              `json:"beatsPerMeasure"`
	GoalTempo       int64              `json:"goalTempo"`
	Stage           string             `json:"stage"`
	Spots           []ImportExportSpot `json:"spots"`
}

func (s *Server) exportPiece(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	pieceID := chi.URLParam(r, "pieceID")
	queries := db.New(s.DB)
	piece, err := queries.GetPieceByID(r.Context(), db.GetPieceByIDParams{
		PieceID: pieceID,
		UserID:  user.ID,
	})
	if err != nil || len(piece) == 0 {
		if err != nil || len(piece) == 0 {
			// TODO: create a pretty 404 handler
			log.Default().Println(err)
			if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
				Message:  "Could not find matching piece",
				Title:    "Not Found",
				Variant:  "error",
				Duration: 3000,
			}); err != nil {
				log.Default().Println(err)
			}
			http.Error(w, "Could not find matching piece", http.StatusNotFound)
			return
		}
	}
	exportPiece := ImportExportPiece{
		Title: piece[0].Title,
		Stage: piece[0].Stage,
		Spots: make([]ImportExportSpot, 0, len(piece)),
	}
	if piece[0].Description.Valid {
		exportPiece.Description = piece[0].Description.String
	} else {
		exportPiece.Description = ""
	}
	if piece[0].Composer.Valid {
		exportPiece.Composer = piece[0].Composer.String
	} else {
		exportPiece.Composer = ""
	}
	if piece[0].Measures.Valid {
		exportPiece.Measures = piece[0].Measures.Int64
	} else {
		exportPiece.Measures = 0
	}
	if piece[0].GoalTempo.Valid {
		exportPiece.GoalTempo = piece[0].GoalTempo.Int64
	} else {
		exportPiece.GoalTempo = 0
	}
	if piece[0].BeatsPerMeasure.Valid {
		exportPiece.BeatsPerMeasure = piece[0].BeatsPerMeasure.Int64
	} else {
		exportPiece.BeatsPerMeasure = 0
	}
	for _, row := range piece {
		if !row.SpotID.Valid || !row.SpotName.Valid {
			continue
		}
		exportSpot := ImportExportSpot{
			Name:           row.SpotName.String,
			Stage:          "repeat",
			Measures:       "",
			AudioPromptUrl: "",
			ImagePromptUrl: "",
			NotesPrompt:    "",
			TextPrompt:     "",
			CurrentTempo:   0,
			Priority:       0,
		}
		if row.SpotMeasures.Valid {
			exportSpot.Measures = row.SpotMeasures.String
		}
		if row.SpotAudioPromptUrl.Valid && row.SpotAudioPromptUrl.String != "" {
			if strings.Contains(row.SpotAudioPromptUrl.String, "https://") ||
				strings.Contains(row.SpotAudioPromptUrl.String, "http://") {
				exportSpot.AudioPromptUrl = row.SpotAudioPromptUrl.String
			} else {
				exportSpot.AudioPromptUrl = "https://" + s.Hostname + row.SpotAudioPromptUrl.String
			}
		}
		if row.SpotImagePromptUrl.Valid && row.SpotImagePromptUrl.String != "" {
			if strings.Contains(row.SpotImagePromptUrl.String, "https://") ||
				strings.Contains(row.SpotImagePromptUrl.String, "http://") {
				exportSpot.ImagePromptUrl = row.SpotImagePromptUrl.String
			} else {
				exportSpot.ImagePromptUrl = "https://" + s.Hostname + row.SpotImagePromptUrl.String
			}
		}
		if row.SpotNotesPrompt.Valid {
			exportSpot.NotesPrompt = row.SpotNotesPrompt.String
		}
		if row.SpotTextPrompt.Valid {
			exportSpot.TextPrompt = row.SpotTextPrompt.String
		}
		if row.SpotCurrentTempo.Valid {
			exportSpot.CurrentTempo = row.SpotCurrentTempo.Int64
		}
		exportPiece.Spots = append(exportPiece.Spots, exportSpot)
	}

	jsonBytes, err := json.Marshal(exportPiece)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="`+exportPiece.Title+`.json"`)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(jsonBytes); err != nil {
		log.Default().Println(err)
	}
}

func (s *Server) importPiece(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)

	url := r.URL.Query().Get("url")
	if url != "" {
		resp, err := http.Get(url)
		if err != nil {
			log.Default().Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer resp.Body.Close()
		var p ImportExportPiece
		if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
			log.Default().Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		pieceID, err := s.createPieceWithSpots(r.Context(), p, user.ID)
		if err != nil {
			log.Default().Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		// 	Message:  "Your piece has been imported and you can start practicing!",
		// 	Title:    "Imported Piece",
		// 	Variant:  "success",
		// 	Duration: 3000,
		// }); err != nil {
		// 	log.Default().Println(err)
		// }
		//
		// htmx.PushURL(r, "/library/pieces/"+pieceID)
		// TODO: show an alert
		http.Redirect(w, r, "/library/pieces/"+pieceID, http.StatusSeeOther)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}

func (s *Server) createPieceWithSpots(ctx context.Context, p ImportExportPiece, userID string) (string, error) {
	queries := db.New(s.DB)

	measures := sql.NullInt64{Int64: 0, Valid: false}
	if p.Measures > 0 {
		measures = sql.NullInt64{Int64: p.Measures, Valid: true}
	}

	beatsPerMeasure := sql.NullInt64{Int64: 0, Valid: false}
	if p.BeatsPerMeasure > 0 {
		beatsPerMeasure = sql.NullInt64{Int64: p.BeatsPerMeasure, Valid: true}
	}
	goalTempo := sql.NullInt64{Int64: 0, Valid: false}
	if p.GoalTempo > 0 {
		goalTempo = sql.NullInt64{Int64: p.GoalTempo, Valid: true}
	}
	description := sql.NullString{String: "", Valid: false}
	if p.Description != "" {
		description = sql.NullString{String: p.Description, Valid: true}
	}
	composer := sql.NullString{String: "", Valid: false}
	if p.Composer != "" {
		composer = sql.NullString{String: p.Composer, Valid: true}
	}

	pieceID := cuid2.Generate()

	piece, err := queries.CreatePiece(ctx, db.CreatePieceParams{
		ID:              pieceID,
		Title:           p.Title,
		Description:     description,
		Composer:        composer,
		Measures:        measures,
		BeatsPerMeasure: beatsPerMeasure,
		GoalTempo:       goalTempo,
		UserID:          userID,
	})
	if err != nil {
		return "", err
	}

	for _, spot := range p.Spots {
		_, err := queries.CreateSpot(ctx, db.CreateSpotParams{
			ID:          cuid2.Generate(),
			Name:        spot.Name,
			Stage:       spot.Stage,
			NotesPrompt: spot.NotesPrompt,
			TextPrompt:  spot.TextPrompt,
			CurrentTempo: sql.NullInt64{
				Int64: spot.CurrentTempo,
				Valid: true,
			},
			Measures: sql.NullString{
				String: spot.Measures,
				Valid:  true,
			},
			AudioPromptUrl: spot.AudioPromptUrl,
			ImagePromptUrl: spot.ImagePromptUrl,
			PieceID:        pieceID,
			UserID:         userID,
		})
		if err != nil {
			return "", err
		}
	}
	return piece.ID, nil
}

func (s *Server) uploadPieceFile(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.UploadPieceFileForm(token), "Upload Piece File")

}

func (s *Server) importPieceFromFile(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "invalid file", http.StatusBadRequest)
		return
	}
	if fileHeader.Filename[len(fileHeader.Filename)-4:] != "json" {
		log.Default().Println(err)
		http.Error(w, "The file must be a JSON file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	var p ImportExportPiece
	if err := json.NewDecoder(file).Decode(&p); err != nil {
		log.Default().Println(err)
		http.Error(w, "Could not decode file", http.StatusBadRequest)
		return
	}

	pieceID, err := s.createPieceWithSpots(r.Context(), p, user.ID)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Added Piece: " + p.Title,
		Title:    "Piece Imported!",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}
	htmx.Redirect(r, "/library/pieces/"+pieceID)
	http.Redirect(w, r, "/library/pieces/"+pieceID, http.StatusSeeOther)

}
