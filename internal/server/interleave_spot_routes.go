package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"practicebetter/internal/ck"
	"practicebetter/internal/components"
	"practicebetter/internal/db"
	"practicebetter/internal/pages/librarypages"
	"practicebetter/internal/pages/planpages"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/mavolin/go-htmx"
)

func (s *Server) getInterleaveList(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	planID := chi.URLParam(r, "planID")
	queries := db.New(s.DB)

	interleaveSpots, err := queries.GetPracticePlanInterleaveSpots(r.Context(), db.GetPracticePlanInterleaveSpotsParams{
		PlanID: planID,
		UserID: user.ID,
	})
	if err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not get interleave spots",
			Title:    "Not Found",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	allCompleted := true
	spotInfo := make([]planpages.PracticePlanSpot, 0, len(interleaveSpots))
	for _, interleaveSpot := range interleaveSpots {
		if !interleaveSpot.Completed {
			allCompleted = false
		}
		spotInfo = append(spotInfo, planpages.PracticePlanSpot{
			ID:         interleaveSpot.SpotID,
			Name:       interleaveSpot.SpotName.String,
			Measures:   interleaveSpot.SpotMeasures.String,
			Completed:  interleaveSpot.Completed,
			PieceID:    interleaveSpot.SpotPieceID.String,
			PieceTitle: interleaveSpot.SpotPieceTitle,
		})
	}
	rand.Shuffle(len(spotInfo), func(i, j int) {
		spotInfo[i], spotInfo[j] = spotInfo[j], spotInfo[i]
	})

	token := csrf.Token(r)

	hxRequest := htmx.Request(r)
	if hxRequest == nil {
		http.Redirect(w, r, fmt.Sprintf("/library/plans/%s", planID), http.StatusSeeOther)
		return
	}
	if err := planpages.PracticePlanInterleaveSpots(spotInfo, planID, token, allCompleted, false, false).Render(r.Context(), w); err != nil {
		log.Default().Println(err)
		http.Error(w, "Render Error", http.StatusInternalServerError)
	}
}

type PlanInterleaveSpotInfo struct {
	SpotID  string
	PieceID string
}

func (s *Server) startInterleavePracticing(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	planID := chi.URLParam(r, "planID")

	// Clear stale state
	if _, ok := s.SM.Get(r.Context(), "interleaveList").([]PlanInterleaveSpotInfo); ok {
		s.SM.Remove(r.Context(), "interleaveList")
	}
	if _, ok := s.SM.Get(r.Context(), "interleaveListIndex").(int); ok {
		s.SM.Remove(r.Context(), "interleaveListIndex")
	}

	queries := db.New(s.DB)

	interleaveSpots, err := queries.GetPracticePlanInterleaveSpots(r.Context(), db.GetPracticePlanInterleaveSpotsParams{
		PlanID: planID,
		UserID: user.ID,
	})

	if err != nil {
		if err := htmx.Trigger(r, "FinishedInterleave", components.INTERLEAVE_SPOT_DIALOG_ID); err != nil {
			log.Default().Println(err)
		}
		s.DatabaseError(w, r, err, "Could not get interleave spots")
		return
	}
	if len(interleaveSpots) == 0 {
		if err := htmx.Trigger(r, "FinishedInterleave", components.INTERLEAVE_SPOT_DIALOG_ID); err != nil {
			log.Default().Println(err)
		}
		s.DatabaseError(w, r, err, "No interleave spots")
		return
	}
	interleaveList := make([]PlanInterleaveSpotInfo, len(interleaveSpots))
	for i, interleaveSpot := range interleaveSpots {
		interleaveList[i] = PlanInterleaveSpotInfo{
			SpotID:  interleaveSpot.SpotID,
			PieceID: interleaveSpot.SpotPieceID.String,
		}
	}

	rand.Shuffle(len(interleaveList), func(i, j int) {
		interleaveList[i], interleaveList[j] = interleaveList[j], interleaveList[i]
	})
	s.SM.Put(r.Context(), "interleaveList", interleaveList)
	s.SM.Put(r.Context(), "interleaveListIndex", 0)

	if r.URL.Query().Get("goOn") == "true" {
		s.SM.Put(r.Context(), "interleaveGoOn", true)
	} else {

		s.SM.Put(r.Context(), "interleaveGoOn", false)
	}

	firstSpot, err := queries.GetSpot(r.Context(), db.GetSpotParams{
		SpotID:  interleaveList[0].SpotID,
		UserID:  user.ID,
		PieceID: interleaveList[0].PieceID,
	})
	if err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "FinishedInterleave", components.INTERLEAVE_SPOT_DIALOG_ID); err != nil {
			log.Default().Println(err)
		}
		s.DatabaseError(w, r, err, "Could not get first spot")
		return
	}

	displaySpot := DisplaySpot{
		ID:             firstSpot.ID,
		Name:           firstSpot.Name,
		Stage:          firstSpot.Stage,
		AudioPromptURL: firstSpot.AudioPromptUrl,
		ImagePromptURL: firstSpot.ImagePromptUrl,
		NotesPrompt:    firstSpot.NotesPrompt,
		TextPrompt:     firstSpot.TextPrompt,
	}

	if firstSpot.Measures.Valid {
		displaySpot.Measures = firstSpot.Measures.String
	}

	if firstSpot.CurrentTempo.Valid {
		displaySpot.CurrentTempo = &firstSpot.CurrentTempo.Int64
	}

	spotJSON, err := json.Marshal(displaySpot)
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

	token := csrf.Token(r)
	if err := librarypages.InterleavePracticeSpotDisplay(string(spotJSON), firstSpot.PieceID, firstSpot.PieceTitle, firstSpot.ID, planID, token).Render(r.Context(), w); err != nil {
		log.Default().Println(err)
		http.Error(w, "Render Error", http.StatusInternalServerError)
	}
}

func (s *Server) saveInterleaveResult(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	planID := chi.URLParam(r, "planID")

	activePlanID, ok := r.Context().Value(ck.ActivePlanKey).(string)
	if !ok || activePlanID != planID {
		log.Default().Println("Invalid plan ID")
		if err := htmx.Trigger(r, "FinishedInterleave", components.INTERLEAVE_SPOT_DIALOG_ID); err != nil {
			log.Default().Println(err)
		}
		if err := htmx.TriggerAfterSwap(r, "ShowAlert", ShowAlertEvent{
			Message:  "Cannot practice spot from inactive plan.",
			Title:    "Bad Request",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Invalid plan ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "FinishedInterleave", components.INTERLEAVE_SPOT_DIALOG_ID); err != nil {
			log.Default().Println(err)
		}
		if err := htmx.TriggerAfterSwap(r, "ShowAlert", ShowAlertEvent{
			Message:  "Invalid data submitted",
			Title:    "Bad Request",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	interleaveList, ok := s.SM.Get(r.Context(), "interleaveList").([]PlanInterleaveSpotInfo)
	if !ok {
		// TODO: close open dialog as well
		log.Default().Println("Missing interleave list")
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not get Interleave list",
			Title:    "Session Error",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Session Error", http.StatusInternalServerError)
		return

	}
	interleaveListIndex, ok := s.SM.Get(r.Context(), "interleaveListIndex").(int)
	if !ok {
		// TODO: close open dialog as well
		log.Default().Println("Missing interleave list")
		if err := htmx.Trigger(r, "FinishedInterleave", components.INTERLEAVE_SPOT_DIALOG_ID); err != nil {
			log.Default().Println(err)
		}
		if err := htmx.TriggerAfterSwap(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not get Interleave list Info",
			Title:    "Session Error",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Session Error", http.StatusInternalServerError)
		return
	}
	spotID := r.FormValue("spotID")
	pieceID := r.FormValue("pieceID")
	evaluation := r.FormValue("evaluation")

	if interleaveListIndex >= len(interleaveList) {
		// TODO: close open dialog as well
		log.Default().Println("Invalid index")
		s.SM.Remove(r.Context(), "interleaveList")
		s.SM.Remove(r.Context(), "interleaveListIndex")
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Invalid index",
			Title:    "Session Error",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Session Error", http.StatusInternalServerError)
		return
	}

	if interleaveList[interleaveListIndex].SpotID != spotID || interleaveList[interleaveListIndex].PieceID != pieceID {
		if err := htmx.Trigger(r, "FinishedInterleave", components.INTERLEAVE_SPOT_DIALOG_ID); err != nil {
			log.Default().Println(err)
		}
		log.Default().Println("Invalid spot or piece id")
		s.SM.Remove(r.Context(), "interleaveList")
		s.SM.Remove(r.Context(), "interleaveListIndex")
		if err := htmx.TriggerAfterSwap(r, "ShowAlert", ShowAlertEvent{
			Message:  "Invalid spot or piece id",
			Title:    "Bad Request",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	queries := db.New(s.DB)
	if evaluation == "excellent" || evaluation == "fine" || evaluation == "poor" {
		if err := queries.UpdateSpotEvaluation(r.Context(), db.UpdateSpotEvaluationParams{
			Evaluation: sql.NullString{String: evaluation, Valid: true},
			PlanID:     planID,
			UserID:     user.ID,
			SpotID:     spotID,
		}); err != nil {
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

		if err := queries.UpdatePiecePracticed(r.Context(), db.UpdatePiecePracticedParams{
			UserID:  user.ID,
			PieceID: pieceID,
		}); err != nil {
			s.DatabaseError(w, r, err, "Could not update piece practiced")
			return
		}
	} else {
		log.Default().Println("Missing interleave list")
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not get Interleave list Info",
			Title:    "Bad Request",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}
	interleaveListIndex++

	if interleaveListIndex >= len(interleaveList) {
		s.SM.Remove(r.Context(), "interleaveList")
		s.SM.Remove(r.Context(), "interleaveListIndex")
		goOn, ok := s.SM.Get(r.Context(), "interleaveGoOn").(bool)
		if ok && goOn {
			// TODO: double redirect gets weirddddd, reuse code from next function here
			// htmx.Redirect(r, "/library/plans/"+planID+"/next")
			// http.Redirect(w, r, "/library/plans/"+planID+"/next", http.StatusSeeOther)
			// htmx.Retarget(r, "#main-content")
			if err := htmx.TriggerAfterSwap(r, "ShowAlert", ShowAlertEvent{
				Message:  "Go again in a few minutes!",
				Title:    "Finished Interleave Spots",
				Variant:  "success",
				Duration: 3000,
			}); err != nil {
				log.Default().Println(err)
			}
			if err := htmx.Trigger(r, "FinishedInterleave", components.INTERLEAVE_SPOT_DIALOG_ID); err != nil {
				log.Default().Println(err)
			}
			return
		}
		if err := htmx.TriggerAfterSwap(r, "ShowAlert", ShowAlertEvent{
			Message:  "Go again in a few minutes!",
			Title:    "Finished Interleave Spots",
			Variant:  "success",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		if err := htmx.Trigger(r, "FinishedInterleave", components.INTERLEAVE_SPOT_DIALOG_ID); err != nil {
			log.Default().Println(err)
		}
		return
	}
	s.SM.Put(r.Context(), "interleaveListIndex", interleaveListIndex)

	nextSpot, err := queries.GetSpot(r.Context(), db.GetSpotParams{
		SpotID:  interleaveList[interleaveListIndex].SpotID,
		UserID:  user.ID,
		PieceID: interleaveList[interleaveListIndex].PieceID,
	})
	if err != nil {
		log.Default().Println(err)
		if err := htmx.TriggerAfterSwap(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not get next spot",
			Title:    "Database Error",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		if err := htmx.Trigger(r, "FinishedInterleave", components.INTERLEAVE_SPOT_DIALOG_ID); err != nil {
			log.Default().Println(err)
		}
		s.DatabaseError(w, r, err, "Could not get next spot")
		return
	}

	displaySpot := DisplaySpot{
		ID:             nextSpot.ID,
		Name:           nextSpot.Name,
		Stage:          nextSpot.Stage,
		AudioPromptURL: nextSpot.AudioPromptUrl,
		ImagePromptURL: nextSpot.ImagePromptUrl,
		NotesPrompt:    nextSpot.NotesPrompt,
		TextPrompt:     nextSpot.TextPrompt,
	}

	if nextSpot.Measures.Valid {
		displaySpot.Measures = nextSpot.Measures.String
	}

	if nextSpot.CurrentTempo.Valid {
		displaySpot.CurrentTempo = &nextSpot.CurrentTempo.Int64
	}
	spotJSON, err := json.Marshal(displaySpot)
	if err != nil {
		log.Default().Println(err)
		if err := htmx.TriggerAfterSwap(r, "ShowAlert", ShowAlertEvent{
			Message:  "Invalid Spot",
			Title:    "Error",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		if err := htmx.TriggerAfterSettle(r, "FinishedInterleave", components.INTERLEAVE_SPOT_DIALOG_ID); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token := csrf.Token(r)
	if err := librarypages.InterleavePracticeSpotDisplay(string(spotJSON), nextSpot.PieceID, nextSpot.PieceTitle, nextSpot.ID, planID, token).Render(r.Context(), w); err != nil {
		log.Default().Println(err)
		http.Error(w, "Render Error", http.StatusInternalServerError)
	}
}

// measures are missing and evaluation buttons are weird on narrow screen after swap
func (s *Server) startInfrequentPracticing(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	planID := chi.URLParam(r, "planID")

	queries := db.New(s.DB)

	firstSpot, err := queries.GetNextInfrequentSpot(r.Context(), db.GetNextInfrequentSpotParams{
		PlanID: planID,
		UserID: user.ID,
	})
	if err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "FinishedInterleave", components.INTERLEAVE_SPOT_DIALOG_ID); err != nil {
			log.Default().Println(err)
		}
		s.DatabaseError(w, r, err, "Could not get first spot")
		return
	}

	displaySpot := DisplaySpot{
		ID:             firstSpot.ID,
		Name:           firstSpot.Name,
		Stage:          firstSpot.Stage,
		AudioPromptURL: firstSpot.AudioPromptUrl,
		ImagePromptURL: firstSpot.ImagePromptUrl,
		NotesPrompt:    firstSpot.NotesPrompt,
		TextPrompt:     firstSpot.TextPrompt,
	}

	if firstSpot.Measures.Valid {
		displaySpot.Measures = firstSpot.Measures.String
	}

	if firstSpot.CurrentTempo.Valid {
		displaySpot.CurrentTempo = &firstSpot.CurrentTempo.Int64
	}

	spotJSON, err := json.Marshal(displaySpot)
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

	token := csrf.Token(r)
	if err := librarypages.InfrequentPracticeSpotDisplay(string(spotJSON), firstSpot.PieceID, firstSpot.PieceTitle, firstSpot.ID, planID, token).Render(r.Context(), w); err != nil {
		log.Default().Println(err)
		http.Error(w, "Render Error", http.StatusInternalServerError)
	}
}

type DisplaySpot struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Stage          string `json:"stage"`
	Measures       string `json:"measures"`
	AudioPromptURL string `json:"audioPromptUrl"`
	ImagePromptURL string `json:"imagePromptUrl"`
	NotesPrompt    string `json:"notesPropmt"`
	TextPrompt     string `json:"textPrompt"`
	CurrentTempo   *int64 `json:"currentTempo"`
	PieceID        string `json:"pieceID"`
}

func (s *Server) saveInfrequentResult(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	planID := chi.URLParam(r, "planID")

	activePlanID, ok := r.Context().Value(ck.ActivePlanKey).(string)
	if !ok || activePlanID != planID {
		log.Default().Println("Invalid:plan ID")
		if err := htmx.Trigger(r, "FinishedInterleave", components.INTERLEAVE_SPOT_DIALOG_ID); err != nil {
			log.Default().Println(err)
		}
		if err := htmx.TriggerAfterSwap(r, "ShowAlert", ShowAlertEvent{
			Message:  "Cannot practice spot from inactive plan.",
			Title:    "Bad Request",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Invalid plan ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "FinishedInfrequent", components.INTERLEAVE_SPOT_DIALOG_ID); err != nil {
			log.Default().Println(err)
		}
		if err := htmx.TriggerAfterSwap(r, "ShowAlert", ShowAlertEvent{
			Message:  "Invalid data submitted",
			Title:    "Bad Request",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	spotID := r.FormValue("spotID")
	pieceID := r.FormValue("pieceID")
	evaluation := r.FormValue("evaluation")

	queries := db.New(s.DB)

	if err := queries.CompletePracticePlanSpot(r.Context(), db.CompletePracticePlanSpotParams{
		SpotID: spotID,
		UserID: user.ID,
		PlanID: planID,
	}); err != nil {
		s.DatabaseError(w, r, err, "Could not complete practice plan spot")
		return
	}

	finishedSpot, err := queries.GetSpot(r.Context(), db.GetSpotParams{
		SpotID:  spotID,
		UserID:  user.ID,
		PieceID: pieceID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not get spot")
		return
	}

	skipDays := finishedSpot.SkipDays

	var timeSinceStarted time.Duration
	if finishedSpot.StageStarted.Valid {
		timeSinceStarted = time.Since(time.Unix(finishedSpot.StageStarted.Int64, 0))
	} else {
		timeSinceStarted = 0
		err := queries.FixSpotStageStarted(r.Context(), db.FixSpotStageStartedParams{
			SpotID: finishedSpot.ID,
			UserID: user.ID,
		})
		if err != nil {
			s.DatabaseError(w, r, err, "Could not fix spot stage started")
			return
		}
	}

	// excellent and more than three days old and days is less than 7, double the skip time
	if evaluation == "excellent" &&
		finishedSpot.StageStarted.Valid &&
		timeSinceStarted > 4*24*time.Hour &&
		finishedSpot.SkipDays < 7 {
		skipDays *= 2
		err := queries.UpdateSpotSkipDaysAndPractice(r.Context(), db.UpdateSpotSkipDaysAndPracticeParams{
			SkipDays: skipDays,
			SpotID:   finishedSpot.ID,
			UserID:   user.ID,
		})
		if err != nil {
			s.DatabaseError(w, r, err, "Could not update spot")
			return
		}
		// poor quality resets days to 1 or demotes immediately
	} else if evaluation == "poor" {
		if skipDays < 2 {
			err := queries.DemoteSpotToInterleave(r.Context(), db.DemoteSpotToInterleaveParams{
				SpotID: finishedSpot.ID,
				UserID: user.ID,
			})
			if err != nil {
				s.DatabaseError(w, r, err, "Could not demote spot")
				return
			}
		} else {
			skipDays = 1
			err := queries.UpdateSpotSkipDaysAndPractice(r.Context(), db.UpdateSpotSkipDaysAndPracticeParams{
				SkipDays: skipDays,
				SpotID:   finishedSpot.ID,
				UserID:   user.ID,
			})
			if err != nil {
				log.Default().Println(err)
				if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
					Message:  "Could not update spot",
					Title:    "Database Error",
					Variant:  "error",
					Duration: 3000,
				}); err != nil {
					log.Default().Println(err)
				}
				http.Error(w, "Database Error", http.StatusInternalServerError)
				return
			}
		}
		// completed promotes to completed, also have to verify conditions and not trust the client
	} else if evaluation == "excellent" && skipDays > 6 && timeSinceStarted > 20 {
		err := queries.PromoteSpotToCompleted(r.Context(), db.PromoteSpotToCompletedParams{
			SpotID: finishedSpot.ID,
			UserID: user.ID,
		})
		if err != nil {
			s.DatabaseError(w, r, err, "Could not promote spot")
			return
		}
	} else {
		if err := queries.UpdateSpotPracticed(r.Context(), db.UpdateSpotPracticedParams{
			SpotID: finishedSpot.ID,
			UserID: user.ID,
		}); err != nil {
			s.DatabaseError(w, r, err, "Could not update spot")
			return
		}
	}

	if err := queries.UpdatePiecePracticed(r.Context(), db.UpdatePiecePracticedParams{
		UserID:  user.ID,
		PieceID: pieceID,
	}); err != nil {
		s.DatabaseError(w, r, err, "Could not update piece practiced")
		return
	}

	fSpotInfo := librarypages.FinishedSpotInfo{
		SpotID:     finishedSpot.ID,
		PieceID:    finishedSpot.PieceID,
		PieceTitle: finishedSpot.PieceTitle,
		Name:       finishedSpot.Name,
	}
	if finishedSpot.Measures.Valid {
		fSpotInfo.Measures = finishedSpot.Measures.String
	}

	hasNext, err := queries.HasIncompleteInfrequentSpots(r.Context(), db.HasIncompleteInfrequentSpotsParams{
		PlanID: planID,
		UserID: user.ID,
	})
	if err != nil {
		if err := htmx.Trigger(r, "FinishedInfrequent", components.INTERLEAVE_SPOT_DIALOG_ID); err != nil {
			log.Default().Println(err)
		}
		s.DatabaseError(w, r, err, "Could not find spots")
		return
	}

	plan, err := queries.GetPracticePlanWithTodo(r.Context(), db.GetPracticePlanWithTodoParams{
		ID:     planID,
		UserID: user.ID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not find plan")
		return
	}
	if err := htmx.Trigger(r, "UpdatePlanProgress", UpdatePlanProgressEvent{
		Completed: int(plan.CompletedSpotsCount + plan.CompletedPiecesCount),
		Total:     int(plan.SpotsCount + plan.PiecesCount),
	}); err != nil {
		log.Default().Println(err)
	}

	if !hasNext {
		if err := htmx.Trigger(r, "FinishedInfrequent", components.INTERLEAVE_SPOT_DIALOG_ID); err != nil {
			log.Default().Println(err)
		}
		if err := htmx.TriggerAfterSwap(r, "ShowAlert", ShowAlertEvent{
			Message:  "You finished your infrequent spots for the day.",
			Title:    "Finished",
			Variant:  "success",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		if err := planpages.FinishInfrequentSpots(
			fSpotInfo.PieceID,
			fSpotInfo.SpotID,
			fSpotInfo.Name,
			fSpotInfo.Measures,
			fSpotInfo.PieceTitle,
		).Render(r.Context(), w); err != nil {
			log.Default().Println(err)
			http.Error(w, "Render Error", http.StatusInternalServerError)
		}
		return
	}

	nextSpot, err := queries.GetNextInfrequentSpot(r.Context(), db.GetNextInfrequentSpotParams{
		PlanID: planID,
		UserID: user.ID,
	})
	if err != nil {
		if err := htmx.Trigger(r, "FinishedInfrequent", components.INTERLEAVE_SPOT_DIALOG_ID); err != nil {
			log.Default().Println(err)
		}
		s.DatabaseError(w, r, err, "Could not get next spot")
		return
	}

	displaySpot := DisplaySpot{
		ID:             nextSpot.ID,
		Name:           nextSpot.Name,
		Stage:          nextSpot.Stage,
		AudioPromptURL: nextSpot.AudioPromptUrl,
		ImagePromptURL: nextSpot.ImagePromptUrl,
		NotesPrompt:    nextSpot.NotesPrompt,
		TextPrompt:     nextSpot.TextPrompt,
	}

	if nextSpot.Measures.Valid {
		displaySpot.Measures = nextSpot.Measures.String
	}

	if nextSpot.CurrentTempo.Valid {
		displaySpot.CurrentTempo = &nextSpot.CurrentTempo.Int64
	}

	spotJSON, err := json.Marshal(displaySpot)
	if err != nil {
		log.Default().Println(err)
		if err := htmx.TriggerAfterSwap(r, "ShowAlert", ShowAlertEvent{
			Message:  "Invalid Spot",
			Title:    "Error",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		if err := htmx.TriggerAfterSettle(r, "FinishedInfrequent", components.INTERLEAVE_SPOT_DIALOG_ID); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token := csrf.Token(r)
	if err := librarypages.InfrequentPracticeSpotDisplayWithOOBFinished(
		string(spotJSON),
		nextSpot.PieceID,
		nextSpot.PieceTitle,
		nextSpot.ID,
		planID,
		token,
		fSpotInfo,
	).Render(r.Context(), w); err != nil {
		log.Default().Println(err)
		http.Error(w, "Render Error", http.StatusInternalServerError)
	}
}
