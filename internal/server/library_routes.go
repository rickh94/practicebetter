package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"practicebetter/internal/ck"
	"practicebetter/internal/components"
	"practicebetter/internal/db"
	"practicebetter/internal/pages/librarypages"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gorilla/csrf"
	"github.com/mavolin/go-htmx"
	"github.com/nrednav/cuid2"
)

func (s *Server) libraryDashboard(w http.ResponseWriter, r *http.Request) {
	queries := db.New(s.DB)
	user := r.Context().Value(ck.UserKey).(db.User)
	pieces, err := queries.ListRecentlyPracticedPieces(r.Context(), user.ID)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Could not get pieces", http.StatusInternalServerError)
		return
	}

	hasPlan := false
	activePracticePlanID, ok := s.GetActivePracticePlanID(r.Context())
	var activePlan components.PracticePlanCardInfo
	if ok {
		p, err := queries.GetPracticePlanWithTodo(r.Context(), db.GetPracticePlanWithTodoParams{
			ID:     activePracticePlanID,
			UserID: user.ID,
		})

		if err != nil {
			log.Default().Println(err)
		} else {
			hasPlan = true
			activePlan.ID = p.ID
			activePlan.Date = p.Date
			activePlan.CompletedItems = p.CompletedSpotsCount + p.CompletedPiecesCount
			activePlan.TotalItems = p.PiecesCount + p.SpotsCount
			activePlan.PieceTitles = pieceTitlesForPlanCard(p.PieceTitles, p.SpotPieceTitles)
		}
	}

	recentPracticePlans, err := queries.ListRecentPracticePlans(r.Context(), db.ListRecentPracticePlansParams{
		ID:     activePracticePlanID,
		UserID: user.ID,
	})
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Could not get practice plans", http.StatusInternalServerError)
		return
	}
	recentPlanInfo := make([]components.PracticePlanCardInfo, 0, len(recentPracticePlans))
	for _, p := range recentPracticePlans {
		nextPlanInfo := components.PracticePlanCardInfo{
			ID:             p.ID,
			Date:           p.Date,
			CompletedItems: p.CompletedSpotsCount + p.CompletedPiecesCount,
			TotalItems:     p.PiecesCount + p.SpotsCount,
			PieceTitles:    pieceTitlesForPlanCard(p.PieceTitles, p.SpotPieceTitles),
		}
		recentPlanInfo = append(recentPlanInfo, nextPlanInfo)
	}

	s.HxRender(w, r, librarypages.Dashboard(s, pieces, hasPlan, activePlan, recentPlanInfo), "Library")
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
	Stage          string  `json:"stage"`
	Measures       *string `json:"measures,omitempty"`
	AudioPromptUrl string  `json:"audioPromptUrl,omitempty"`
	ImagePromptUrl string  `json:"imagePromptUrl,omitempty"`
	NotesPrompt    string  `json:"notesPrompt,omitempty"`
	TextPrompt     string  `json:"textPrompt,omitempty"`
	CurrentTempo   *int64  `json:"currentTempo,omitempty"`
	StageStarted   *int64  `json:"stageStarted,omitempty"`
}

const MAX_UPLOAD_SIZE = 1024 * 1024 // 1MiB

func (s *Server) uploadAudio(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
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
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) uploadAudioForm(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.UploadAudioForm(token), "Upload Audio")

}

func (s *Server) saveImage(file multipart.File, fileHeader *multipart.FileHeader, userID string) (string, string, error) {
	buff := make([]byte, 512)
	_, err := file.Read(buff)
	if err != nil {
		log.Default().Println(err)
		return "", "", fmt.Errorf("Failed to read file")
	}

	filetype := mimetype.Detect(buff)
	if !mimetype.EqualsAny(filetype.String(), "image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp") {
		log.Default().Println(filetype)
		return "", "", fmt.Errorf("the provided file format is not allowed. Please upload an image file.")
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		log.Default().Println(err)
		return "", "", fmt.Errorf("Failed to read file")
	}

	// Create the uploads folder if it doesn't
	// already exist
	h := sha256.New()
	h.Write([]byte(userID))
	userIDHash := hex.EncodeToString(h.Sum(nil))[:8]

	userImagePath := path.Join(s.UploadsPath, userIDHash, "images")
	err = os.MkdirAll(userImagePath, os.ModePerm)
	if err != nil {
		log.Default().Println(err)
		return "", "", fmt.Errorf("Failed to create file")
	}

	// Create a new file in the uploads directory
	newFileName := fmt.Sprintf("%s-%s", cuid2.Generate()[:5], fileHeader.Filename)
	newFilePath := path.Join(userImagePath, newFileName)

	dst, err := os.Create(newFilePath)
	if err != nil {
		log.Default().Println(err)
		return "", "", fmt.Errorf("Failed to create file")
	}

	defer dst.Close()

	// Copy the uploaded file to the filesystem
	// at the specified destination
	_, err = io.Copy(dst, file)
	if err != nil {
		log.Default().Println(err)
		return "", "", fmt.Errorf("Failed to copy file")
	}
	return newFileName, fmt.Sprintf("/uploads/%s/images/%s", userIDHash, newFileName), nil
}

func (s *Server) uploadImage(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
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

	newFileName, newFileUrl, err := s.saveImage(file, fileHeader, user.ID)
	if err != nil {
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  err.Error(),
			Title:    "Upload Error",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	data := map[string]string{
		"filename": newFileName,
		"url":      newFileUrl,
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) uploadImageForm(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)
	s.HxRender(w, r, librarypages.UploadImageForm(token), "Upload Image")

}

func pieceTitlesForPlanCard(pieceTitlesIn interface{}, spotPieceTitlesIn interface{}) []string {
	pieceTitles, ok := pieceTitlesIn.(string)
	if !ok {
		pieceTitles = ""
	}
	spotPieceTitles, ok := spotPieceTitlesIn.(string)
	if !ok {
		spotPieceTitles = ""
	}
	seenPieceTitles := make(map[string]struct{}, 0)
	uniquePieceTitlesList := make([]string, 0, len(pieceTitles))
	for _, pieceTitle := range strings.Split(strings.Trim(pieceTitles, "@"), "@,") {
		if _, ok := seenPieceTitles[pieceTitle]; ok || pieceTitle == "" {
			continue
		}
		uniquePieceTitlesList = append(uniquePieceTitlesList, pieceTitle)
		seenPieceTitles[pieceTitle] = struct{}{}
	}
	for _, pieceTitle := range strings.Split(strings.Trim(spotPieceTitles, "@"), "@,") {
		if _, ok := seenPieceTitles[pieceTitle]; ok || pieceTitle == "" {
			continue
		}
		uniquePieceTitlesList = append(uniquePieceTitlesList, pieceTitle)
		seenPieceTitles[pieceTitle] = struct{}{}
	}
	return uniquePieceTitlesList
}
