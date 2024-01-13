package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"practicebetter/internal/ck"
	"practicebetter/internal/db"
	"practicebetter/internal/pages/librarypages"
	"strconv"

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
		log.Default().Println(err)
		http.Error(w, "Failed to create piece", http.StatusBadRequest)
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
		Message:  "Successfully added piece: " + piece.Title,
		Title:    "Piece Created!",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}

	w.WriteHeader(http.StatusCreated)
	// TODO: add spots message here for sure
	s.HxRender(w, r, librarypages.AddSpotPage(s, token, pieceID, piece.Title, make([]db.ListPieceSpotsRow, 0)), piece.Title)
}

const piecesPerPage = 20

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
		Limit:  piecesPerPage,
		Offset: int64((pageNum - 1) * piecesPerPage),
	})
	if err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not load pieces",
			Title:    "Database Error",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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
	/*
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
	*/
	breakdown := getSpotBreakdown(piece)

	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.SinglePiece(s, token, piece, breakdown), piece[0].Title)
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
		Limit:  piecesPerPage,
		Offset: 0,
	})
	if err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Something went wrong",
			Title:    "Database Error",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Database Error", http.StatusInternalServerError)
		return
	}
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
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not update piece in the database",
			Title:    "Database Error",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Could not update piece", http.StatusInternalServerError)
		return
	}

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

	w.Header().Set("Content-Type", "text/html")
	htmx.PushURL(r, "/library/pieces/"+pieceID)
	if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Successfully updated piece: " + piece[0].Title,
		Title:    "Piece Updated!",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}

	breakdown := getSpotBreakdown(piece)

	w.WriteHeader(http.StatusCreated)
	if err := librarypages.SinglePiece(s, token, piece, breakdown).Render(r.Context(), w); err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not save your changes",
			Title:    "Database Error",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
			log.Default().Println(err)
			http.Error(w, "Could not complete practice plan piece", http.StatusInternalServerError)
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
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not save your changes",
			Title:    "Database Error",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
