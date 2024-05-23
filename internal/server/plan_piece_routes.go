package server

import (
	"log"
	"net/http"
	"practicebetter/internal/ck"
	"practicebetter/internal/db"
	"practicebetter/internal/pages/planpages"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/mavolin/go-htmx"
)

func (s *Server) deletePieceFromPracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	activePlanID, ok := s.GetActivePracticePlanID(r.Context())
	if !ok {
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot modify an inactive practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Cannot modify inactive plan", http.StatusBadRequest)
		return
	}
	planID := chi.URLParam(r, "planID")
	if activePlanID != planID {
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot modify an inactive practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Cannot modify inactive plan", http.StatusBadRequest)
		return
	}
	pieceID := chi.URLParam(r, "pieceID")
	practiceType := chi.URLParam(r, "practiceType")
	queries := db.New(s.DB)
	err := queries.DeletePracticePlanPiece(r.Context(), db.DeletePracticePlanPieceParams{
		PlanID:       planID,
		UserID:       user.ID,
		PieceID:      pieceID,
		PracticeType: practiceType,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not remove piece from practice plan")
		return
	}
	if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Your piece has been removed from this practice plan.",
		Title:    "Removed",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}
	w.WriteHeader(http.StatusOK)

}

func (s *Server) getPiecesForPracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	activePlanID, ok := s.GetActivePracticePlanID(r.Context())
	if !ok {
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot modify an inactive practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Cannot modify inactive plan", http.StatusBadRequest)
		return
	}
	planID := chi.URLParam(r, "planID")
	if activePlanID != planID {
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot modify an inactive practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Cannot modify inactive plan", http.StatusBadRequest)
		return
	}
	practiceType := chi.URLParam(r, "practiceType")
	queries := db.New(s.DB)
	var pieceInfo []planpages.PracticePlanPiece
	switch practiceType {
	case "random_spots":
		pieces, err := queries.ListRandomSpotPiecesForPlan(r.Context(), db.ListRandomSpotPiecesForPlanParams{
			UserID: user.ID,
			PlanID: planID,
		})
		if err != nil {
			s.DatabaseError(w, r, err, "There was an error retrieving your pieces")
			return
		}
		for _, row := range pieces {
			var piece planpages.PracticePlanPiece
			piece.ID = row.ID
			piece.Title = row.Title
			if row.Composer.Valid {
				piece.Composer = row.Composer.String
			}
			piece.Completed = false
			piece.RandomSpots = row.RandomSpotCount
			pieceInfo = append(pieceInfo, piece)
		}
	case "starting_point":
		pieces, err := queries.ListActivePiecesWithCompletedSpotsForPlan(r.Context(), db.ListActivePiecesWithCompletedSpotsForPlanParams{
			UserID: user.ID,
			PlanID: planID,
		})
		if err != nil {
			s.DatabaseError(w, r, err, "There was an error retrieving your pieces")
			return
		}
		for _, row := range pieces {
			var piece planpages.PracticePlanPiece
			piece.ID = row.ID
			piece.Title = row.Title
			if row.Composer.Valid {
				piece.Composer = row.Composer.String
			}
			piece.Completed = false
			pieceInfo = append(pieceInfo, piece)
		}
	default:
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "That type of piece does not exist",
			Title:    "Invalid Piece Practice Type",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Invalid Spot practice type", http.StatusBadRequest)
		return
	}
	token := csrf.Token(r)
	if err := planpages.AddPieceFormList(pieceInfo, token, planID, practiceType).Render(r.Context(), w); err != nil {
		log.Default().Println(err)
		http.Error(w, "Render Error", http.StatusInternalServerError)
	}
}

func (s *Server) addPiecesToPracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	activePlanID, ok := s.GetActivePracticePlanID(r.Context())
	if !ok {
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot modify an inactive practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Cannot modify inactive plan", http.StatusBadRequest)
		return
	}
	planID := chi.URLParam(r, "planID")
	if activePlanID != planID {
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot modify an inactive practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Cannot modify inactive plan", http.StatusBadRequest)
		return
	}
	practiceType := chi.URLParam(r, "practiceType")
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
		http.Error(w, "Invalid Form", http.StatusBadRequest)
		return
	}

	maxIdxRes, err := queries.GetMaxPieceIdx(r.Context(), db.GetMaxPieceIdxParams{
		PlanID: planID,
		UserID: user.ID,
	})
	maxIdx, ok := maxIdxRes.(int64)
	if err != nil || !ok {
		maxIdx = 0
	}

	for i, pieceID := range r.Form["add-pieces"] {
		_, err := queries.CreatePracticePlanPieceWithIdx(r.Context(), db.CreatePracticePlanPieceWithIdxParams{
			PracticePlanID: planID,
			PieceID:        pieceID,
			PracticeType:   practiceType,
			Idx:            maxIdx + int64(i) + 1,
		})
		if err != nil {
			s.DatabaseError(w, r, err, "There was an error creating the piece")
		}
	}

	if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Spots added to practice plan",
		Title:    "Success",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}

	pieces, err := queries.ListPracticePlanPiecesInCategory(r.Context(), db.ListPracticePlanPiecesInCategoryParams{
		PracticeType: practiceType,
		PlanID:       planID,
		UserID:       user.ID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "There was an error retrieving your spots")
		return
	}
	var pieceInfo []planpages.PracticePlanPiece
	for _, row := range pieces {
		var piece planpages.PracticePlanPiece
		piece.ID = row.PieceID
		piece.Title = row.PieceTitle
		if row.PieceComposer.Valid {
			piece.Composer = row.PieceComposer.String
		}
		piece.Completed = row.PieceCompleted
		piece.ActiveSpots = row.PieceActiveSpots
		piece.CompletedSpots = row.PieceCompletedSpots
		piece.RandomSpots = row.PieceRandomSpots

		pieceInfo = append(pieceInfo, piece)
	}

	switch practiceType {
	case "random_spots":
		if err := planpages.EditRandomPieceList(pieceInfo, true, planID, csrf.Token(r)).Render(r.Context(), w); err != nil {
			log.Default().Println(err)
			http.Error(w, "Render Error", http.StatusInternalServerError)
		}
		return
	case "starting_point":
		if err := planpages.EditStartingPointList(pieceInfo, true, planID, csrf.Token(r)).Render(r.Context(), w); err != nil {
			log.Default().Println(err)
			http.Error(w, "Render Error", http.StatusInternalServerError)
		}
		return
	default:
		http.Error(w, "Invalid practice type", http.StatusBadRequest)
	}
}
