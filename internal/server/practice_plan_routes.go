package server

import (
	"cmp"
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"practicebetter/internal/components"
	"practicebetter/internal/db"
	"practicebetter/internal/pages/planpages"
	"slices"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/mavolin/go-htmx"
	"github.com/nrednav/cuid2"
)

func (s *Server) createPracticePlanForm(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	token := csrf.Token(r)
	queries := db.New(s.DB)

	activePieces, err := queries.ListActiveUserPieces(r.Context(), user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.HxRender(w, r, planpages.CreatePracticePlanPage(s, token, activePieces), "Create Practice Plan")
}

// Consider making these configurable
const (
	LIGHT_MAX_NEW_SPOTS         = 0
	MEDIUM_MAX_NEW_SPOTS        = 5
	HEAVY_MAX_NEW_SPOTS         = 10
	LIGHT_MAX_INFREQUENT_SPOTS  = 7
	MEDIUM_MAX_INFREQUENT_SPOTS = 14
	HEAVY_MAX_INFREQUENT_SPOTS  = 20
	LIGHT_MAX_INTERLEAVE_SPOTS  = 8
	MEDIUM_MAX_INTERLEAVE_SPOTS = 12
	HEAVY_MAX_INTERLEAVE_SPOTS  = 20
)

type PotentialInfrequentSpot struct {
	ID        string
	TimeSince time.Duration
}

func (s *Server) createPracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	r.ParseForm()
	pieceIDs := r.Form["pieces"]
	tx, err := s.DB.Begin()
	if err != nil {
		log.Default().Printf("Database error: %v\n", err)
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	queries := db.New(s.DB)
	qtx := queries.WithTx(tx)

	practiceSessionID := cuid2.Generate()
	if err := qtx.CreatePracticeSession(r.Context(), db.CreatePracticeSessionParams{
		ID:              practiceSessionID,
		UserID:          user.ID,
		DurationMinutes: 0,
		Date:            time.Now().Unix(),
	}); err != nil {
		log.Default().Println(err)
		http.Error(w, "Could not create practice session", http.StatusInternalServerError)
		return
	}

	// We're going to carry forward failed new spots, so we need to get the new spots from the last plan that are
	// still in the repeat practice stage
	failedNewSpotIDs := make(map[string]struct{}, 0)
	failedNewSpots, err := qtx.GetPracticePlanFailedNewSpots(r.Context(), db.GetPracticePlanFailedNewSpotsParams{
		UserID:   user.ID,
		PieceIDs: pieceIDs,
	})
	if err != nil {
		log.Default().Println(err)
	} else {
		log.Default().Println("Found", len(failedNewSpots), "failed new spots")
		for _, spot := range failedNewSpots {
			failedNewSpotIDs[spot.SpotID] = struct{}{}
		}
	}

	newPlan, err := qtx.CreatePracticePlan(r.Context(), db.CreatePracticePlanParams{
		ID:                cuid2.Generate(),
		UserID:            user.ID,
		Intensity:         r.FormValue("intensity"),
		PracticeSessionID: sql.NullString{Valid: true, String: practiceSessionID},
	})
	if err != nil {
		log.Default().Printf("Database error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	maybeNewSpotIDs := make([]string, 0, len(pieceIDs)*10)
	maybeInfrequentSpots := make([]PotentialInfrequentSpot, 0, len(pieceIDs)*10)
	extraRepeatSpotIDs := make([]string, 0, len(pieceIDs)*10)
	interleaveSpotIDs := make([]string, 0, len(pieceIDs)*10)
	randomSpotPieceIDs := make([]string, 0, len(pieceIDs))
	startingPointPieceIDs := make([]string, 0, len(pieceIDs))

	for _, pieceID := range pieceIDs {
		pieceRows, err := qtx.GetPieceForPlan(r.Context(), db.GetPieceForPlanParams{
			PieceID: pieceID,
			UserID:  user.ID,
		})
		if err != nil {
			log.Default().Printf("Database error: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		pieceNewSpotIDs := make([]string, 0, len(pieceRows)/2)
		pieceRandomSpotCount := 0
		pieceExtraRepeatSpotCount := 0
		pieceCompletedSpotCount := 0

		for _, row := range pieceRows {
			if !row.SpotStage.Valid || !row.SpotID.Valid {
				continue
			}
			switch row.SpotStage.String {
			case "repeat":
				// we're going to combine the lists later, so make need to prevent duplicates
				if _, ok := failedNewSpotIDs[row.SpotID.String]; !ok {
					pieceNewSpotIDs = append(pieceNewSpotIDs, row.SpotID.String)
				}
			case "extra_repeat":
				extraRepeatSpotIDs = append(extraRepeatSpotIDs, row.SpotID.String)
				pieceExtraRepeatSpotCount++
			case "interleave":
				interleaveSpotIDs = append(interleaveSpotIDs, row.SpotID.String)
			case "interleave_days":
				if !row.SpotSkipDays.Valid {
					err := qtx.UpdateSpotSkipDays(r.Context(), db.UpdateSpotSkipDaysParams{
						SkipDays: 1,
						SpotID:   row.SpotID.String,
						UserID:   user.ID,
					})
					if err != nil {
						log.Default().Printf("Database error: %v\n", err)
						htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
							Message:  "Could not update invalid interleave days spot",
							Title:    "Failed to update",
							Variant:  "error",
							Duration: 3000,
						})
						return
					}
				}

				if !row.SpotLastPracticed.Valid || time.Since(time.Unix(row.SpotLastPracticed.Int64, 0)) > time.Duration(row.SpotSkipDays.Int64)*24*time.Hour {
					maybeInfrequentSpots = append(maybeInfrequentSpots, PotentialInfrequentSpot{
						ID:        row.SpotID.String,
						TimeSince: time.Since(time.Unix(row.SpotLastPracticed.Int64, 0)),
					})
				}
			case "random":
				pieceRandomSpotCount++
			case "completed":
				pieceCompletedSpotCount++
			default:
				continue
			}
		}

		// Only new spots if there aren't too many extra repeat/random spots.
		if pieceExtraRepeatSpotCount+pieceRandomSpotCount < 25 {
			maybeNewSpotIDs = append(maybeNewSpotIDs, pieceNewSpotIDs...)
		}

		if r.FormValue("practice_random_single") == "on" &&
			pieceRandomSpotCount > 2 {
			randomSpotPieceIDs = append(randomSpotPieceIDs, pieceID)
		}

		if r.FormValue("practice_starting_point") == "on" &&
			pieceCompletedSpotCount > 5 {
			startingPointPieceIDs = append(startingPointPieceIDs, pieceID)
		}
	}
	var maxNewSpots int
	var maxInfrequentSpots int
	var maxInterleaveSpots int
	switch r.FormValue("intensity") {
	case "light":
		maxNewSpots = LIGHT_MAX_NEW_SPOTS
		maxInfrequentSpots = LIGHT_MAX_INFREQUENT_SPOTS
		maxInterleaveSpots = LIGHT_MAX_INTERLEAVE_SPOTS
	case "medium":
		maxNewSpots = MEDIUM_MAX_NEW_SPOTS
		maxInfrequentSpots = MEDIUM_MAX_INFREQUENT_SPOTS
		maxInterleaveSpots = MEDIUM_MAX_INTERLEAVE_SPOTS
	case "heavy":
		maxNewSpots = HEAVY_MAX_NEW_SPOTS
		maxInfrequentSpots = HEAVY_MAX_INFREQUENT_SPOTS
		maxInterleaveSpots = HEAVY_MAX_INTERLEAVE_SPOTS
	}

	// Add Spots
	// extra repeat
	rand.Shuffle(len(extraRepeatSpotIDs), func(i, j int) {
		extraRepeatSpotIDs[i], extraRepeatSpotIDs[j] = extraRepeatSpotIDs[j], extraRepeatSpotIDs[i]
	})
	for i, spotID := range extraRepeatSpotIDs {
		_, err := qtx.CreatePracticePlanSpotWithIdx(r.Context(), db.CreatePracticePlanSpotWithIdxParams{
			PracticePlanID: newPlan.ID,
			SpotID:         spotID,
			PracticeType:   "extra_repeat",
			Idx:            int64(i),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
				Message:  "Could not add extra repeat spot",
				Title:    "Database Error",
				Variant:  "error",
				Duration: 3000,
			})
			return
		}
	}

	// interleave
	rand.Shuffle(len(interleaveSpotIDs), func(i, j int) {
		interleaveSpotIDs[i], interleaveSpotIDs[j] = interleaveSpotIDs[j], interleaveSpotIDs[i]
	})
	for i, spotID := range interleaveSpotIDs {
		if i >= maxInterleaveSpots {
			break
		}
		_, err := qtx.CreatePracticePlanSpotWithIdx(r.Context(), db.CreatePracticePlanSpotWithIdxParams{
			PracticePlanID: newPlan.ID,
			SpotID:         spotID,
			PracticeType:   "interleave",
			Idx:            int64(i),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
				Message:  "Could not add interleave spot",
				Title:    "Database Error",
				Variant:  "error",
				Duration: 3000,
			})
			return
		}
	}

	// prioritize infrequent spots with the spots that are the lest recently practiced first
	slices.SortFunc(maybeInfrequentSpots, func(a, b PotentialInfrequentSpot) int {
		return cmp.Compare(b.TimeSince, a.TimeSince)
	})

	// infrequent spots
	for i, spot := range maybeInfrequentSpots {
		if i >= maxInfrequentSpots {
			break
		}
		_, err := qtx.CreatePracticePlanSpotWithIdx(r.Context(), db.CreatePracticePlanSpotWithIdxParams{
			PracticePlanID: newPlan.ID,
			SpotID:         spot.ID,
			PracticeType:   "interleave_days",
			Idx:            int64(i),
		})
		if err != nil {
			htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
				Message:  "Failed to create Spot",
				Title:    "Database Error",
				Variant:  "error",
				Duration: 3000,
			})
			http.Error(w, "Database Error", http.StatusInternalServerError)
			return
		}
	}

	// new spots
	// Randomly select the new spots from new spots, but the spots from yesterday should go first and not be randomized
	rand.Shuffle(len(maybeNewSpotIDs), func(i, j int) {
		maybeNewSpotIDs[i], maybeNewSpotIDs[j] = maybeNewSpotIDs[j], maybeNewSpotIDs[i]
	})
	failedNewSpotIDList := make([]string, 0, len(failedNewSpots))
	for spotID := range failedNewSpotIDs {
		failedNewSpotIDList = append(failedNewSpotIDList, spotID)
	}
	maybeNewSpotIDs = append(failedNewSpotIDList, maybeNewSpotIDs...)
	for i, spotID := range maybeNewSpotIDs {
		if i >= maxNewSpots {
			break
		}
		_, err := qtx.CreatePracticePlanSpotWithIdx(r.Context(), db.CreatePracticePlanSpotWithIdxParams{
			PracticePlanID: newPlan.ID,
			SpotID:         spotID,
			PracticeType:   "new",
			Idx:            int64(i),
		})
		if err != nil {
			log.Default().Println(err)
			htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
				Message:  "Failed to add spot",
				Title:    "Database Error",
				Variant:  "error",
				Duration: 3000,
			})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// add pieces
	// random spots pieces
	rand.Shuffle(len(randomSpotPieceIDs), func(i, j int) {
		randomSpotPieceIDs[i], randomSpotPieceIDs[j] = randomSpotPieceIDs[j], randomSpotPieceIDs[i]
	})
	for i, pieceID := range randomSpotPieceIDs {
		_, err := qtx.CreatePracticePlanPieceWithIdx(r.Context(), db.CreatePracticePlanPieceWithIdxParams{
			PracticePlanID: newPlan.ID,
			PieceID:        pieceID,
			PracticeType:   "random_spots",
			Idx:            int64(i),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
				Message:  "Could not add random spot piece",
				Title:    "Database Error",
				Variant:  "error",
				Duration: 3000,
			})
			return
		}
	}
	// random starting point pieces
	rand.Shuffle(len(startingPointPieceIDs), func(i, j int) {
		startingPointPieceIDs[i], startingPointPieceIDs[j] = startingPointPieceIDs[j], startingPointPieceIDs[i]
	})
	for i, pieceID := range startingPointPieceIDs {
		_, err := qtx.CreatePracticePlanPieceWithIdx(r.Context(), db.CreatePracticePlanPieceWithIdxParams{
			PracticePlanID: newPlan.ID,
			PieceID:        pieceID,
			PracticeType:   "starting_point",
			Idx:            int64(i),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
				Message:  "Could not add random starting point piece",
				Title:    "Database Error",
				Variant:  "error",
				Duration: 3000,
			})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.SetActivePracticePlanID(r.Context(), newPlan.ID, user.ID)
	// update user (with newly added practice plan id) and practice plan manually before continuing
	user, err = queries.GetUserByID(r.Context(), user.ID)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.SM.RenewToken(r.Context()); err != nil {
		log.Default().Println(err)
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not renew session",
			Title:    "Failed to renew",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.FormValue("customize") == "on" {
		htmx.PushURL(r, "/library/plans/"+newPlan.ID+"/edit")
		ctx := context.WithValue(r.Context(), "activePracticePlanID", newPlan.ID)
		ctx = context.WithValue(ctx, "user", user)
		s.renderEditPracticePlanPage(w, r.WithContext(ctx), newPlan.ID, user.ID)
	} else {
		htmx.PushURL(r, "/library/plans/"+newPlan.ID)
		ctx := context.WithValue(r.Context(), "activePracticePlanID", newPlan.ID)
		ctx = context.WithValue(ctx, "user", user)
		s.renderPracticePlanPage(w, r.WithContext(ctx), newPlan.ID, user.ID)
	}
}

func (s *Server) singlePracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	planID := chi.URLParam(r, "planID")

	s.renderPracticePlanPage(w, r, planID, user.ID)
}

// TODO: find some way to decide whether to random spot or random starting point a piece depending on spot stages.

// TODO: make a distinction between practice the active practice plans and past practice plans

func (s *Server) renderPracticePlanPage(w http.ResponseWriter, r *http.Request, planID string, userID string) {
	queries := db.New(s.DB)
	totalItems := 0
	completedItems := 0
	planPieces, err := queries.GetPracticePlanWithPieces(r.Context(), db.GetPracticePlanWithPiecesParams{
		ID:     planID,
		UserID: userID,
	})
	if err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	planSpots, err := queries.GetPracticePlanWithSpots(r.Context(), db.GetPracticePlanWithSpotsParams{
		ID:     planID,
		UserID: userID,
	})

	if err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	activePracticePlanID, _ := s.GetActivePracticePlanID(r.Context())

	var planData planpages.PracticePlanData
	planData.ID = planID
	planData.IsActive = planID == activePracticePlanID
	if len(planPieces) == 0 {
		plan, err := queries.GetPracticePlan(r.Context(), db.GetPracticePlanParams{
			ID:     planID,
			UserID: userID,
		})
		if err != nil {
			log.Default().Println(err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		planData.Date = plan.Date
		planData.Completed = plan.Completed
		planData.InterleaveDaysSpotsCompleted = true
		planData.InterleaveSpotsCompleted = true
		planData.Intensity = plan.Intensity
	} else {
		planData.Date = planPieces[0].Date
		planData.Completed = planPieces[0].Completed
		planData.InterleaveDaysSpotsCompleted = true
		planData.InterleaveSpotsCompleted = true
		planData.Intensity = planPieces[0].Intensity
	}

	for _, row := range planPieces {
		if row.PieceID.Valid {
			totalItems++
			if row.PieceCompleted {
				completedItems++
			}
			var piece planpages.PracticePlanPiece
			piece.ID = row.PieceID.String
			piece.Title = row.PieceTitle.String
			piece.ActiveSpots = row.PieceActiveSpots
			piece.CompletedSpots = row.PieceCompletedSpots
			piece.RandomSpots = row.PieceRandomSpots
			if row.PieceComposer.Valid {
				piece.Composer = row.PieceComposer.String
			} else {
				piece.Composer = "Unknown"
			}
			piece.Completed = row.PieceCompleted

			if row.PiecePracticeType == "random_spots" {
				planData.RandomSpotsPieces = append(planData.RandomSpotsPieces, piece)
			} else if row.PiecePracticeType == "starting_point" {
				planData.RandomStartPieces = append(planData.RandomStartPieces, piece)
			}

		}
	}
	for _, row := range planSpots {
		if row.SpotID.Valid {
			totalItems++
			if row.SpotCompleted {
				completedItems++
			}
			var spot planpages.PracticePlanSpot
			spot.ID = row.SpotID.String
			spot.Name = row.SpotName.String
			if row.SpotMeasures.Valid {
				spot.Measures = row.SpotMeasures.String
			} else {
				spot.Measures = ""
			}

			if row.SpotStageStarted.Valid {
				spot.DaysSinceStarted = int64(time.Since(time.Unix(row.SpotStageStarted.Int64, 0)).Hours() / 24)
			} else {
				spot.DaysSinceStarted = 0
				err := queries.FixSpotStageStarted(r.Context(), db.FixSpotStageStartedParams{
					SpotID: row.SpotID.String,
					UserID: userID,
				})
				if err != nil {
					log.Default().Println(err)
				}
			}

			if row.SpotSkipDays.Valid {
				spot.SkipDays = row.SpotSkipDays.Int64
			} else {
				spot.SkipDays = 0
			}

			spot.PieceTitle = row.SpotPieceTitle
			spot.PieceID = row.SpotPieceID.String
			spot.Completed = row.SpotCompleted

			if row.SpotPracticeType == "interleave" {
				planData.InterleaveSpots = append(planData.InterleaveSpots, spot)
				if !row.SpotCompleted {
					planData.InterleaveSpotsCompleted = false
				}
			} else if row.SpotPracticeType == "interleave_days" {
				planData.InterleaveDaysSpots = append(planData.InterleaveDaysSpots, spot)
				if !row.SpotCompleted {
					planData.InterleaveDaysSpotsCompleted = false
				}
			} else if row.SpotPracticeType == "extra_repeat" {
				planData.ExtraRepeatSpots = append(planData.ExtraRepeatSpots, spot)
			} else if row.SpotPracticeType == "new" {
				planData.NewSpots = append(planData.NewSpots, spot)
			}
		}
	}
	planData.TotalItems = totalItems
	planData.CompletedItems = completedItems

	if planData.IsActive {
		rand.Shuffle(len(planData.InterleaveSpots), func(i, j int) {
			planData.InterleaveSpots[i], planData.InterleaveSpots[j] = planData.InterleaveSpots[j], planData.InterleaveSpots[i]
		})
	}

	token := csrf.Token(r)
	s.HxRender(w, r, planpages.PracticePlanPage(planData, token), "Practice Plan")
}

func (s *Server) editPracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	planID := chi.URLParam(r, "planID")

	s.renderEditPracticePlanPage(w, r, planID, user.ID)
}

func (s *Server) renderEditPracticePlanPage(w http.ResponseWriter, r *http.Request, planID string, userID string) {
	queries := db.New(s.DB)
	totalItems := 0
	completedItems := 0
	planPieces, err := queries.GetPracticePlanWithPieces(r.Context(), db.GetPracticePlanWithPiecesParams{
		ID:     planID,
		UserID: userID,
	})
	if err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	planSpots, err := queries.GetPracticePlanWithSpots(r.Context(), db.GetPracticePlanWithSpotsParams{
		ID:     planID,
		UserID: userID,
	})

	if err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	activePracticePlanID, _ := s.GetActivePracticePlanID(r.Context())

	var planData planpages.PracticePlanData
	planData.ID = planID
	planData.IsActive = planID == activePracticePlanID
	planData.IsActive = planID == activePracticePlanID
	if len(planPieces) == 0 {
		plan, err := queries.GetPracticePlan(r.Context(), db.GetPracticePlanParams{
			ID:     planID,
			UserID: userID,
		})
		if err != nil {
			log.Default().Println(err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		planData.Date = plan.Date
		planData.Completed = plan.Completed
		planData.InterleaveDaysSpotsCompleted = true
		planData.InterleaveSpotsCompleted = true
		planData.Intensity = plan.Intensity
	} else {
		planData.Date = planPieces[0].Date
		planData.Completed = planPieces[0].Completed
		planData.InterleaveDaysSpotsCompleted = true
		planData.InterleaveSpotsCompleted = true
		planData.Intensity = planPieces[0].Intensity
	}

	for _, row := range planPieces {
		if row.PieceID.Valid {
			totalItems++
			if row.PieceCompleted {
				completedItems++
			}
			var piece planpages.PracticePlanPiece
			piece.ID = row.PieceID.String
			piece.Title = row.PieceTitle.String
			piece.ActiveSpots = row.PieceActiveSpots
			piece.CompletedSpots = row.PieceCompletedSpots
			piece.RandomSpots = row.PieceRandomSpots
			if row.PieceComposer.Valid {
				piece.Composer = row.PieceComposer.String
			} else {
				piece.Composer = "Unknown"
			}
			piece.Completed = row.PieceCompleted

			if row.PiecePracticeType == "random_spots" {
				planData.RandomSpotsPieces = append(planData.RandomSpotsPieces, piece)
			} else if row.PiecePracticeType == "starting_point" {
				planData.RandomStartPieces = append(planData.RandomStartPieces, piece)
			}

		}
	}
	for _, row := range planSpots {
		if row.SpotID.Valid {
			totalItems++
			if row.SpotCompleted {
				completedItems++
			}
			var spot planpages.PracticePlanSpot
			spot.ID = row.SpotID.String
			spot.Name = row.SpotName.String
			if row.SpotMeasures.Valid {
				spot.Measures = row.SpotMeasures.String
			} else {
				spot.Measures = ""
			}

			if row.SpotStageStarted.Valid {
				spot.DaysSinceStarted = int64(time.Since(time.Unix(row.SpotStageStarted.Int64, 0)).Hours() / 24)
			} else {
				spot.DaysSinceStarted = 0
				err := queries.FixSpotStageStarted(r.Context(), db.FixSpotStageStartedParams{
					SpotID: row.SpotID.String,
					UserID: userID,
				})
				if err != nil {
					log.Default().Println(err)
				}
			}

			if row.SpotSkipDays.Valid {
				spot.SkipDays = row.SpotSkipDays.Int64
			} else {
				spot.SkipDays = 0
			}

			spot.PieceTitle = row.SpotPieceTitle
			spot.PieceID = row.SpotPieceID.String
			spot.Completed = row.SpotCompleted

			if row.SpotPracticeType == "interleave" {
				planData.InterleaveSpots = append(planData.InterleaveSpots, spot)
				if !row.SpotCompleted {
					planData.InterleaveSpotsCompleted = false
				}
			} else if row.SpotPracticeType == "interleave_days" {
				planData.InterleaveDaysSpots = append(planData.InterleaveDaysSpots, spot)
				if !row.SpotCompleted {
					planData.InterleaveDaysSpotsCompleted = false
				}
			} else if row.SpotPracticeType == "extra_repeat" {
				planData.ExtraRepeatSpots = append(planData.ExtraRepeatSpots, spot)
			} else if row.SpotPracticeType == "new" {
				planData.NewSpots = append(planData.NewSpots, spot)
			}
		}
	}
	planData.TotalItems = totalItems
	planData.CompletedItems = completedItems

	token := csrf.Token(r)
	s.HxRender(w, r, planpages.EditPracticePlanPage(planData, token), "Customize Practice Plan")
}

/*
after day 4 it can increase
excellent doubles until it reaches 7, on day 25 it can be completed
interleave days spot list should be randomized

*/

func (s *Server) completeInterleaveDaysPlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	planID := chi.URLParam(r, "planID")
	activePlanID, ok := s.GetActivePracticePlanID(r.Context())
	if !ok || planID != activePlanID {
		http.Error(w, "Cannot update inactive plan", http.StatusBadRequest)
		return
	}

	tx, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	queries := db.New(s.DB)
	qtx := queries.WithTx(tx)

	plan, err := qtx.GetPracticePlan(r.Context(), db.GetPracticePlanParams{
		ID:     planID,
		UserID: user.ID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var practiceSessionID string
	if plan.PracticeSessionID.Valid {
		practiceSessionID = plan.PracticeSessionID.String
		if err := qtx.ExtendPracticeSessionToNow(r.Context(), db.ExtendPracticeSessionToNowParams{
			ID:     practiceSessionID,
			UserID: user.ID,
		}); err != nil {
			log.Default().Println(err)
			http.Error(w, "Could not extend practice session", http.StatusInternalServerError)
			return
		}
	} else {
		practiceSessionID = cuid2.Generate()
		if err := qtx.CreatePracticeSession(r.Context(), db.CreatePracticeSessionParams{
			ID:              practiceSessionID,
			UserID:          user.ID,
			DurationMinutes: 5,
			Date:            time.Now().Unix() - 5*60,
		}); err != nil {
			log.Default().Println(err)
			http.Error(w, "Could not create practice session", http.StatusInternalServerError)
			return
		}
	}

	interleaveDaysSpots, err := qtx.GetPracticePlanInterleaveDaysSpots(r.Context(), db.GetPracticePlanInterleaveDaysSpotsParams{
		PlanID: planID,
		UserID: user.ID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	spotInfo := make([]planpages.PracticePlanSpot, 0, len(interleaveDaysSpots))
	for _, interleaveDaysSpot := range interleaveDaysSpots {
		if err := qtx.CompletePracticePlanSpot(r.Context(), db.CompletePracticePlanSpotParams{
			SpotID: interleaveDaysSpot.SpotID,
			UserID: user.ID,
			PlanID: planID,
		}); err != nil {
			log.Default().Println(err)
			http.Error(w, "Could not update spot", http.StatusInternalServerError)
			return
		}

		quality := r.FormValue(fmt.Sprintf("%s.quality", interleaveDaysSpot.SpotID))
		var skipDays int64
		if interleaveDaysSpot.SpotSkipDays.Valid {
			skipDays = interleaveDaysSpot.SpotSkipDays.Int64
		} else {
			skipDays = 1

			err := qtx.UpdateSpotSkipDays(r.Context(), db.UpdateSpotSkipDaysParams{
				SkipDays: 1,
				SpotID:   interleaveDaysSpot.SpotID,
				UserID:   user.ID,
			})

			if err != nil {
				log.Default().Println(err)
				htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
					Message:  "Could not fix interleave days spot",
					Title:    "Error",
					Variant:  "error",
					Duration: 3000,
				})
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		var timeSinceStarted time.Duration
		if interleaveDaysSpot.SpotStageStarted.Valid {
			timeSinceStarted = time.Since(time.Unix(interleaveDaysSpot.SpotStageStarted.Int64, 0))
		} else {
			timeSinceStarted = 0
			err := qtx.FixSpotStageStarted(r.Context(), db.FixSpotStageStartedParams{
				SpotID: interleaveDaysSpot.SpotID,
				UserID: user.ID,
			})
			if err != nil {
				log.Default().Println(err)
				htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
					Message:  "Could not fix interleave days spot",
					Title:    "Error",
					Variant:  "error",
					Duration: 3000,
				})
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		// excellent and more than three days old and days is less than 7, double the skip time
		if quality == "excellent" &&
			interleaveDaysSpot.SpotStageStarted.Valid &&
			timeSinceStarted > 4*24*time.Hour &&
			interleaveDaysSpot.SpotSkipDays.Int64 < 7 {
			skipDays *= 2
			err := qtx.UpdateSpotSkipDays(r.Context(), db.UpdateSpotSkipDaysParams{
				SkipDays: skipDays,
				SpotID:   interleaveDaysSpot.SpotID,
				UserID:   user.ID,
			})
			if err != nil {
				log.Default().Println(err)
				htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
					Message:  "Could not update spot",
					Title:    "Error",
					Variant:  "error",
					Duration: 3000,
				})
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// poor quality resets days to 1 or demotes immediately
		} else if quality == "poor" {
			if skipDays < 2 {
				err := qtx.DemoteSpotToInterleave(r.Context(), db.DemoteSpotToInterleaveParams{
					SpotID: interleaveDaysSpot.SpotID,
					UserID: user.ID,
				})
				if err != nil {
					log.Default().Println(err)
					htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
						Message:  "Could not demote spot",
						Title:    "Error",
						Variant:  "error",
						Duration: 3000,
					})
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			} else {
				skipDays = 1
				err := qtx.UpdateSpotSkipDays(r.Context(), db.UpdateSpotSkipDaysParams{
					SkipDays: skipDays,
					SpotID:   interleaveDaysSpot.SpotID,
					UserID:   user.ID,
				})
				if err != nil {
					log.Default().Println(err)
					htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
						Message:  "Could not update spot",
						Title:    "Error",
						Variant:  "error",
						Duration: 3000,
					})
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
			// completed promotes to completed, also have to verify conditions and not trust the client
		} else if quality == "completed" && skipDays > 6 && timeSinceStarted > 20 {
			err := qtx.PromoteSpotToCompleted(r.Context(), db.PromoteSpotToCompletedParams{
				SpotID: interleaveDaysSpot.SpotID,
				UserID: user.ID,
			})
			if err != nil {
				log.Default().Println(err)
				htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
					Message:  "Could not promote spot",
					Title:    "Error",
					Variant:  "error",
					Duration: 3000,
				})
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			if err := qtx.UpdateSpotPracticed(r.Context(), db.UpdateSpotPracticedParams{
				SpotID: interleaveDaysSpot.SpotID,
				UserID: user.ID,
			}); err != nil {
				log.Default().Println(err)
				htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
					Message:  "Could not update spot",
					Title:    "Error",
					Variant:  "error",
					Duration: 3000,
				})
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		if err := qtx.CreatePracticeSpot(r.Context(), db.CreatePracticeSpotParams{
			UserID:            user.ID,
			SpotID:            interleaveDaysSpot.SpotID,
			PracticeSessionID: practiceSessionID,
		}); err != nil {
			if err := qtx.AddRepToPracticeSpot(r.Context(), db.AddRepToPracticeSpotParams{
				UserID:            user.ID,
				SpotID:            interleaveDaysSpot.SpotID,
				PracticeSessionID: practiceSessionID,
			}); err != nil {
				log.Default().Println(err)
				http.Error(w, "Could not practice spot", http.StatusInternalServerError)
				return
			}
		}
		spotInfo = append(spotInfo, planpages.PracticePlanSpot{
			ID:               interleaveDaysSpot.SpotID,
			Name:             interleaveDaysSpot.SpotName.String,
			Measures:         interleaveDaysSpot.SpotMeasures.String,
			Completed:        true,
			PieceID:          interleaveDaysSpot.SpotPieceID.String,
			PieceTitle:       interleaveDaysSpot.SpotPieceTitle,
			SkipDays:         skipDays,
			DaysSinceStarted: int64(timeSinceStarted.Hours() / 24),
		})
	}

	if err := tx.Commit(); err != nil {
		log.Default().Println(err)
		http.Error(w, "Could not commit practice session", http.StatusInternalServerError)
		return
	}

	token := csrf.Token(r)
	htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "You completed your interleaved days spots!",
		Title:    "Completed!",
		Variant:  "success",
		Duration: 3000,
	})
	planpages.PracticePlanInterleaveDaysSpots(spotInfo, planID, token, true, true).Render(r.Context(), w)
}

func (s *Server) getNextPlanItem(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	planID := chi.URLParam(r, "planID")

	queries := db.New(s.DB)
	activePracticePlanID, _ := s.GetActivePracticePlanID(r.Context())

	if planID != activePracticePlanID {
		http.Redirect(w, r, "/library/plans/"+planID, http.StatusSeeOther)
		return
	}

	plan, err := queries.GetPracticePlan(r.Context(), db.GetPracticePlanParams{
		ID:     planID,
		UserID: user.ID,
	})
	if err != nil {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not get next plan item",
			Title:    "Database Error",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	extraRepeatSpots, err := queries.GetPracticePlanIncompleteExtraRepeatSpots(r.Context(), db.GetPracticePlanIncompleteExtraRepeatSpotsParams{
		PlanID: planID,
		UserID: user.ID,
	})
	if err != nil {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not get next plan item",
			Title:    "Database Error",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(extraRepeatSpots) > 0 {
		spotID := extraRepeatSpots[0].SpotID
		pieceID := extraRepeatSpots[0].SpotPieceID
		htmx.Redirect(r, "/library/pieces/"+pieceID+"/spots/"+spotID+"/practice/repeat")
		http.Redirect(w, r, "/library/pieces/"+pieceID+"/spots/"+spotID+"/practice/repeat", http.StatusSeeOther)
		return
	}
	randomPieces, err := queries.GetPracticePlanIncompleteRandomPieces(r.Context(), db.GetPracticePlanIncompleteRandomPiecesParams{
		PlanID: planID,
		UserID: user.ID,
	})
	if err != nil {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not get next plan item",
			Title:    "Database Error",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(randomPieces) > 0 {
		url := components.GetPiecePracticeUrl(false, true, randomPieces[0].PieceID, "random_spots", plan.Intensity)
		htmx.Redirect(r, url)
		http.Redirect(w, r, url, http.StatusSeeOther)
		return
	}
	startingPointPieces, err := queries.GetPracticePlanIncompleteStartingPointPieces(r.Context(), db.GetPracticePlanIncompleteStartingPointPiecesParams{
		PlanID: planID,
		UserID: user.ID,
	})
	if err != nil {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not get next plan item",
			Title:    "Database Error",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(startingPointPieces) > 0 {
		url := components.GetPiecePracticeUrl(false, true, startingPointPieces[0].PieceID, "starting_point", plan.Intensity)
		htmx.Redirect(r, url)
		http.Redirect(w, r, url, http.StatusSeeOther)
		return
	}

	newSpots, err := queries.GetPracticePlanIncompleteNewSpots(r.Context(), db.GetPracticePlanIncompleteNewSpotsParams{
		PlanID: planID,
		UserID: user.ID,
	})
	if err != nil {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not get next plan item",
			Title:    "Database Error",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(newSpots) > 0 {
		spotID := newSpots[0].SpotID
		pieceID := newSpots[0].SpotPieceID
		htmx.Redirect(r, "/library/pieces/"+pieceID+"/spots/"+spotID+"/practice/repeat")
		http.Redirect(w, r, "/library/pieces/"+pieceID+"/spots/"+spotID+"/practice/repeat", http.StatusSeeOther)
		return
	}

	htmx.TriggerAfterSettle(r, "ShowAlert", ShowAlertEvent{
		Message:  "No more items to practice. Check you infrequent and interleave spots one last time and you are done.",
		Title:    "Almost Done",
		Variant:  "success",
		Duration: 3000,
	})

	htmx.Redirect(r, "/library/plans/"+planID)
	http.Redirect(w, r, "/library/plans/"+planID, http.StatusSeeOther)
}

/*
excellent after 7 days = promote
poor ever = demote
fine = stay
fine after 10 days = demote
*/

func (s *Server) completeInterleavePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	planID := chi.URLParam(r, "planID")
	activePlanID, ok := s.GetActivePracticePlanID(r.Context())
	if !ok || planID != activePlanID {
		http.Error(w, "Cannot update inactive plan", http.StatusBadRequest)
		return
	}

	tx, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	queries := db.New(s.DB)
	qtx := queries.WithTx(tx)

	plan, err := qtx.GetPracticePlan(r.Context(), db.GetPracticePlanParams{
		ID:     planID,
		UserID: user.ID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var practiceSessionID string
	if plan.PracticeSessionID.Valid {
		practiceSessionID = plan.PracticeSessionID.String
		if err := qtx.ExtendPracticeSessionToNow(r.Context(), db.ExtendPracticeSessionToNowParams{
			ID:     practiceSessionID,
			UserID: user.ID,
		}); err != nil {
			log.Default().Println(err)
			http.Error(w, "Could not extend practice session", http.StatusInternalServerError)
			return
		}
	} else {
		practiceSessionID = cuid2.Generate()
		if err := qtx.CreatePracticeSession(r.Context(), db.CreatePracticeSessionParams{
			ID:              practiceSessionID,
			UserID:          user.ID,
			DurationMinutes: 5,
			Date:            time.Now().Unix() - 5*60,
		}); err != nil {
			log.Default().Println(err)
			http.Error(w, "Could not create practice session", http.StatusInternalServerError)
			return
		}
	}

	interleaveSpots, err := qtx.GetPracticePlanInterleaveSpots(r.Context(), db.GetPracticePlanInterleaveSpotsParams{
		PlanID: planID,
		UserID: user.ID,
	})
	if err != nil {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not get interleave spots",
			Title:    "Error",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	spotInfo := make([]planpages.PracticePlanSpot, 0, len(interleaveSpots))
	for _, interleaveSpot := range interleaveSpots {
		if err := qtx.CompletePracticePlanSpot(r.Context(), db.CompletePracticePlanSpotParams{
			SpotID: interleaveSpot.SpotID,
			UserID: user.ID,
			PlanID: planID,
		}); err != nil {
			log.Default().Println(err)
			htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
				Message:  "Could not complete spot",
				Title:    "Error",
				Variant:  "error",
				Duration: 3000,
			})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !interleaveSpot.SpotStageStarted.Valid {
			err := qtx.FixSpotStageStarted(r.Context(), db.FixSpotStageStartedParams{
				SpotID: interleaveSpot.SpotID,
				UserID: user.ID,
			})
			if err != nil {
				log.Default().Println(err)
				htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
					Message:  "Could not fix spot started time",
					Title:    "Error",
					Variant:  "error",
					Duration: 3000,
				})
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		quality := r.FormValue(fmt.Sprintf("%s.quality", interleaveSpot.SpotID))
		if quality == "excellent" && interleaveSpot.SpotStageStarted.Valid && time.Since(time.Unix(interleaveSpot.SpotStageStarted.Int64, 0)) > 7*24*time.Hour {
			err := qtx.PromoteSpotToInterleaveDays(r.Context(), db.PromoteSpotToInterleaveDaysParams{
				SpotID: interleaveSpot.SpotID,
				UserID: user.ID,
			})
			if err != nil {
				log.Default().Println(err)
				htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
					Message:  "Could not promote spot",
					Title:    "Error",
					Variant:  "error",
					Duration: 3000,
				})
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else if quality == "poor" || quality == "fine" && interleaveSpot.SpotStageStarted.Valid && time.Since(time.Unix(interleaveSpot.SpotStageStarted.Int64, 0)) > 10*24*time.Hour {
			err := qtx.DemoteSpotToRandom(r.Context(), db.DemoteSpotToRandomParams{
				SpotID: interleaveSpot.SpotID,
				UserID: user.ID,
			})
			if err != nil {
				log.Default().Println(err)
				htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
					Message:  "Could not demote spot",
					Title:    "Error",
					Variant:  "error",
					Duration: 3000,
				})
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			if err := qtx.UpdateSpotPracticed(r.Context(), db.UpdateSpotPracticedParams{
				SpotID: interleaveSpot.SpotID,
				UserID: user.ID,
			}); err != nil {
				log.Default().Println(err)
				htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
					Message:  "Could not update spot",
					Title:    "Error",
					Variant:  "error",
					Duration: 3000,
				})
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		if err := qtx.CreatePracticeSpot(r.Context(), db.CreatePracticeSpotParams{
			UserID:            user.ID,
			SpotID:            interleaveSpot.SpotID,
			PracticeSessionID: practiceSessionID,
		}); err != nil {
			if err := qtx.AddRepToPracticeSpot(r.Context(), db.AddRepToPracticeSpotParams{
				UserID:            user.ID,
				SpotID:            interleaveSpot.SpotID,
				PracticeSessionID: practiceSessionID,
			}); err != nil {
				log.Default().Println(err)
				htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
					Message:  "Could not add rep to spot",
					Title:    "Error",
					Variant:  "error",
					Duration: 3000,
				})
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		spotInfo = append(spotInfo, planpages.PracticePlanSpot{
			ID:         interleaveSpot.SpotID,
			Name:       interleaveSpot.SpotName.String,
			Measures:   interleaveSpot.SpotMeasures.String,
			Completed:  true,
			PieceID:    interleaveSpot.SpotPieceID.String,
			PieceTitle: interleaveSpot.SpotPieceTitle,
		})
	}

	if err := tx.Commit(); err != nil {
		log.Default().Println(err)
		http.Error(w, "Could not commit practice session", http.StatusInternalServerError)
		return
	}

	token := csrf.Token(r)
	htmx.TriggerAfterSettle(r, "ShowAlert", ShowAlertEvent{
		Message:  "You completed your interleaved spots!",
		Title:    "Completed!",
		Variant:  "success",
		Duration: 3000,
	})
	planpages.PracticePlanInterleaveSpots(spotInfo, planID, token, true, true, true).Render(r.Context(), w)

}

func (s *Server) getInterleaveList(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	planID := chi.URLParam(r, "planID")
	queries := db.New(s.DB)

	interleaveSpots, err := queries.GetPracticePlanInterleaveSpots(r.Context(), db.GetPracticePlanInterleaveSpotsParams{
		PlanID: planID,
		UserID: user.ID,
	})
	if err != nil {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not get interleave spots",
			Title:    "Error",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, err.Error(), http.StatusNotFound)
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
	planpages.PracticePlanInterleaveSpots(spotInfo, planID, token, allCompleted, false, false).Render(r.Context(), w)
}

const plansPerPage = 20

func (s *Server) planList(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}
	pageNum, err := strconv.Atoi(page)
	if err != nil {
		pageNum = 1
	}
	s.renderPlanListPage(w, r, user.ID, pageNum)
}

func (s *Server) renderPlanListPage(w http.ResponseWriter, r *http.Request, userID string, pageNum int) {
	queries := db.New(s.DB)
	plans, err := queries.ListPaginatedPracticePlans(r.Context(), db.ListPaginatedPracticePlansParams{
		UserID: userID,
		Limit:  piecesPerPage,
		Offset: int64((pageNum - 1) * piecesPerPage),
	})
	if err != nil {
		log.Default().Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	totalPlans, err := queries.CountUserPracticePlans(r.Context(), userID)
	if err != nil {
		log.Default().Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	totalPages := int(math.Ceil(float64(totalPlans) / float64(plansPerPage)))
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	planInfo := make([]components.PracticePlanCardInfo, 0, len(plans))
	for _, p := range plans {
		nextPlanInfo := components.PracticePlanCardInfo{
			ID:             p.ID,
			Date:           p.Date,
			CompletedItems: p.CompletedSpotsCount + p.CompletedPiecesCount,
			TotalItems:     p.PiecesCount + p.SpotsCount,
			PieceTitles:    pieceTitlesForPlanCard(p.PieceTitles, p.SpotPieceTitles),
		}

		planInfo = append(planInfo, nextPlanInfo)
	}
	w.WriteHeader(http.StatusOK)
	s.HxRender(w, r, planpages.PlanList(planInfo, pageNum, totalPages), "Your Practice Plans")
}

func (s *Server) deletePracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	planID := chi.URLParam(r, "planID")
	queries := db.New(s.DB)
	if err := queries.DeletePracticePlan(r.Context(), db.DeletePracticePlanParams{ID: planID, UserID: user.ID}); err != nil {
		log.Default().Println(err)
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not delete practice plan",
			Title:    "Error",
			Variant:  "danger",
			Duration: 3000,
		})
		http.Error(w, "Could not delete practice plan", http.StatusInternalServerError)
		return
	}
	htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Practice plan deleted",
		Title:    "Deleted!",
		Variant:  "success",
		Duration: 3000,
	})
	htmx.PushURL(r, "/library/plans")

	// refresh user from database in case the active plan was deleted
	user, err := queries.GetUserByID(r.Context(), user.ID)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	ctx := context.WithValue(r.Context(), "user", user)
	if !user.ActivePracticePlanID.Valid {
		ctx = context.WithValue(ctx, "activePracticePlanID", "")
	}
	w.WriteHeader(http.StatusOK)
	s.renderPlanListPage(w, r.WithContext(ctx), user.ID, 1)
}

func (s *Server) resumePracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	planID := chi.URLParam(r, "planID")
	queries := db.New(s.DB)
	plan, err := queries.GetPracticePlan(r.Context(), db.GetPracticePlanParams{
		ID:     planID,
		UserID: user.ID,
	})
	if err != nil {
		http.Error(w, "Could not find matching practice plan", http.StatusNotFound)
		return
	}
	activePlanID, ok := s.GetActivePracticePlanID(r.Context())
	if ok && activePlanID == plan.ID {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You are already using this practice plan.",
			Title:    "Already Active",
			Variant:  "warning",
			Duration: 3000,
		})
		w.WriteHeader(http.StatusBadRequest)
		return

	}
	if time.Since(time.Unix(plan.Date, 0)) > 5*time.Hour {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot resume a practice plan this old. Please create a new one instead.",
			Title:    "Too Old",
			Variant:  "warning",
			Duration: 3000,
		})
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = s.SetActivePracticePlanID(r.Context(), planID, user.ID)
	if err != nil {
		log.Default().Println(err)
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not activate this practice plan.",
			Title:    "Error",
			Variant:  "danger",
			Duration: 3000,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// update user (with newly added practice plan id) and practice plan manually before continuing
	user, err = queries.GetUserByID(r.Context(), user.ID)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "You resumed your practice plan!",
		Title:    "Resumed!",
		Variant:  "success",
		Duration: 3000,
	})
	ctx := context.WithValue(r.Context(), "activePracticePlanID", plan.ID)
	ctx = context.WithValue(ctx, "user", user)
	s.renderPracticePlanPage(w, r.WithContext(ctx), plan.ID, user.ID)
}

func (s *Server) stopPracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	planID := chi.URLParam(r, "planID")
	queries := db.New(s.DB)
	plan, err := queries.GetPracticePlan(r.Context(), db.GetPracticePlanParams{
		ID:     planID,
		UserID: user.ID,
	})
	if err != nil {
		http.Error(w, "Could not find matching practice plan", http.StatusNotFound)
		return
	}
	activePlanID, ok := s.GetActivePracticePlanID(r.Context())
	if !ok || activePlanID != plan.ID {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "This plan is not active",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		})
		w.WriteHeader(http.StatusBadRequest)
		return

	}
	err = queries.ClearActivePracticePlan(r.Context(), user.ID)
	if err != nil {
		log.Default().Println(err)
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not stop this practice plan.",
			Title:    "Error",
			Variant:  "danger",
			Duration: 3000,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// update user (with removed practice plan id) and practice plan manually before continuing
	user, err = queries.GetUserByID(r.Context(), user.ID)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "You stopped your practice plan!",
		Title:    "Stopped",
		Variant:  "success",
		Duration: 3000,
	})
	ctx := context.WithValue(r.Context(), "activePracticePlanID", "")
	ctx = context.WithValue(ctx, "user", user)
	s.renderPracticePlanPage(w, r.WithContext(ctx), plan.ID, user.ID)
}

func (s *Server) deleteSpotFromPracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	activePlanID, ok := s.GetActivePracticePlanID(r.Context())
	if !ok {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot modify an inactive practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Cannot modify inactive plan", http.StatusBadRequest)
		return
	}
	planID := chi.URLParam(r, "planID")
	if activePlanID != planID {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot modify an inactive practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		})
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
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "There was an error removing your spot",
			Title:    "Database Error",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Database Error", http.StatusInternalServerError)
		return
	}
	htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Your spot has been removed from this practice plan.",
		Title:    "Removed",
		Variant:  "success",
		Duration: 3000,
	})
	w.WriteHeader(http.StatusOK)
}

func (s *Server) deletePieceFromPracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	activePlanID, ok := s.GetActivePracticePlanID(r.Context())
	if !ok {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot modify an inactive practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Cannot modify inactive plan", http.StatusBadRequest)
		return
	}
	planID := chi.URLParam(r, "planID")
	if activePlanID != planID {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot modify an inactive practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		})
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
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "There was an error removing your piece",
			Title:    "Database Error",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Database Error", http.StatusInternalServerError)
		return
	}
	htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Your piece has been removed from this practice plan.",
		Title:    "Removed",
		Variant:  "success",
		Duration: 3000,
	})
	w.WriteHeader(http.StatusOK)

}

func (s *Server) getNewSpotPiecesForPracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	activePlanID, ok := s.GetActivePracticePlanID(r.Context())
	if !ok {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot modify an inactive practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Cannot modify inactive plan", http.StatusBadRequest)
		return
	}
	planID := chi.URLParam(r, "planID")
	if activePlanID != planID {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot modify an inactive practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Cannot modify inactive plan", http.StatusBadRequest)
		return
	}
	queries := db.New(s.DB)

	pieces, err := queries.ListPiecesWithNewSpotsForPlan(r.Context(), db.ListPiecesWithNewSpotsForPlanParams{
		PlanID: planID,
		UserID: user.ID,
	})
	if err != nil {
		log.Default().Println(err)
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "There was an error retrieving your spots",
			Title:    "Database Error",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Database Error", http.StatusInternalServerError)
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
	planpages.AddNewSpotPieceList(pieceInfo, planID).Render(r.Context(), w)
}

func (s *Server) getNewSpotsForPracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	activePlanID, ok := s.GetActivePracticePlanID(r.Context())
	if !ok {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot modify an inactive practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Cannot modify inactive plan", http.StatusBadRequest)
		return
	}
	planID := chi.URLParam(r, "planID")
	if activePlanID != planID {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot modify an inactive practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		})
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
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "There was an error retrieving your spots",
			Title:    "Database Error",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Database Error", http.StatusInternalServerError)
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
	planpages.AddNewSpotFormList(spotInfo, token, spots[0].PieceTitle, planID).Render(r.Context(), w)
}

func (s *Server) getSpotsForPracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	activePlanID, ok := s.GetActivePracticePlanID(r.Context())
	if !ok {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot modify an inactive practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Cannot modify inactive plan", http.StatusBadRequest)
		return
	}
	planID := chi.URLParam(r, "planID")
	if activePlanID != planID {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot modify an inactive practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		})
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
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "That type of spot does not exist",
			Title:    "Invalid Spot Practice Type",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Invalid Spot practice type", http.StatusBadRequest)
		return
	}
	spots, err := queries.ListSpotsForPlanStage(r.Context(), db.ListSpotsForPlanStageParams{
		UserID: user.ID,
		Stage:  spotStage,
		PlanID: planID,
	})
	if err != nil {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "There was an error retrieving your spots",
			Title:    "Database Error",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Database Error", http.StatusInternalServerError)
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
	planpages.AddSpotFormList(spotInfo, planID, practiceType, token).Render(r.Context(), w)
}

func (s *Server) addSpotsToPracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	activePlanID, ok := s.GetActivePracticePlanID(r.Context())
	if !ok {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot modify an inactive practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Cannot modify inactive plan", http.StatusBadRequest)
		return
	}
	planID := chi.URLParam(r, "planID")
	if activePlanID != planID {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot modify an inactive practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Cannot modify inactive plan", http.StatusBadRequest)
		return
	}
	practiceType := chi.URLParam(r, "practiceType")
	queries := db.New(s.DB)

	r.ParseForm()

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
			htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
				Message:  "There was an error adding the spot",
				Title:    "Database Error",
				Variant:  "error",
				Duration: 3000,
			})
			http.Error(w, "Database Error", http.StatusInternalServerError)
			return
		}
	}

	htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Spots added to practice plan",
		Title:    "Success",
		Variant:  "success",
		Duration: 3000,
	})

	spots, err := queries.ListPracticePlanSpotsInCategory(r.Context(), db.ListPracticePlanSpotsInCategoryParams{
		PracticeType: practiceType,
		PlanID:       planID,
		UserID:       user.ID,
	})
	if err != nil {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "There was an error retrieving your spots",
			Title:    "Database Error",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Database Error", http.StatusInternalServerError)
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
		planpages.EditExtraRepeatSpotList(spotInfo, planID, csrf.Token(r)).Render(r.Context(), w)
		return
	case "interleave_days":
		planpages.EditInterleaveDaysSpotList(spotInfo, planID, csrf.Token(r)).Render(r.Context(), w)
		return
	case "interleave":
		planpages.EditInterleaveSpotList(spotInfo, planID, csrf.Token(r)).Render(r.Context(), w)
		return
	case "new":
		planpages.EditNewSpotList(spotInfo, planID, csrf.Token(r)).Render(r.Context(), w)
		return
	default:
		http.Error(w, "Invalid practice type", http.StatusBadRequest)
	}
}

func (s *Server) getPiecesForPracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	activePlanID, ok := s.GetActivePracticePlanID(r.Context())
	if !ok {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot modify an inactive practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Cannot modify inactive plan", http.StatusBadRequest)
		return
	}
	planID := chi.URLParam(r, "planID")
	if activePlanID != planID {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot modify an inactive practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		})
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
			htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
				Message:  "There was an error retrieving your pieces",
				Title:    "Database Error",
				Variant:  "error",
				Duration: 3000,
			})
			http.Error(w, "Database Error", http.StatusInternalServerError)
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
			log.Default().Println(err)
			htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
				Message:  "There was an error retrieving your pieces",
				Title:    "Database Error",
				Variant:  "error",
				Duration: 3000,
			})
			http.Error(w, "Database Error", http.StatusInternalServerError)
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
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "That type of piece does not exist",
			Title:    "Invalid Piece Practice Type",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Invalid Spot practice type", http.StatusBadRequest)
		return
	}
	token := csrf.Token(r)
	planpages.AddPieceFormList(pieceInfo, token, planID, practiceType).Render(r.Context(), w)
}

func (s *Server) addPiecesToPracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	activePlanID, ok := s.GetActivePracticePlanID(r.Context())
	if !ok {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot modify an inactive practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Cannot modify inactive plan", http.StatusBadRequest)
		return
	}
	planID := chi.URLParam(r, "planID")
	if activePlanID != planID {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot modify an inactive practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Cannot modify inactive plan", http.StatusBadRequest)
		return
	}
	practiceType := chi.URLParam(r, "practiceType")
	queries := db.New(s.DB)

	r.ParseForm()

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
			htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
				Message:  "There was an error creating the piece",
				Title:    "Database Error",
				Variant:  "error",
				Duration: 3000,
			})
			http.Error(w, "Database Error", http.StatusInternalServerError)
			return
		}
	}

	htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Spots added to practice plan",
		Title:    "Success",
		Variant:  "success",
		Duration: 3000,
	})

	pieces, err := queries.ListPracticePlanPiecesInCategory(r.Context(), db.ListPracticePlanPiecesInCategoryParams{
		PracticeType: practiceType,
		PlanID:       planID,
		UserID:       user.ID,
	})
	if err != nil {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "There was an error retrieving your spots",
			Title:    "Database Error",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Database Error", http.StatusInternalServerError)
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
		planpages.EditRandomPieceList(pieceInfo, true, planID, csrf.Token(r)).Render(r.Context(), w)
		return
	case "starting_point":
		planpages.EditStartingPointList(pieceInfo, true, planID, csrf.Token(r)).Render(r.Context(), w)
		return
	default:
		http.Error(w, "Invalid practice type", http.StatusBadRequest)
	}
}

func (s *Server) duplicatePracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	planID := chi.URLParam(r, "planID")

	tx, err := s.DB.Begin()
	if err != nil {
		log.Default().Printf("Database error: %v\n", err)
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	queries := db.New(s.DB)
	qtx := queries.WithTx(tx)

	activePracticePlanID, ok := s.GetActivePracticePlanID(r.Context())
	if ok && planID == activePracticePlanID {
		htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot duplicate an active practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		})
		http.Error(w, "Cannot duplicate active plan", http.StatusBadRequest)
		return
	}

	oldPlan, err := qtx.GetPracticePlan(r.Context(), db.GetPracticePlanParams{
		ID:     planID,
		UserID: user.ID,
	})
	if err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	planPieces, err := qtx.GetPracticePlanWithPieces(r.Context(), db.GetPracticePlanWithPiecesParams{
		ID:     planID,
		UserID: user.ID,
	})
	if err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	planSpots, err := qtx.GetPracticePlanWithSpots(r.Context(), db.GetPracticePlanWithSpotsParams{
		ID:     planID,
		UserID: user.ID,
	})

	if err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	practiceSessionID := cuid2.Generate()
	if err := qtx.CreatePracticeSession(r.Context(), db.CreatePracticeSessionParams{
		ID:              practiceSessionID,
		UserID:          user.ID,
		DurationMinutes: 0,
		Date:            time.Now().Unix(),
	}); err != nil {
		log.Default().Println(err)
		http.Error(w, "Could not create practice session", http.StatusInternalServerError)
		return
	}

	newPlan, err := qtx.CreatePracticePlan(r.Context(), db.CreatePracticePlanParams{
		ID:                cuid2.Generate(),
		UserID:            user.ID,
		Intensity:         oldPlan.Intensity,
		PracticeSessionID: sql.NullString{Valid: true, String: practiceSessionID},
	})
	if err != nil {
		log.Default().Printf("Database error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, piece := range planPieces {
		if !piece.PieceID.Valid {
			continue
		}
		_, err := qtx.CreatePracticePlanPiece(r.Context(), db.CreatePracticePlanPieceParams{
			PracticePlanID: newPlan.ID,
			PieceID:        piece.PieceID.String,
			PracticeType:   piece.PiecePracticeType,
		})
		if err != nil {
			log.Default().Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	for _, spot := range planSpots {
		if !spot.SpotID.Valid {
			continue
		}
		_, err := qtx.CreatePracticePlanSpot(r.Context(), db.CreatePracticePlanSpotParams{
			PracticePlanID: newPlan.ID,
			SpotID:         spot.SpotID.String,
			PracticeType:   spot.SpotPracticeType,
		})
		if err != nil {
			log.Default().Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.SetActivePracticePlanID(r.Context(), newPlan.ID, user.ID)
	// update user (with newly added practice plan id) and practice plan manually before continuing
	user, err = queries.GetUserByID(r.Context(), user.ID)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	htmx.PushURL(r, "/library/plans/"+newPlan.ID)
	htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Your plan has been successfully duplicated",
		Title:    "Duplicated",
		Variant:  "success",
		Duration: 3000,
	})
	ctx := context.WithValue(r.Context(), "activePracticePlanID", newPlan.ID)
	ctx = context.WithValue(ctx, "user", user)
	s.renderPracticePlanPage(w, r.WithContext(ctx), newPlan.ID, user.ID)
}
