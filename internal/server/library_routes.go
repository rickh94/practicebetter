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

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/mavolin/go-htmx"
	"github.com/nrednav/cuid2"
)

func (s *Server) libraryDashboard(w http.ResponseWriter, r *http.Request) {
	queries := db.New(s.DB)
	user := r.Context().Value("user").(db.User)
	pieces, err := queries.ListRecentlyPracticedPieces(r.Context(), user.ID)
	if err != nil {
		log.Default().Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not get pieces"))
		return
	}
	s.HxRender(w, r, librarypages.Dashboard(pieces))
}

func (s *Server) createPieceForm(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.CreatePiecePage(s, token))
}

func (s *Server) createPiece(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	user := r.Context().Value("user").(db.User)
	tx, err := s.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to connect to database"))
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
	b, err := strconv.Atoi(r.Form.Get("beats_per_measure"))
	beatsPerMeasure := sql.NullInt64{Int64: int64(b), Valid: true}
	if err != nil {
		beatsPerMeasure = sql.NullInt64{Valid: false}
	}
	g, err := strconv.Atoi(r.Form.Get("goal_tempo"))
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
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to create piece"))
		return
	}
	for _, s := range r.Form["spots"] {
		var spot db.CreateSpotParams
		err = json.Unmarshal([]byte(s), &spot)
		if err != nil {
			log.Default().Println(err)
		}
		spot.PieceID = pieceID
		spot.UserID = user.ID
		spot.ID = cuid2.Generate()
		_, err = qtx.CreateSpot(r.Context(), spot)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to create spot"))
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
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Could not find matching piece"))
		return
	}

	err = tx.Commit()
	if err != nil {
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

const piecesPerPage = 10

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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	s.HxRender(w, r, librarypages.PieceList(pieces, pageNum, totalPages))
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
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Could not find matching piece"))
		return
	}
	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.SinglePiece(s, token, piece))
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
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Could not find matching piece"))
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	totalPages := int(math.Ceil(float64(totalPieces) / float64(piecesPerPage)))
	if err != nil {
		log.Default().Println(err)
		w.WriteHeader(http.StatusInternalServerError)
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
	s.HxRender(w, r, librarypages.PieceList(pieces, 1, totalPages))
}

type PieceFormData struct {
	ID              string  `json:"id"`
	Title           string  `json:"title"`
	Description     *string `json:"description,omitempty"`
	Composer        *string `json:"composer,omitempty"`
	Measures        *int64  `json:"measures,omitempty"`
	BeatsPerMeasure *int64  `json:"beatsPerMeasure,omitempty"`
	PracticeNotes   *string `json:"practiceNotes,omitempty"`
	GoalTempo       *int64  `json:"goalTempo,omitempty"`
}

type SpotFormData struct {
	ID             *string `json:"id,omitempty"`
	Name           string  `json:"name"`
	Idx            *int64  `json:"idx,omitempty"`
	Stage          string  `json:"stage"`
	Measures       *string `json:"measures,omitempty"`
	AudioPromptUrl string  `json:"audioPromptUrl,omitempty"`
	ImagePromptUrl string  `json:"imagePromptUrl,omitempty"`
	NotesPrompt    string  `json:"notesPrompt,omitempty"`
	TextPrompt     string  `json:"textPrompt,omitempty"`
	CurrentTempo   *int64  `json:"currentTempo,omitempty"`
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
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Could not find matching piece"))
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	spotsJson, err := json.Marshal(spotsFormData)
	if err != nil {
		log.Default().Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.EditPiecePage(s, token, pieceFormData.Title, pieceID, string(pieceJson), string(spotsJson)))
}

func makeSpotFormDataFromRow(row db.GetPieceByIDRow) SpotFormData {
	var spot SpotFormData
	spot.ID = &row.SpotID.String
	if row.SpotName.Valid {
		spot.Name = row.SpotName.String
	}
	if row.SpotIdx.Valid {
		spot.Idx = &row.SpotIdx.Int64
	}
	if row.SpotStage.Valid {
		spot.Stage = row.SpotStage.String
	}
	if row.SpotMeasures.Valid {
		spot.Measures = &row.SpotMeasures.String
	}
	if row.SpotTextPrompt.Valid {
		spot.TextPrompt = row.SpotTextPrompt.String
	}
	if row.SpotAudioPromptUrl.Valid {
		spot.AudioPromptUrl = row.SpotAudioPromptUrl.String
	}
	if row.SpotImagePromptUrl.Valid {
		spot.ImagePromptUrl = row.SpotImagePromptUrl.String
	}
	if row.SpotNotesPrompt.Valid {
		spot.NotesPrompt = row.SpotNotesPrompt.String
	}
	if row.SpotCurrentTempo.Valid && row.SpotCurrentTempo.Int64 > 0 {
		spot.CurrentTempo = &row.SpotCurrentTempo.Int64
	}
	return spot
}

func (s *Server) updatePiece(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	user := r.Context().Value("user").(db.User)
	tx, err := s.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to connect to database"))
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
	b, err := strconv.Atoi(r.Form.Get("beats_per_measure"))
	beatsPerMeasure := sql.NullInt64{Int64: int64(b), Valid: true}
	if err != nil {
		beatsPerMeasure = sql.NullInt64{Valid: false}
	}
	// TODO: refactor some of this (and in create piece) to maybe avoid saving weird values in the database
	g, err := strconv.Atoi(r.Form.Get("goal_tempo"))
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
	})
	if err != nil {
		log.Default().Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to create piece"))
		return
	}
	var keepSpotIDs []string
	for _, s := range r.Form["spots"] {
		var spot SpotFormData
		err = json.Unmarshal([]byte(s), &spot)
		if err != nil {
			log.Default().Println(err)
		}
		currentTempo := sql.NullInt64{Valid: false}
		if spot.CurrentTempo != nil && *spot.CurrentTempo > 0 {
			currentTempo = sql.NullInt64{Int64: *spot.CurrentTempo, Valid: true}
		}
		if spot.ID != nil {
			keepSpotIDs = append(keepSpotIDs, *spot.ID)
			_, err := qtx.UpdateSpot(r.Context(), db.UpdateSpotParams{
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
			})
			if err != nil {
				log.Default().Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Failed to update spot"))
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
			})
			if err != nil {
				log.Default().Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Failed to create spot"))
				return
			}
			keepSpotIDs = append(keepSpotIDs, newSpotID)
		}

	}
	log.Default().Println(keepSpotIDs)
	err = qtx.DeleteSpotsExcept(r.Context(), db.DeleteSpotsExceptParams{
		SpotIDs: keepSpotIDs,
		UserID:  user.ID,
		PieceID: pieceID,
	})
	if err != nil {
		log.Default().Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to delete spots"))
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
