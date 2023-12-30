package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"practicebetter/internal/db"
	"practicebetter/internal/pages/pspages"
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
	s.HxRender(w, r, pspages.CreatePracticePlanPage(s, token, activePieces), "Create Practice Agenda")
}

// TODO: possibly change back to new spots period and calculate spots per piece each time using type casting and rounding to get the right number
const (
	LIGHT_NEW_SPOTS_PER_PIECE  = 1
	MEDIUM_NEW_SPOTS_PER_PIECE = 2
	HEAVY_NEW_SPOTS_PER_PIECE  = 3
)

// TODO: check that a piece has random spots before putting it in random spots practice
// TODO: practice plans can generate with no spots or pieces leading to a not found

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

	var numNewSpots int
	if r.FormValue("practice_new") != "on" {
		numNewSpots = 0
	} else if r.FormValue("intensity") == "light" {
		numNewSpots = LIGHT_NEW_SPOTS_PER_PIECE
	} else if r.FormValue("intensity") == "medium" {
		numNewSpots = MEDIUM_NEW_SPOTS_PER_PIECE
	} else if r.FormValue("intensity") == "heavy" {
		numNewSpots = HEAVY_NEW_SPOTS_PER_PIECE
	}

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

		if r.FormValue("practice_starting_point") == "on" {
			_, err := qtx.CreatePracticePlanPiece(r.Context(), db.CreatePracticePlanPieceParams{
				PracticePlanID: newPlan.ID,
				PieceID:        pieceID,
				PracticeType:   "starting_point",
			})
			if err != nil {
				log.Default().Printf("Database error: %v\n", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		maybeNewSpots := make([]db.GetPieceForPlanRow, 0, len(pieceRows)/2)
		canRandomSpotsPractice := false

		for _, row := range pieceRows {
			if !row.SpotStage.Valid || !row.SpotID.Valid {
				continue
			}
			if row.SpotStage.String == "repeat" && r.FormValue("practice_new") == "on" {
				maybeNewSpots = append(maybeNewSpots, row)
				// TODO: change more repeat to be extra repeat everywhere
			} else if row.SpotStage.String == "extra_repeat" {
				_, err := qtx.CreatePracticePlanSpot(r.Context(), db.CreatePracticePlanSpotParams{
					PracticePlanID: newPlan.ID,
					SpotID:         row.SpotID.String,
					PracticeType:   "extra_repeat",
				})
				if err != nil {
					log.Default().Printf("Database error: %v\n", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else if row.SpotStage.String == "interleave" {
				_, err := qtx.CreatePracticePlanSpot(r.Context(), db.CreatePracticePlanSpotParams{
					PracticePlanID: newPlan.ID,
					SpotID:         row.SpotID.String,
					PracticeType:   "interleave",
				})
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else if row.SpotStage.String == "interleave_days" {
				// if the spot doesn't have a last practiced date, or if it wasn't practiced yesterday (roughly)
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
					_, err := qtx.CreatePracticePlanSpot(r.Context(), db.CreatePracticePlanSpotParams{
						PracticePlanID: newPlan.ID,
						SpotID:         row.SpotID.String,
						PracticeType:   "interleave_days",
					})
					if err != nil {
						log.Default().Printf("Database error: %v\n", err)
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}
			} else if row.SpotStage.String == "random" {
				canRandomSpotsPractice = true
			}
		}

		newSpotIndexes := make(map[int]struct{}, 0)
		if len(maybeNewSpots) < numNewSpots {
			numNewSpots = len(maybeNewSpots)
		} else {
			for i := 0; i < numNewSpots; i++ {
				var nextSpotIndex int
				// this index means nothing, just there to avoid an infinite loop
				for j := 0; j < 100; j++ {
					nextSpotIndex = rand.Intn(len(maybeNewSpots))
					if _, ok := newSpotIndexes[nextSpotIndex]; !ok {
						break
					}
				}
				// there's a small change that it went around 100 times, so we'll double check to avoid duplicates
				if _, ok := newSpotIndexes[nextSpotIndex]; ok {
					break
				}
				newSpotIndexes[nextSpotIndex] = struct{}{}
				_, err := qtx.CreatePracticePlanSpot(r.Context(), db.CreatePracticePlanSpotParams{
					PracticePlanID: newPlan.ID,
					SpotID:         maybeNewSpots[nextSpotIndex].SpotID.String,
					PracticeType:   "new",
				})
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		if r.FormValue("practice_random_single") == "on" && canRandomSpotsPractice {
			_, err := qtx.CreatePracticePlanPiece(r.Context(), db.CreatePracticePlanPieceParams{
				PracticePlanID: newPlan.ID,
				PieceID:        pieceID,
				PracticeType:   "random_spots",
			})
			if err != nil {
				log.Default().Printf("Database error: %v\n", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
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
	ctx := context.WithValue(r.Context(), "activePracticePlanID", newPlan.ID)
	ctx = context.WithValue(ctx, "user", user)
	s.renderPracticePlanPage(w, r.WithContext(ctx), newPlan.ID, user.ID)
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

	var planData pspages.PracticePlanData
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
			var piece pspages.PracticePlanPiece
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
			var spot pspages.PracticePlanSpot
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
	// ctx := context.WithValue(r.Context(), "activePracticePlanID", activePracticePlanID)
	s.HxRender(w, r, pspages.PracticePlanPage(planData, token), "Practice Plan")
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

	spotInfo := make([]pspages.PracticePlanSpot, 0, len(interleaveDaysSpots))
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
		spotInfo = append(spotInfo, pspages.PracticePlanSpot{
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
	pspages.PracticePlanInterleaveDaysSpots(spotInfo, planID, token, true, true).Render(r.Context(), w)
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
		pieceID := randomPieces[0].PieceID
		htmx.Redirect(r, "/library/pieces/"+pieceID+"/practice/random")
		http.Redirect(w, r, "/library/pieces/"+pieceID+"/practice/random", http.StatusSeeOther)
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
		pieceID := startingPointPieces[0].PieceID
		htmx.Redirect(r, "/library/pieces/"+pieceID+"/practice/starting-point")
		http.Redirect(w, r, "/library/pieces/"+pieceID+"/practice/starting-point", http.StatusSeeOther)
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

	spotInfo := make([]pspages.PracticePlanSpot, 0, len(interleaveSpots))
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
		spotInfo = append(spotInfo, pspages.PracticePlanSpot{
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
	pspages.PracticePlanInterleaveSpots(spotInfo, planID, token, true, true, true).Render(r.Context(), w)

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
	spotInfo := make([]pspages.PracticePlanSpot, 0, len(interleaveSpots))
	for _, interleaveSpot := range interleaveSpots {
		if !interleaveSpot.Completed {
			allCompleted = false
		}
		spotInfo = append(spotInfo, pspages.PracticePlanSpot{
			ID:         interleaveSpot.SpotID,
			Name:       interleaveSpot.SpotName.String,
			Measures:   interleaveSpot.SpotMeasures.String,
			Completed:  interleaveSpot.Completed,
			PieceID:    interleaveSpot.SpotPieceID.String,
			PieceTitle: interleaveSpot.SpotPieceTitle,
		})
	}

	token := csrf.Token(r)

	hxRequest := htmx.Request(r)
	if hxRequest == nil {
		http.Redirect(w, r, fmt.Sprintf("/library/plans/%s", planID), http.StatusSeeOther)
		return
	}
	pspages.PracticePlanInterleaveSpots(spotInfo, planID, token, allCompleted, false, false).Render(r.Context(), w)
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
	queries := db.New(s.DB)
	plans, err := queries.ListPaginatedPracticePlans(r.Context(), db.ListPaginatedPracticePlansParams{
		UserID: user.ID,
		Limit:  piecesPerPage,
		Offset: int64((pageNum - 1) * piecesPerPage),
	})
	totalPlans, err := queries.CountUserPracticePlans(r.Context(), user.ID)
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
	w.WriteHeader(http.StatusOK)
	s.HxRender(w, r, pspages.PlanList(plans, pageNum, totalPages), "Pieces")
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

	plans, err := queries.ListPaginatedPracticePlans(r.Context(), db.ListPaginatedPracticePlansParams{
		UserID: user.ID,
		Limit:  piecesPerPage,
		Offset: 0,
	})
	totalPlans, err := queries.CountUserPracticePlans(r.Context(), user.ID)
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
	// refresh user from database in case the active plan was deleted
	user, err = queries.GetUserByID(r.Context(), user.ID)
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
	s.HxRender(w, r.WithContext(ctx), pspages.PlanList(plans, 1, totalPages), "Pieces")
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
