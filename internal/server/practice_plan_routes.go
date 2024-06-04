package server

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"practicebetter/internal/ck"
	"practicebetter/internal/components"
	"practicebetter/internal/config"
	"practicebetter/internal/db"
	"practicebetter/internal/pages/ewspages"
	"practicebetter/internal/pages/planpages"
	"practicebetter/internal/pages/readingpages"
	"slices"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/mavolin/go-htmx"
	"github.com/nrednav/cuid2"
)

func (s *Server) createPracticePlanForm(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	token := csrf.Token(r)
	queries := db.New(s.DB)

	activePieces, err := queries.ListActiveUserPieces(r.Context(), user.ID)
	if err != nil {
		s.DatabaseError(w, r, err, "Failed to load active pieces")
		return
	}
	s.HxRender(w, r, planpages.CreatePracticePlanPage(s, token, activePieces, planpages.PlanCreationErrors{}), "Create Practice Plan")
}

type PotentialInfrequentSpot struct {
	ID        string
	TimeSince time.Duration
}

func (s *Server) fixSpotSkipDays(ctx context.Context, spotID string, userID string) {
	queries := db.New(s.DB)
	if err := queries.UpdateSpotSkipDays(ctx, db.UpdateSpotSkipDaysParams{
		SkipDays: 1,
		UserID:   userID,
		SpotID:   spotID,
	}); err != nil {
		log.Default().Println("Error fixing spot skip days:", err)
	}
}

type PlanPieceInfo struct {
	NewSpotIDs               []string
	ExtraRepeatSpotIDs       []string
	InterleaveSpotIDs        []string
	PotentialInfrequentSpots []PotentialInfrequentSpot
	RandomSpotCount          int
	ExtraRepeatSpotCount     int
	CompletedSpotCount       int
}

func (s *Server) generatePiecePlanInfo(ctx context.Context, rows []db.GetPieceForPlanRow, failedNewSpotIDs map[string]struct{}, userID string) PlanPieceInfo {

	newSpotIDs := make([]string, 0, len(rows)/2)
	extraRepeatSpotIDs := make([]string, 0, len(rows)/4)
	interleaveSpotIDs := make([]string, 0, len(rows)/4)
	potentialInfrequentSpots := make([]PotentialInfrequentSpot, 0, len(rows)/4)
	randomSpotCount := 0
	extraRepeatSpotCount := 0
	completedSpotCount := 0
	for _, row := range rows {
		if !row.SpotStage.Valid || !row.SpotID.Valid {
			continue
		}
		switch row.SpotStage.String {
		case "repeat":
			// we're going to combine the lists later, so make need to prevent duplicates
			if _, ok := failedNewSpotIDs[row.SpotID.String]; !ok {
				newSpotIDs = append(newSpotIDs, row.SpotID.String)
			}
		case "extra_repeat":
			extraRepeatSpotIDs = append(extraRepeatSpotIDs, row.SpotID.String)
			extraRepeatSpotCount += 1
		case "interleave":
			interleaveSpotIDs = append(interleaveSpotIDs, row.SpotID.String)
		case "interleave_days":
			if !row.SpotSkipDays.Valid {
				go s.fixSpotSkipDays(ctx, row.SpotID.String, userID)
			}

			// we want to make sure that it has been more than the skip days value since the last time
			// this spot was practiced. I've made the offset 12 hours as a reasonable way to avoid adding
			// an extra day because someone practiced in the evening one day and in the morning the next
			if !row.SpotLastPracticed.Valid ||
				time.Since(time.Unix(row.SpotLastPracticed.Int64, 0)) > (time.Duration(row.SpotSkipDays.Int64)+1)*24*time.Hour+config.INFREQUENT_SPOT_OFFSET {
				potentialInfrequentSpots = append(potentialInfrequentSpots, PotentialInfrequentSpot{
					ID:        row.SpotID.String,
					TimeSince: time.Since(time.Unix(row.SpotLastPracticed.Int64, 0)),
				})
			}
		case "random":
			randomSpotCount += 1
		case "completed":
			completedSpotCount += 1
		default:
			continue
		}
	}
	return PlanPieceInfo{
		NewSpotIDs:               newSpotIDs,
		ExtraRepeatSpotIDs:       extraRepeatSpotIDs,
		InterleaveSpotIDs:        interleaveSpotIDs,
		PotentialInfrequentSpots: potentialInfrequentSpots,
		RandomSpotCount:          randomSpotCount,
		ExtraRepeatSpotCount:     extraRepeatSpotCount,
		CompletedSpotCount:       completedSpotCount,
	}

}

func (s *Server) createPracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	if err := r.ParseForm(); err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Invalid information in form",
			Title:    "Invalid Plan",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Invalid Form Data", http.StatusBadRequest)
		return
	}

	pieceIDs := r.Form["pieces"]

	if len(pieceIDs) == 0 {
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You need to select at least one piece to practice.",
			Title:    "Invalid Plan",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		queries := db.New(s.DB)
		activePieces, err := queries.ListActiveUserPieces(r.Context(), user.ID)
		if err != nil {
			s.DatabaseError(w, r, err, "Failed to load active pieces")
			return
		}
		token := csrf.Token(r)
		s.HxRender(w, r, planpages.CreatePracticePlanPage(s, token, activePieces,
			planpages.PlanCreationErrors{
				Pieces: "You need to select at least one piece to practice.",
			},
		), "Create Practice Plan")
		return
	}

	tx, err := s.DB.Begin()
	if err != nil {
		log.Default().Printf("Database error: %v\n", err)
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Default().Println("Error in createPractciePlan rollback:", err)
		}
	}()
	queries := db.New(s.DB)
	qtx := queries.WithTx(tx)

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
		for _, spot := range failedNewSpots {
			failedNewSpotIDs[spot.SpotID] = struct{}{}
		}
	}

	newPlan, err := qtx.CreatePracticePlan(r.Context(), db.CreatePracticePlanParams{
		ID:        cuid2.Generate(),
		UserID:    user.ID,
		Intensity: r.FormValue("intensity"),
	})
	if err != nil {
		log.Default().Printf("Database error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var maxNewSpots int64
	var maxInfrequentSpots int
	var maxInterleaveSpots int
	var maxSightReading int
	switch r.FormValue("intensity") {
	case "light":
		maxNewSpots = config.LIGHT_MAX_NEW_SPOTS
		maxInfrequentSpots = config.LIGHT_MAX_INFREQUENT_SPOTS
		maxInterleaveSpots = config.LIGHT_MAX_INTERLEAVE_SPOTS
		maxSightReading = config.LIGHT_SIGHT_READING
	case "medium":
		maxNewSpots = config.MEDIUM_MAX_NEW_SPOTS
		maxInfrequentSpots = config.MEDIUM_MAX_INFREQUENT_SPOTS
		maxInterleaveSpots = config.MEDIUM_MAX_INTERLEAVE_SPOTS
		maxSightReading = config.MEDIUM_SIGHT_READING
	case "heavy":
		maxNewSpots = config.HEAVY_MAX_NEW_SPOTS
		maxInfrequentSpots = config.HEAVY_MAX_INFREQUENT_SPOTS
		maxInterleaveSpots = config.HEAVY_MAX_INTERLEAVE_SPOTS
		maxSightReading = config.HEAVY_SIGHT_READING
	}

	if r.FormValue("scale") == "on" {
		workingScales, err := qtx.ListWorkingScales(r.Context(), user.ID)
		if err != nil {
			log.Default().Println(err)
		}
		if len(workingScales) > 0 {
			for i, scale := range workingScales {
				_, err = qtx.CreatePracticePlanScaleWithIdx(r.Context(), db.CreatePracticePlanScaleWithIdxParams{
					PracticePlanID: newPlan.ID,
					UserScaleID:    scale.ID,
					Idx:            int64(i),
				})
				if err != nil {
					log.Default().Println(err)
				}
			}
		} else {
			var selectedScaleID int64
			if r.FormValue("modal-scales") == "on" {
				allScales, err := qtx.ListScales(r.Context())
				if err != nil {
					s.DatabaseError(w, r, err, "Failed to load scales")
					return
				}
				rand.Shuffle(len(allScales), func(i, j int) {
					allScales[i], allScales[j] = allScales[j], allScales[i]
				})
				selectedScaleID = allScales[0].ID
			} else {
				basicScales, err := qtx.ListBasicScales(r.Context())
				if err != nil {
					s.DatabaseError(w, r, err, "Failed to load scales")
					return
				}
				rand.Shuffle(len(basicScales), func(i, j int) {
					basicScales[i], basicScales[j] = basicScales[j], basicScales[i]
				})
				selectedScaleID = basicScales[0].ID
			}

			var userScaleID string
			userScaleID, err = qtx.CheckForUserScale(r.Context(), db.CheckForUserScaleParams{
				UserID:  user.ID,
				ScaleID: selectedScaleID,
			})
			if err != nil || userScaleID == "" {
				userScale, err := qtx.CreateUserScale(r.Context(), db.CreateUserScaleParams{
					ID:            cuid2.Generate(),
					UserID:        user.ID,
					ScaleID:       selectedScaleID,
					PracticeNotes: "",
					Reference:     "",
				})
				if err != nil {
					s.DatabaseError(w, r, err, "Failed to create user scale")
					return
				}
				userScaleID = userScale.ID
			}

			_, err = qtx.CreatePracticePlanScaleWithIdx(r.Context(), db.CreatePracticePlanScaleWithIdxParams{
				PracticePlanID: newPlan.ID,
				UserScaleID:    userScaleID,
				Idx:            0,
			})
			if err != nil {
				s.DatabaseError(w, r, err, "Failed to create practice plan scale")
				return
			}
		}
	}

	if r.FormValue("reading") == "on" {
		items, err := qtx.ListIncompleteUserReadingItems(r.Context(), user.ID)
		if err != nil {
			log.Default().Println(err)
		}
		rand.Shuffle(len(items), func(i, j int) {
			items[i], items[j] = items[j], items[i]
		})
		i := 0
		for i < int(math.Min(float64(len(items)), float64(maxSightReading))) {
			_, err = qtx.CreatePracticePlanReadingWithIdx(r.Context(), db.CreatePracticePlanReadingWithIdxParams{
				PracticePlanID: newPlan.ID,
				ReadingID:      items[i].ID,
				Idx:            int64(i),
			})
			if err != nil {
				log.Default().Println(err)
			}
			i++
		}

	}

	maybeNewSpotLists := make([][]string, 0, len(pieceIDs))
	potentialInfrequentSpots := make([]PotentialInfrequentSpot, 0, len(pieceIDs)*10)
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
			s.DatabaseError(w, r, err, "Could not get piece")
			return
		}

		pieceInfo := s.generatePiecePlanInfo(r.Context(), pieceRows, failedNewSpotIDs, user.ID)

		extraRepeatSpotIDs = append(extraRepeatSpotIDs, pieceInfo.ExtraRepeatSpotIDs...)
		interleaveSpotIDs = append(interleaveSpotIDs, pieceInfo.InterleaveSpotIDs...)
		potentialInfrequentSpots = append(potentialInfrequentSpots, pieceInfo.PotentialInfrequentSpots...)

		// Only new spots if there aren't too many extra repeat/random spots.
		if (pieceInfo.ExtraRepeatSpotCount + pieceInfo.RandomSpotCount) < config.MAX_ALLOWED_RANDOM_SPOTS {
			maybeNewSpotLists = append(maybeNewSpotLists, pieceInfo.NewSpotIDs)
		}

		if r.FormValue("practice_random_single") == "on" &&
			pieceInfo.RandomSpotCount > 2 {
			randomSpotPieceIDs = append(randomSpotPieceIDs, pieceID)
		}

		if r.FormValue("practice_starting_point") == "on" &&
			pieceInfo.CompletedSpotCount > 5 {
			startingPointPieceIDs = append(startingPointPieceIDs, pieceID)
		}
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
			s.DatabaseError(w, r, err, "Could not add extra repeat spot")
			return
		}
	}

	if r.FormValue("practice_interleave") == "on" {
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
				s.DatabaseError(w, r, err, "Could not add interleave spot")
				return
			}
		}

		// prioritize infrequent spots with the spots that are the lest recently practiced first
		slices.SortFunc(potentialInfrequentSpots, func(a, b PotentialInfrequentSpot) int {
			return cmp.Compare(b.TimeSince, a.TimeSince)
		})

		// infrequent spots
		for i, spot := range potentialInfrequentSpots {
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
				s.DatabaseError(w, r, err, "Could not add infrequent spot")
				return
			}
		}
	}

	if r.FormValue("practice_new") == "on" {
		var newSpotIdx int64 = 0
		// Put in all the failed spots from the previous day
		for spotID := range failedNewSpotIDs {
			_, err := qtx.CreatePracticePlanSpotWithIdx(r.Context(), db.CreatePracticePlanSpotWithIdxParams{
				PracticePlanID: newPlan.ID,
				SpotID:         spotID,
				PracticeType:   "new",
				Idx:            newSpotIdx,
			})
			if err != nil {
				s.DatabaseError(w, r, err, "Could not add spot")
				return
			}
			newSpotIdx++
		}

		// then if there's room, put in one spot from each piece, and save the others to be randomized
		additionalNewSpots := make([]string, 0, len(maybeNewSpotLists)*10)
		for _, pieceSpotList := range maybeNewSpotLists {
			if newSpotIdx >= maxNewSpots {
				break
			}
			// shuffle the spots will be random
			rand.Shuffle(len(pieceSpotList), func(i, j int) {
				pieceSpotList[i], pieceSpotList[j] = pieceSpotList[j], pieceSpotList[i]
			})
			for i, spotID := range pieceSpotList {
				if newSpotIdx >= maxNewSpots {
					break
				}
				if i == 0 {
					_, err := qtx.CreatePracticePlanSpotWithIdx(r.Context(), db.CreatePracticePlanSpotWithIdxParams{
						PracticePlanID: newPlan.ID,
						SpotID:         spotID,
						PracticeType:   "new",
						Idx:            newSpotIdx,
					})
					if err != nil {
						s.DatabaseError(w, r, err, "Could not add spot")
						return
					}
					newSpotIdx++
				} else {
					additionalNewSpots = append(additionalNewSpots, spotID)
				}
			}

		}

		// new spots
		// if there's room, shuffle the remaining spots and add them until it's full
		if newSpotIdx < maxNewSpots {
			rand.Shuffle(len(additionalNewSpots), func(i, j int) {
				additionalNewSpots[i], additionalNewSpots[j] = additionalNewSpots[j], additionalNewSpots[i]
			})
			for _, spotID := range additionalNewSpots {
				if newSpotIdx >= maxNewSpots {
					break
				}
				_, err := qtx.CreatePracticePlanSpotWithIdx(r.Context(), db.CreatePracticePlanSpotWithIdxParams{
					PracticePlanID: newPlan.ID,
					SpotID:         spotID,
					PracticeType:   "new",
					Idx:            newSpotIdx,
				})
				if err != nil {
					s.DatabaseError(w, r, err, "Could not add spot")
					return
				}
				newSpotIdx++
			}

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
			s.DatabaseError(w, r, err, "Could not add random spot piece")
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
			s.DatabaseError(w, r, err, "Could not add random starting point piece")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.SetActivePracticePlanID(r.Context(), newPlan.ID, user.ID); err != nil {
		s.DatabaseError(w, r, err, "Could not set active plan")
		return
	}
	// update user (with newly added practice plan id) and practice plan manually before continuing
	user, err = queries.GetUserByID(r.Context(), user.ID)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.SM.RenewToken(r.Context()); err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not renew session",
			Title:    "Failed to renew",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.ClearLastBreak(r.Context())
	s.SetLastBreak(r.Context(), newPlan.ID)

	if r.FormValue("customize") == "on" {
		htmx.PushURL(r, "/library/plans/"+newPlan.ID+"/edit")
		ctx := context.WithValue(r.Context(), ck.ActivePlanKey, newPlan.ID)
		ctx = context.WithValue(ctx, ck.UserKey, user)
		s.renderEditPracticePlanPage(w, r.WithContext(ctx), newPlan.ID, user.ID)
	} else {
		htmx.PushURL(r, "/library/plans/"+newPlan.ID)
		ctx := context.WithValue(r.Context(), ck.ActivePlanKey, newPlan.ID)
		ctx = context.WithValue(ctx, ck.UserKey, user)
		s.renderPracticePlanPage(w, r.WithContext(ctx), newPlan.ID, user.ID)
	}
}

func (s *Server) singlePracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	planID := chi.URLParam(r, "planID")

	s.renderPracticePlanPage(w, r, planID, user.ID)
}

func (s *Server) renderPracticePlanPage(w http.ResponseWriter, r *http.Request, planID string, userID string) {
	queries := db.New(s.DB)
	totalItems := 0
	completedItems := 0
	planPieces, err := queries.GetPracticePlanWithPieces(r.Context(), db.GetPracticePlanWithPiecesParams{
		ID:     planID,
		UserID: userID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not get practice plan with pieces")
		return
	}

	planSpots, err := queries.GetPracticePlanWithSpots(r.Context(), db.GetPracticePlanWithSpotsParams{
		ID:     planID,
		UserID: userID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not get practice plan with spots")
		return
	}

	planScales, err := queries.GetPracticePlanWithScales(r.Context(), db.GetPracticePlanWithScalesParams{
		PracticePlanID: planID,
		UserID:         userID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not get practice plan with scales")
		return
	}

	planReading, err := queries.GetPracticePlanWithReading(r.Context(), db.GetPracticePlanWithReadingParams{
		PracticePlanID: planID,
		UserID:         userID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not get practice plan with scales")
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

	if planData.IsActive {
		needsBreak, err := s.needsBreak(r.Context(), planID, userID)
		if err != nil {
			log.Default().Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		planData.NeedsBreak = needsBreak
	} else {
		planData.NeedsBreak = false
	}

	planData.Scales = make([]planpages.PracticePlanScale, 0, len(planScales))
	for _, row := range planScales {
		totalItems++
		if row.ScaleCompleted {
			completedItems++
		}
		var scale planpages.PracticePlanScale
		scale.Completed = row.ScaleCompleted
		scale.UserScaleInfo = components.UserScaleInfo{
			UserScaleID:   row.UserScaleID,
			KeyName:       row.ScaleKeyName,
			ModeName:      row.ScaleMode,
			PracticeNotes: row.ScalePracticeNotes,
			Reference:     row.ScaleReference,
		}
		planData.Scales = append(planData.Scales, scale)
	}

	planData.SightReadingItems = make([]components.PlanSightReadingItem, 0, len(planReading))
	for _, row := range planReading {
		totalItems++
		if row.Completed {
			completedItems++
		}
		var item components.PlanSightReadingItem
		item.Completed = row.ReadingCompleted
		item.ReadingID = row.ReadingID
		item.Title = row.ReadingTitle
		if row.ReadingComposer.Valid {
			item.Composer = row.ReadingComposer.String
		}
		planData.SightReadingItems = append(planData.SightReadingItems, item)
	}
	log.Default().Println(planData.SightReadingItems)

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
			piece.Completed = row.PieceCompleted
			if row.PieceComposer.Valid {
				piece.Composer = row.PieceComposer.String
			} else {
				piece.Composer = "Unknown"
			}

			if row.PiecePracticeType == "random_spots" {
				planData.RandomSpotsPieces = append(planData.RandomSpotsPieces, piece)
			} else if row.PiecePracticeType == "starting_point" {
				planData.RandomStartPieces = append(planData.RandomStartPieces, piece)
			}

		}
	}
	// TODO: if spot has evaluation and plan isn't active, complete it and handle the promotion, and clear the evaluation.
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
	s.HxRender(w, r, planpages.PracticePlanPage(s, planData, token), "Practice Plan")
}

func (s *Server) editPracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
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

	planScales, err := queries.GetPracticePlanWithScales(r.Context(), db.GetPracticePlanWithScalesParams{
		PracticePlanID: planID,
		UserID:         userID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not get practice plan with scales")
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

	planData.Scales = make([]planpages.PracticePlanScale, 0, len(planScales))
	for _, row := range planScales {
		var scale planpages.PracticePlanScale
		scale.Completed = row.ScaleCompleted
		scale.UserScaleInfo = components.UserScaleInfo{
			UserScaleID:   row.UserScaleID,
			KeyName:       row.ScaleKeyName,
			ModeName:      row.ScaleMode,
			PracticeNotes: row.ScalePracticeNotes,
			Reference:     row.ScaleReference,
		}
		planData.Scales = append(planData.Scales, scale)
	}
	planData.TotalItems = totalItems
	planData.CompletedItems = completedItems

	token := csrf.Token(r)
	s.HxRender(w, r, planpages.EditPracticePlanPage(planData, token), "Customize Practice Plan")
}

type UpdatePlanProgressEvent struct {
	Completed int `json:"completed"`
	Total     int `json:"total"`
}

func (s *Server) redirectToNextPlanItem(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
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
		s.DatabaseError(w, r, err, "Could not get next plan item")
		return
	}
	isPlanPage := r.Header.Get("X-Plan-Page") == "true"

	scales, err := queries.GetPracticePlanWithIncompleteScales(r.Context(), db.GetPracticePlanWithIncompleteScalesParams{
		PracticePlanID: activePracticePlanID,
		UserID:         user.ID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not get next plan item")
		return
	}

	if len(scales) > 0 && isPlanPage {
		htmx.Retarget(r, "#practice-scale-dialog-contents")
		if err := htmx.Trigger(r, "ShowModal", "practice-scale-dialog"); err != nil {
			log.Default().Println(err)
		}
		info := ewspages.ScaleInfo{
			ID:            scales[0].UserScaleID,
			KeyName:       scales[0].ScaleKeyName,
			Mode:          scales[0].ScaleMode,
			PracticeNotes: scales[0].ScalePracticeNotes,
			LastPracticed: scales[0].ScaleLastPracticed,
			Reference:     scales[0].ScaleReference,
			Working:       scales[0].ScaleWorking,
		}
		htmx.PreventPushURL(r)
		if err := ewspages.PracticeScaleDisplay(info, csrf.Token(r)).Render(r.Context(), w); err != nil {
			log.Default().Println(err)
		}
		return
	}

	readingItems, err := queries.GetPracticePlanWithIncompleteReading(r.Context(), db.GetPracticePlanWithIncompleteReadingParams{
		PracticePlanID: activePracticePlanID,
		UserID:         user.ID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not get next plan item")
		return
	}

	if len(readingItems) > 0 && isPlanPage {
		htmx.Retarget(r, "#practice-reading-dialog-contents")
		htmx.Reswap(r, "innerHTML")
		if err := htmx.Trigger(r, "ShowModal", "practice-reading-dialog"); err != nil {
			log.Default().Println(err)
		}
		info := readingpages.SingleReadingItemInfo{
			ID:        readingItems[0].ReadingID,
			Title:     readingItems[0].ReadingTitle,
			Composer:  readingItems[0].ReadingComposer,
			Completed: readingItems[0].ReadingCompleted,
			Info:      readingItems[0].ReadingInfo,
		}
		htmx.PreventPushURL(r)
		if err := readingpages.PracticeReadingDisplay(info, csrf.Token(r)).Render(r.Context(), w); err != nil {
			log.Default().Println(err)
		}
		return
	}

	hasInfrequent, err := queries.HasIncompleteInfrequentSpots(r.Context(), db.HasIncompleteInfrequentSpotsParams{
		PlanID: planID,
		UserID: user.ID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not get next plan item")
		return
	}
	if hasInfrequent && isPlanPage {
		htmx.Retarget(r, "#"+components.INFREQUENT_SPOT_DIALOG_CONTENTS_ID)
		htmx.Reswap(r, "innerHTML")
		if err := htmx.Trigger(r, "ShowModal", components.INFREQUENT_SPOT_DIALOG_ID); err != nil {
			log.Default().Println(err)
		}
		htmx.PreventPushURL(r)
		s.startInfrequentPracticing(w, r)
		return
	}

	extraRepeatSpots, err := queries.GetPracticePlanIncompleteExtraRepeatSpots(r.Context(), db.GetPracticePlanIncompleteExtraRepeatSpotsParams{
		PlanID: planID,
		UserID: user.ID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not get next plan item")
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
		s.DatabaseError(w, r, err, "Could not get next plan item")
		return
	}
	htmx.Retarget(r, "#main-content")
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
		s.DatabaseError(w, r, err, "Could not get next plan item")
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
		s.DatabaseError(w, r, err, "Could not get next plan item")
		return
	}
	if len(newSpots) > 0 {
		spotID := newSpots[0].SpotID
		pieceID := newSpots[0].SpotPieceID
		htmx.Redirect(r, "/library/pieces/"+pieceID+"/spots/"+spotID+"/practice/repeat")
		http.Redirect(w, r, "/library/pieces/"+pieceID+"/spots/"+spotID+"/practice/repeat", http.StatusSeeOther)
		return
	}

	if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "No more items to practice. Check you infrequent and interleave spots one last time and you are done.",
		Title:    "Almost Done",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}

	htmx.Redirect(r, "/library/plans/"+planID)
	http.Redirect(w, r, "/library/plans/"+planID, http.StatusSeeOther)
}

/*
excellent after 7 days = promote
poor ever = demote
fine = stay
fine after 10 days = demote
*/

func (s *Server) planList(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
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
		Limit:  config.ItemsPerPage,
		Offset: int64((pageNum - 1) * config.ItemsPerPage),
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
	totalPages := int(math.Ceil(float64(totalPlans) / float64(config.ItemsPerPage)))
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
			CompletedItems: p.CompletedSpotsCount + p.CompletedPiecesCount + p.CompletedScalesCount,
			TotalItems:     p.PiecesCount + p.SpotsCount + p.ScalesCount,
			PieceTitles:    pieceTitlesForPlanCard(p.PieceTitles, p.SpotPieceTitles),
		}

		planInfo = append(planInfo, nextPlanInfo)
	}
	w.WriteHeader(http.StatusOK)
	s.HxRender(w, r, planpages.PlanList(planInfo, pageNum, totalPages), "Your Practice Plans")
}

func (s *Server) deletePracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	planID := chi.URLParam(r, "planID")
	queries := db.New(s.DB)
	if err := queries.DeletePracticePlan(r.Context(), db.DeletePracticePlanParams{ID: planID, UserID: user.ID}); err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not delete practice plan",
			Title:    "Error",
			Variant:  "danger",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Could not delete practice plan", http.StatusInternalServerError)
		return
	}
	if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Practice plan deleted",
		Title:    "Deleted!",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}
	htmx.PushURL(r, "/library/plans")

	// refresh user from database in case the active plan was deleted
	user, err := queries.GetUserByID(r.Context(), user.ID)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	ctx := context.WithValue(r.Context(), ck.UserKey, user)
	if !user.ActivePracticePlanID.Valid {
		ctx = context.WithValue(ctx, ck.ActivePlanKey, "")
	}
	w.WriteHeader(http.StatusOK)
	s.renderPlanListPage(w, r.WithContext(ctx), user.ID, 1)
}

func (s *Server) resumePracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
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
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You are already using this practice plan.",
			Title:    "Already Active",
			Variant:  "warning",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "You are already using this practice plan.", http.StatusBadRequest)
		return

	}
	if time.Since(time.Unix(plan.Date, 0)) > config.RESUME_PLAN_TIME_LIMIT {
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot resume a practice plan this old. Please create a new one instead.",
			Title:    "Too Old",
			Variant:  "warning",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "You cannot resume a practice plan this old. Please create a new one instead.", http.StatusBadRequest)
		return
	}
	err = s.SetActivePracticePlanID(r.Context(), planID, user.ID)
	if err != nil {
		s.DatabaseError(w, r, err, "Could not activate this practice plan.")
		return
	}
	// update user (with newly added practice plan id) and practice plan manually before continuing
	user, err = queries.GetUserByID(r.Context(), user.ID)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "You resumed your practice plan!",
		Title:    "Resumed!",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}
	ctx := context.WithValue(r.Context(), ck.ActivePlanKey, plan.ID)
	ctx = context.WithValue(ctx, ck.UserKey, user)
	s.renderPracticePlanPage(w, r.WithContext(ctx), plan.ID, user.ID)
}

func (s *Server) stopPracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	planID := chi.URLParam(r, "planID")

	err := r.ParseForm()
	if err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not stop practice plan.",
			Title:    "Form Error",
			Variant:  "danger",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Form Error", http.StatusBadRequest)
		return
	}

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
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "This plan is not active",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "This plan is not active", http.StatusBadRequest)
		return

	}
	err = queries.ClearActivePracticePlan(r.Context(), user.ID)
	if err != nil {
		s.DatabaseError(w, r, err, "Could not stop practice plan")
		return
	}
	s.completeInterleaveSpots(w, r, planID, user.ID)

	err = queries.CompletePracticePlan(r.Context(), db.CompletePracticePlanParams{
		ID:     plan.ID,
		UserID: user.ID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not stop practice plan")
		return
	}
	// update user (with removed practice plan id) and practice plan manually before continuing
	user, err = queries.GetUserByID(r.Context(), user.ID)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Come back tomorrow!",
		Title:    "Done Practicing",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}
	ctx := context.WithValue(r.Context(), ck.ActivePlanKey, "")
	ctx = context.WithValue(ctx, ck.UserKey, user)
	s.renderPracticePlanPage(w, r.WithContext(ctx), plan.ID, user.ID)
}

func (s *Server) duplicatePracticePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	planID := chi.URLParam(r, "planID")

	tx, err := s.DB.Begin()
	if err != nil {
		log.Default().Printf("Database error: %v\n", err)
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Default().Println(err)
		}
	}()
	queries := db.New(s.DB)
	qtx := queries.WithTx(tx)

	activePracticePlanID, ok := s.GetActivePracticePlanID(r.Context())
	if ok && planID == activePracticePlanID {
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "You cannot duplicate an active practice plan",
			Title:    "Not Active",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
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

	planScales, err := qtx.GetPracticePlanWithScales(r.Context(), db.GetPracticePlanWithScalesParams{
		PracticePlanID: planID,
		UserID:         user.ID,
	})
	if err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	planReading, err := qtx.GetPracticePlanWithReading(r.Context(), db.GetPracticePlanWithReadingParams{
		PracticePlanID: planID,
		UserID:         user.ID,
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

	newPlan, err := qtx.CreatePracticePlan(r.Context(), db.CreatePracticePlanParams{
		ID:        cuid2.Generate(),
		UserID:    user.ID,
		Intensity: oldPlan.Intensity,
	})
	if err != nil {
		log.Default().Printf("Database error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i, scale := range planScales {
		_, err := qtx.CreatePracticePlanScaleWithIdx(r.Context(), db.CreatePracticePlanScaleWithIdxParams{
			PracticePlanID: newPlan.ID,
			UserScaleID:    scale.UserScaleID,
			Idx:            int64(i),
		})
		if err != nil {
			log.Default().Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	for i, item := range planReading {
		_, err := qtx.CreatePracticePlanReadingWithIdx(r.Context(), db.CreatePracticePlanReadingWithIdxParams{
			PracticePlanID: newPlan.ID,
			ReadingID:      item.ReadingID,
			Idx:            int64(i),
		})
		if err != nil {
			log.Default().Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
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
		s.DatabaseError(w, r, err, "Database error")
		return
	}

	if err := s.SetActivePracticePlanID(r.Context(), newPlan.ID, user.ID); err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Failed to set active plan",
			Title:    "Failed",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Failed to set active plan", http.StatusInternalServerError)
		return
	}
	// update user (with newly added practice plan id) and practice plan manually before continuing
	user, err = queries.GetUserByID(r.Context(), user.ID)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	htmx.PushURL(r, "/library/plans/"+newPlan.ID)
	if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Your plan has been successfully duplicated",
		Title:    "Duplicated",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}
	ctx := context.WithValue(r.Context(), ck.ActivePlanKey, newPlan.ID)
	ctx = context.WithValue(ctx, ck.UserKey, user)

	s.ClearLastBreak(r.Context())
	s.SetLastBreak(r.Context(), newPlan.ID)
	s.renderPracticePlanPage(w, r.WithContext(ctx), newPlan.ID, user.ID)
}

func (s *Server) takeABreak(w http.ResponseWriter, r *http.Request) {
	planID, ok := r.Context().Value(ck.ActivePlanKey).(string)
	if !ok {
		http.Error(w, "No active plan", http.StatusBadRequest)
		return
	}

	s.SetLastBreak(r.Context(), planID)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) needsBreak(ctx context.Context, planID string, userID string) (bool, error) {
	// check whether the last break was too long ago
	lastBreakTime, ok := s.GetLastBreak(ctx, planID)
	if !ok {
		s.SetLastBreak(ctx, planID)
		return false, fmt.Errorf("Could not get last break")
	}
	if time.Since(lastBreakTime) > config.TIME_BETWEEN_BREAKS {
		queries := db.New(s.DB)
		lastPracticed, err := queries.GetPlanLastPracticed(ctx, db.GetPlanLastPracticedParams{
			ID:     planID,
			UserID: userID,
		})
		if err != nil {
			log.Default().Println(err)
		} else {
			log.Default().Println(lastPracticed)
		}
		if !lastPracticed.Valid {
			s.SetLastBreak(ctx, planID)
			return false, nil
		}
		// if you haven't practiced anything in more than an entire session of time, you don't need a break
		if err == nil && lastPracticed.Valid && time.Since(time.Unix(lastPracticed.Int64, 0)) > 2*config.TIME_BETWEEN_BREAKS {
			s.SetLastBreak(ctx, planID)
			return false, nil
		}
		return true, nil
	} else {
		return false, nil
	}
}

func (s *Server) shouldRecommendBreak(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	planID, ok := r.Context().Value(ck.ActivePlanKey).(string)
	if !ok {
		http.Error(w, "No active plan", http.StatusBadRequest)
		return
	}
	shouldBreak, err := s.needsBreak(r.Context(), planID, user.ID)
	if err != nil {
		log.Default().Println(err)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(fmt.Sprintf("%t", shouldBreak))); err != nil {
		log.Default().Println(err)
	}
}

func (s *Server) lastBreak(w http.ResponseWriter, r *http.Request) {
	planID, ok := r.Context().Value(ck.ActivePlanKey).(string)
	if !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	lastBreakTime, ok := s.GetLastBreak(r.Context(), planID)
	if !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(lastBreakTime.String())); err != nil {
		log.Default().Println(err)
	}
}
