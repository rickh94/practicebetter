package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"practicebetter/internal/db"
	"practicebetter/internal/pages/librarypages"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gorilla/csrf"
	"github.com/nrednav/cuid2"
)

func (s *Server) libraryDashboard(w http.ResponseWriter, r *http.Request) {
	queries := db.New(s.DB)
	user := r.Context().Value("user").(db.User)
	pieces, err := queries.ListRecentlyPracticedPieces(r.Context(), user.ID)
	practiceSessions, err := queries.GetRecentPracticeSessions(r.Context(), user.ID)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	for _, row := range practiceSessions {
		log.Default().Println(row)
	}
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	s.HxRender(w, r, librarypages.Dashboard(pieces), "Library")
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
	s.HxRender(w, r, librarypages.UploadAudioForm(token), "Upload Audio")

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
	s.HxRender(w, r, librarypages.UploadImageForm(token), "Upload Image")

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
	if row.Measures.Valid {
		spot.Measures = &row.Measures.String
	}
	return spot
}

// TODO: maybe add render or redirect function
