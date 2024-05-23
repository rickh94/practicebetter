package server

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"practicebetter/internal/ck"
	"practicebetter/internal/config"
	"practicebetter/internal/db"
	"practicebetter/internal/pages/planpages"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/mavolin/go-htmx"
)

// Hack enum *sigh*
type PromoteDemote int

const (
	Promote PromoteDemote = 0
	Demote  PromoteDemote = 1
	Neither PromoteDemote = 2
)

func promoteDemoteInterleaveSpot(e string, started sql.NullInt64) PromoteDemote {
	if e == "excellent" &&
		started.Valid &&
		time.Since(time.Unix(started.Int64, 0)) > config.INTERLEAVE_SPOT_MIN_DAYS*24*time.Hour {
		return Promote
	}
	if e == "poor" || e == "fine" &&
		started.Valid &&
		time.Since(time.Unix(started.Int64, 0)) > config.INTERLEAVE_SPOT_MAX_DAYS*24*time.Hour {
		return Demote
	}
	return Neither
}

func (s *Server) completeInterleaveSpots(w http.ResponseWriter, r *http.Request, planID string, userID string) {
	tx, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		s.DatabaseError(w, r, err, "Could not start transaction")
		return
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Default().Println(err)
		}
	}()
	queries := db.New(s.DB)
	qtx := queries.WithTx(tx)

	spots, err := qtx.GetPracticePlanEvaluatedInterleaveSpots(r.Context(), db.GetPracticePlanEvaluatedInterleaveSpotsParams{
		PlanID: planID,
		UserID: userID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not get interleave spots")
		return
	}

	for _, sp := range spots {
		if !sp.Evaluation.Valid {
			continue
		}
		if err := qtx.CompletePracticePlanSpot(r.Context(), db.CompletePracticePlanSpotParams{
			PlanID: planID,
			UserID: userID,
			SpotID: sp.SpotID,
		}); err != nil {
			s.DatabaseError(w, r, err, "Could not complete spot")
			return
		}
		if !sp.SpotStageStarted.Valid {
			err := qtx.FixSpotStageStarted(r.Context(), db.FixSpotStageStartedParams{
				SpotID: sp.SpotID,
				UserID: userID,
			})
			if err != nil {
				s.DatabaseError(w, r, err, "Could not fix spot started time")
				return
			}
		}
		switch promoteDemoteInterleaveSpot(sp.Evaluation.String, sp.SpotStageStarted) {
		case Promote:
			err := qtx.PromoteSpotToInterleaveDays(r.Context(), db.PromoteSpotToInterleaveDaysParams{
				SpotID: sp.SpotID,
				UserID: userID,
			})
			if err != nil {
				s.DatabaseError(w, r, err, "Could not promote spot")
				return
			}
		case Demote:
			err := qtx.DemoteSpotToRandom(r.Context(), db.DemoteSpotToRandomParams{
				SpotID: sp.SpotID,
				UserID: userID,
			})
			if err != nil {
				s.DatabaseError(w, r, err, "Could not demote spot")
				return
			}
		case Neither:
			if err := qtx.UpdateSpotPracticed(r.Context(), db.UpdateSpotPracticedParams{
				SpotID: sp.SpotID,
				UserID: userID,
			}); err != nil {
				s.DatabaseError(w, r, err, "Could not update spot")
				return
			}
		}
	}

	if err := tx.Commit(); err != nil {
		s.DatabaseError(w, r, err, "Database error")
		return
	}
}

func (s *Server) deleteSpotFromPracticePlan(w http.ResponseWriter, r *http.Request) {
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
	spotID := chi.URLParam(r, "spotID")
	practiceType := chi.URLParam(r, "practiceType")
	queries := db.New(s.DB)
	err := queries.DeletePracticePlanSpot(r.Context(), db.DeletePracticePlanSpotParams{
		PlanID:       planID,
		UserID:       user.ID,
		SpotID:       spotID,
		PracticeType: practiceType,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not remove spot from practice plan")
		return
	}
	if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Your spot has been removed from this practice plan.",
		Title:    "Removed",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) getNewSpotPiecesForPracticePlan(w http.ResponseWriter, r *http.Request) {
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
	queries := db.New(s.DB)

	pieces, err := queries.ListPiecesWithNewSpotsForPlan(r.Context(), db.ListPiecesWithNewSpotsForPlanParams{
		PlanID: planID,
		UserID: user.ID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not retrieve new spot pieces")
		return
	}

	pieceInfo := make([]planpages.NewSpotPiece, 0, len(pieces))
	for _, row := range pieces {
		piece := planpages.NewSpotPiece{
			Title:         row.Title,
			ID:            row.ID,
			NewSpotsCount: row.NewSpotsCount,
		}
		if row.Composer.Valid {
			piece.Composer = row.Composer.String
		} else {
			piece.Composer = ""
		}
		pieceInfo = append(pieceInfo, piece)
	}
	if err := planpages.AddNewSpotPieceList(pieceInfo, planID).Render(r.Context(), w); err != nil {
		log.Default().Println(err)
		http.Error(w, "Render Error", http.StatusInternalServerError)
	}
}

func (s *Server) getNewSpotsForPracticePlan(w http.ResponseWriter, r *http.Request) {
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
	queries := db.New(s.DB)
	spots, err := queries.ListPieceSpotsInStageForPlan(r.Context(), db.ListPieceSpotsInStageForPlanParams{
		UserID:  user.ID,
		PieceID: pieceID,
		Stage:   "repeat",
		PlanID:  planID,
	})
	if err != nil || len(spots) == 0 {
		s.DatabaseError(w, r, err, "Could not retrieve new spot pieces")
		return
	}

	spotInfo := make([]planpages.PracticePlanSpot, 0, len(spots))
	for _, row := range spots {
		var spot planpages.PracticePlanSpot
		spot.ID = row.ID
		spot.Name = row.Name
		if row.Measures.Valid {
			spot.Measures = row.Measures.String
		} else {
			spot.Measures = ""
		}

		if row.StageStarted.Valid {
			spot.DaysSinceStarted = int64(time.Since(time.Unix(row.StageStarted.Int64, 0)).Hours() / 24)
		} else {
			spot.DaysSinceStarted = 0
			err := queries.FixSpotStageStarted(r.Context(), db.FixSpotStageStartedParams{
				SpotID: row.ID,
				UserID: user.ID,
			})
			if err != nil {
				log.Default().Println(err)
			}
		}

		spot.SkipDays = row.SkipDays

		spot.PieceTitle = row.PieceTitle
		spot.PieceID = row.PieceID
		spot.Completed = false
		spotInfo = append(spotInfo, spot)
	}
	token := csrf.Token(r)
	if err := planpages.AddNewSpotFormList(spotInfo, token, spots[0].PieceTitle, planID).Render(r.Context(), w); err != nil {
		log.Default().Println(err)
		http.Error(w, "Render Error", http.StatusInternalServerError)
	}
}

func (s *Server) getSpotsForPracticePlan(w http.ResponseWriter, r *http.Request) {
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
	var spotStage string
	switch practiceType {
	case "extra_repeat":
		spotStage = "extra_repeat"
	case "interleave_days":
		spotStage = "interleave_days"
	case "new":
		spotStage = "repeat"
	case "interleave":
		spotStage = "interleave"
	default:
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "That type of spot does not exist",
			Title:    "Invalid Spot Practice Type",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Invalid Spot practice type", http.StatusBadRequest)
		return
	}
	spots, err := queries.ListSpotsForPlanStage(r.Context(), db.ListSpotsForPlanStageParams{
		UserID: user.ID,
		Stage:  spotStage,
		PlanID: planID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not retrieve new spot pieces")
		return
	}

	spotInfo := make([]planpages.PracticePlanSpot, 0, len(spots))
	for _, row := range spots {
		var spot planpages.PracticePlanSpot
		spot.ID = row.ID
		spot.Name = row.Name
		if row.Measures.Valid {
			spot.Measures = row.Measures.String
		} else {
			spot.Measures = ""
		}

		if row.StageStarted.Valid {
			spot.DaysSinceStarted = int64(time.Since(time.Unix(row.StageStarted.Int64, 0)).Hours() / 24)
		} else {
			spot.DaysSinceStarted = 0
			err := queries.FixSpotStageStarted(r.Context(), db.FixSpotStageStartedParams{
				SpotID: row.ID,
				UserID: user.ID,
			})
			if err != nil {
				log.Default().Println(err)
			}
		}

		spot.SkipDays = row.SkipDays

		spot.PieceTitle = row.PieceTitle
		spot.PieceID = row.PieceID
		spot.Completed = false
		spotInfo = append(spotInfo, spot)
	}
	token := csrf.Token(r)
	if err := planpages.AddSpotFormList(spotInfo, planID, practiceType, token).Render(r.Context(), w); err != nil {
		log.Default().Println(err)
		http.Error(w, "Render Error", http.StatusInternalServerError)
	}
}

func (s *Server) addSpotsToPracticePlan(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	maxIdxRes, err := queries.GetMaxSpotIdx(r.Context(), db.GetMaxSpotIdxParams{
		PlanID: planID,
		UserID: user.ID,
	})
	maxIdx, ok := maxIdxRes.(int64)
	if err != nil || !ok {
		maxIdx = 0
	}
	for i, spotID := range r.Form["add-spots"] {
		_, err := queries.CreatePracticePlanSpotWithIdx(r.Context(), db.CreatePracticePlanSpotWithIdxParams{
			PracticePlanID: planID,
			SpotID:         spotID,
			PracticeType:   practiceType,
			Idx:            maxIdx + int64(i) + 1,
		})
		if err != nil {
			s.DatabaseError(w, r, err, "There was an error adding the spot")
			return
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

	spots, err := queries.ListPracticePlanSpotsInCategory(r.Context(), db.ListPracticePlanSpotsInCategoryParams{
		PracticeType: practiceType,
		PlanID:       planID,
		UserID:       user.ID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "There was an error retrieving your spots")
		return
	}
	var spotInfo []planpages.PracticePlanSpot
	for _, row := range spots {
		var spot planpages.PracticePlanSpot
		spot.ID = row.ID
		spot.Name = row.Name
		if row.Measures.Valid {
			spot.Measures = row.Measures.String
		} else {
			spot.Measures = ""
		}

		if row.StageStarted.Valid {
			spot.DaysSinceStarted = int64(time.Since(time.Unix(row.StageStarted.Int64, 0)).Hours() / 24)
		} else {
			spot.DaysSinceStarted = 0
			err := queries.FixSpotStageStarted(r.Context(), db.FixSpotStageStartedParams{
				SpotID: row.ID,
				UserID: user.ID,
			})
			if err != nil {
				log.Default().Println(err)
			}
		}

		spot.SkipDays = row.SkipDays

		spot.PieceTitle = row.PieceTitle
		spot.PieceID = row.PieceID
		spot.Completed = row.Completed
		spotInfo = append(spotInfo, spot)
	}

	switch practiceType {
	case "extra_repeat":
		if err := planpages.EditExtraRepeatSpotList(spotInfo, planID, csrf.Token(r)).Render(r.Context(), w); err != nil {
			log.Default().Println(err)
			http.Error(w, "Render Error", http.StatusInternalServerError)
		}
		return
	case "interleave_days":
		if err := planpages.EditInterleaveDaysSpotList(spotInfo, planID, csrf.Token(r)).Render(r.Context(), w); err != nil {
			log.Default().Println(err)
			http.Error(w, "Render Error", http.StatusInternalServerError)
		}
		return
	case "interleave":
		if err := planpages.EditInterleaveSpotList(spotInfo, planID, csrf.Token(r)).Render(r.Context(), w); err != nil {
			log.Default().Println(err)
			http.Error(w, "Render Error", http.StatusInternalServerError)
		}
		return
	case "new":
		if err := planpages.EditNewSpotList(spotInfo, planID, csrf.Token(r)).Render(r.Context(), w); err != nil {
			log.Default().Println(err)
			http.Error(w, "Render Error", http.StatusInternalServerError)
		}
		return
	default:
		http.Error(w, "Invalid practice type", http.StatusBadRequest)
	}
}
