package server

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"practicebetter/internal/components"
	"practicebetter/internal/db"
	"practicebetter/internal/pages/librarypages"
	"strconv"

	"github.com/gabriel-vasile/mimetype"
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
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
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
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
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
		http.Error(w, "Could not find matching piece", http.StatusNotFound)
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

const MAX_UPLOAD_SIZE = 1024 * 1024 // 1MiB

func (s *Server) uploadAudio(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)
	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		log.Default().Println(err)
		http.Error(w, "The uploaded file is too big. Please choose an file that's less than 1MB in size", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer file.Close()

	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filetype := mimetype.Detect(buff)
	if !filetype.Is("audio/mpeg") {
		log.Default().Println(filetype)
		http.Error(w, "The provided file format is not allowed. Please upload an audio file in MP3 format", http.StatusBadRequest)
		return
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create the uploads folder if it doesn't
	// already exist
	h := sha256.New()
	h.Write([]byte(user.ID))
	userIDHash := hex.EncodeToString(h.Sum(nil))[:8]

	userAudioPath := path.Join(s.UploadsPath, userIDHash, "audio")
	err = os.MkdirAll(userAudioPath, os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a new file in the uploads directory
	newFileName := fmt.Sprintf("%s-%s", cuid2.Generate()[:5], fileHeader.Filename)
	newFilePath := path.Join(userAudioPath, newFileName)

	dst, err := os.Create(newFilePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer dst.Close()

	// Copy the uploaded file to the filesystem
	// at the specified destination
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	data := map[string]string{
		"filename": newFileName,
		"url":      fmt.Sprintf("/uploads/%s/audio/%s", userIDHash, newFileName),
	}
	json.NewEncoder(w).Encode(data)
}

func (s *Server) uploadAudioForm(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.UploadAudioForm(token))

}

func (s *Server) uploadImage(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)
	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		log.Default().Println(err)
		http.Error(w, "The uploaded file is too big. Please choose an file that's less than 1MB in size", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer file.Close()

	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filetype := mimetype.Detect(buff)
	if !mimetype.EqualsAny(filetype.String(), "image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp") {
		log.Default().Println(filetype)
		http.Error(w, "The provided file format is not allowed. Please upload an image file.", http.StatusBadRequest)
		return
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create the uploads folder if it doesn't
	// already exist
	h := sha256.New()
	h.Write([]byte(user.ID))
	userIDHash := hex.EncodeToString(h.Sum(nil))[:8]

	userImagePath := path.Join(s.UploadsPath, userIDHash, "images")
	err = os.MkdirAll(userImagePath, os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a new file in the uploads directory
	newFileName := fmt.Sprintf("%s-%s", cuid2.Generate()[:5], fileHeader.Filename)
	newFilePath := path.Join(userImagePath, newFileName)

	dst, err := os.Create(newFilePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer dst.Close()

	// Copy the uploaded file to the filesystem
	// at the specified destination
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	data := map[string]string{
		"filename": newFileName,
		"url":      fmt.Sprintf("/uploads/%s/images/%s", userIDHash, newFileName),
	}
	json.NewEncoder(w).Encode(data)
}

func (s *Server) uploadImageForm(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.UploadImageForm(token))

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

func makeSpotFormDataFromSpot(row db.GetSpotRow) SpotFormData {
	var spot SpotFormData
	spot.ID = &row.ID
	spot.Name = row.Name
	spot.Idx = &row.Idx
	spot.Stage = row.Stage
	spot.TextPrompt = row.TextPrompt
	spot.AudioPromptUrl = row.AudioPromptUrl
	spot.ImagePromptUrl = row.ImagePromptUrl
	spot.NotesPrompt = row.NotesPrompt
	if row.CurrentTempo.Valid && row.CurrentTempo.Int64 > 0 {
		spot.CurrentTempo = &row.CurrentTempo.Int64
	}
	return spot
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

// TODO: maybe add render or redirect function
