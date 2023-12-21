package server

import (
	"database/sql"
	"encoding/json"
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

func (s *Server) listPracticeSessions(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}
	pageNum, err := strconv.Atoi(page)
	log.Default().Println(pageNum)

	queries := db.New(s.DB)
	ps, err := queries.ListPracticeSessions(r.Context(), db.ListPracticeSessionsParams{
		UserID: user.ID,
		Page:   pageNum,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	hasNext, err := queries.HasMorePracticeSessions(r.Context(), db.HasMorePracticeSessionsParams{
		UserID: user.ID,
		Page:   pageNum,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	psData, err := json.Marshal(&ps)
	s.HxRender(w, r, pspages.PSList(s, string(psData), pageNum, hasNext), "Practice Sessions")
}

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
				canRandomSpotsPractice = true
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
				canRandomSpotsPractice = true
				// if the spot doesn't have a last practiced date, or if it wasn't practiced yesterday (roughly)
				if !row.SpotLastPracticed.Valid || row.SpotLastPracticed.Int64 < time.Now().Unix()-60*60*25 {
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
	s.SetActivePracticePlanID(r.Context(), newPlan.ID)
	htmx.PushURL(r, "/library/plans/"+newPlan.ID)
	s.renderPracticePlanPage(w, r, newPlan.ID, user.ID)
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
	planPieces, err := queries.GetPracticePlanWithPieces(r.Context(), db.GetPracticePlanWithPiecesParams{
		ID:     planID,
		UserID: userID,
	})
	planSpots, err := queries.GetPracticePlanWithSpots(r.Context(), db.GetPracticePlanWithSpotsParams{
		ID:     planID,
		UserID: userID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	activePracticePlanID, _ := s.GetActivePracticePlanID(r.Context())

	var planData pspages.PracticePlanData
	planData.ID = planPieces[0].ID
	planData.Date = planPieces[0].Date
	planData.Completed = planPieces[0].Completed
	planData.InterleaveDaysSpotsCompleted = true
	planData.InterleaveSpotsCompleted = true
	planData.IsActive = planID == activePracticePlanID

	for _, row := range planPieces {
		if row.PieceID.Valid {
			var piece pspages.PracticePlanPiece
			piece.ID = row.PieceID.String
			piece.Title = row.PieceTitle.String
			piece.ActiveSpots = row.PieceActiveSpots
			piece.CompletedSpots = row.PieceCompletedSpots
			if row.PieceComposer.Valid {
				piece.Composer = row.PieceComposer.String
			} else {
				piece.Composer = "Unknown"
			}
			if !row.PiecePracticeType.Valid {
				log.Default().Println("Got a piece without a practice type, shouldn't happen")
				continue
			}
			if row.PieceCompleted.Valid {
				piece.Completed = row.PieceCompleted.Bool
			} else {
				piece.Completed = false
			}

			if row.PiecePracticeType.String == "random_spots" {
				planData.RandomSpotsPieces = append(planData.RandomSpotsPieces, piece)
			} else if row.PiecePracticeType.String == "starting_point" {
				planData.RandomStartPieces = append(planData.RandomStartPieces, piece)
			}

		}
	}
	for _, row := range planSpots {
		if row.SpotID.Valid {
			var spot pspages.PracticePlanSpot
			spot.ID = row.SpotID.String
			spot.Name = row.SpotName.String
			if row.SpotMeasures.Valid {
				spot.Measures = row.SpotMeasures.String
			} else {
				spot.Measures = ""
			}
			spot.PieceTitle = row.SpotPieceTitle
			spot.PieceID = row.SpotPieceID.String
			spot.Completed = row.SpotCompleted.Bool

			if !row.SpotPracticeType.Valid {
				log.Default().Println("Got a spot without a practice type, shouldn't happen")
				continue
			}

			if row.SpotPracticeType.String == "interleave" {
				planData.InterleaveSpots = append(planData.InterleaveSpots, spot)
				if !row.SpotCompleted.Bool {
					planData.InterleaveSpotsCompleted = false
				}
			} else if row.SpotPracticeType.String == "interleave_days" {
				planData.InterleaveDaysSpots = append(planData.InterleaveDaysSpots, spot)
				if !row.SpotCompleted.Bool {
					planData.InterleaveDaysSpotsCompleted = false
				}
			} else if row.SpotPracticeType.String == "extra_repeat" {
				planData.ExtraRepeatSpots = append(planData.ExtraRepeatSpots, spot)
			} else if row.SpotPracticeType.String == "new" {
				planData.NewSpots = append(planData.NewSpots, spot)
			}
		}
	}

	token := csrf.Token(r)
	s.HxRender(w, r, pspages.PracticePlanPage(planData, token), "Practice Plan")
}

func (s *Server) completeInterleaveDaysPlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	planID := chi.URLParam(r, "planID")
	activePlanID, err := s.GetActivePracticePlanID(r.Context())
	if err != nil || planID != activePlanID {
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

	planSpots, err := qtx.GetPracticePlanInterleaveDaysSpots(r.Context(), db.GetPracticePlanInterleaveDaysSpotsParams{
		PlanID: planID,
		UserID: user.ID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	spotInfo := make([]pspages.PracticePlanSpot, 0, len(planSpots))
	for _, planSpot := range planSpots {
		if err := qtx.CompletePracticePlanSpot(r.Context(), db.CompletePracticePlanSpotParams{
			SpotID: planSpot.SpotID,
			UserID: user.ID,
			PlanID: planID,
		}); err != nil {
			log.Default().Println(err)
			http.Error(w, "Could not update spot", http.StatusInternalServerError)
			return
		}
		if err := qtx.UpdateSpotPracticed(r.Context(), db.UpdateSpotPracticedParams{
			SpotID: planSpot.SpotID,
			UserID: user.ID,
		}); err != nil {
			log.Default().Println(err)
			http.Error(w, "Could not update spot", http.StatusInternalServerError)
			return
		}

		if err := qtx.PracticeSpot(r.Context(), db.PracticeSpotParams{
			UserID:            user.ID,
			SpotID:            planSpot.SpotID,
			PracticeSessionID: practiceSessionID,
		}); err != nil {
			log.Default().Println(err)
			http.Error(w, "Could not practice spot", http.StatusInternalServerError)
			return
		}
		spotInfo = append(spotInfo, pspages.PracticePlanSpot{
			ID:         planSpot.SpotID,
			Name:       planSpot.SpotName.String,
			Measures:   planSpot.SpotMeasures.String,
			Completed:  true,
			PieceID:    planSpot.SpotPieceID.String,
			PieceTitle: planSpot.SpotPieceTitle,
		})
	}

	if err := tx.Commit(); err != nil {
		log.Default().Println(err)
		http.Error(w, "Could not commit practice session", http.StatusInternalServerError)
		return
	}

	token := csrf.Token(r)
	htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "You've completed your interleaved days spots!",
		Title:    "Completed!",
		Variant:  "success",
		Duration: 3000,
	})
	pspages.PracticePlanInterleaveDaysSpots(spotInfo, planID, token, true).Render(r.Context(), w)
}

func (s *Server) completeInterleavePlan(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	planID := chi.URLParam(r, "planID")
	activePlanID, err := s.GetActivePracticePlanID(r.Context())
	if err != nil || planID != activePlanID {
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

	planSpots, err := qtx.GetPracticePlanInterleaveSpots(r.Context(), db.GetPracticePlanInterleaveSpotsParams{
		PlanID: planID,
		UserID: user.ID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	spotInfo := make([]pspages.PracticePlanSpot, 0, len(planSpots))
	for _, planSpot := range planSpots {
		if err := qtx.CompletePracticePlanSpot(r.Context(), db.CompletePracticePlanSpotParams{
			SpotID: planSpot.SpotID,
			UserID: user.ID,
			PlanID: planID,
		}); err != nil {
			log.Default().Println(err)
			http.Error(w, "Could not update spot", http.StatusInternalServerError)
			return
		}
		if err := qtx.UpdateSpotPracticed(r.Context(), db.UpdateSpotPracticedParams{
			SpotID: planSpot.SpotID,
			UserID: user.ID,
		}); err != nil {
			log.Default().Println(err)
			http.Error(w, "Could not update spot", http.StatusInternalServerError)
			return
		}

		if err := qtx.PracticeSpot(r.Context(), db.PracticeSpotParams{
			UserID:            user.ID,
			SpotID:            planSpot.SpotID,
			PracticeSessionID: practiceSessionID,
		}); err != nil {
			log.Default().Println(err)
			http.Error(w, "Could not practice spot", http.StatusInternalServerError)
			return
		}
		spotInfo = append(spotInfo, pspages.PracticePlanSpot{
			ID:         planSpot.SpotID,
			Name:       planSpot.SpotName.String,
			Measures:   planSpot.SpotMeasures.String,
			Completed:  true,
			PieceID:    planSpot.SpotPieceID.String,
			PieceTitle: planSpot.SpotPieceTitle,
		})
	}

	if err := tx.Commit(); err != nil {
		log.Default().Println(err)
		http.Error(w, "Could not commit practice session", http.StatusInternalServerError)
		return
	}

	token := csrf.Token(r)
	htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "You've completed your interleaved spots!",
		Title:    "Completed!",
		Variant:  "success",
		Duration: 3000,
	})
	pspages.PracticePlanInterleaveDaysSpots(spotInfo, planID, token, true).Render(r.Context(), w)
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
