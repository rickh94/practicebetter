package server

import (
	"log"
	"net/http"
	"practicebetter/internal/ck"
	"practicebetter/internal/components"
	"practicebetter/internal/db"
	"practicebetter/internal/pages/ewspages"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/mavolin/go-htmx"
	"github.com/nrednav/cuid2"
)

// func (s *Server) ewsDashboard(w http.ResponseWriter, r *http.Request) {
// 	// queries := db.New(s.DB)
// 	// user := r.Context().Value(ck.UserKey).(db.User)
//
// 	s.HxRender(w, r, ewspages.Dashboard(), "EWS Dashboard")
// }

func (s *Server) scalePicker(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	queries := db.New(s.DB)
	userScales, err := queries.ListUserScales(r.Context(), user.ID)
	if err != nil {
		s.DatabaseError(w, r, err, "Failed to load user scales")
		return
	}
	userScaleLookup := make(map[int64]db.UserScale)
	for _, scale := range userScales {
		userScaleLookup[scale.ScaleID] = scale
	}

	modeID, err := strconv.ParseInt(r.URL.Query().Get("mode"), 10, 64)

	if err != nil || modeID == 0 {
		modeID = 1
	}

	selectedMode, err := queries.GetMode(r.Context(), modeID)
	if err != nil {
		log.Default().Println(err)
		selectedMode, err = queries.GetMode(r.Context(), 1)
		if err != nil {
			s.DatabaseError(w, r, err, "Failed to load mode")
			return
		}

	}

	scales, err := queries.ListScalesForMode(r.Context(), modeID)
	if err != nil {
		s.DatabaseError(w, r, err, "Failed to load scales")
		return

	}

	basicOnly, err := strconv.ParseBool(r.URL.Query().Get("basicOnly"))
	if err != nil {
		basicOnly = true
	}

	allModes, err := queries.ListModes(r.Context(), basicOnly)
	if err != nil {
		s.DatabaseError(w, r, err, "Failed to load modes")
		return
	}

	props := ewspages.ScalePickerProps{
		UserScales:   userScaleLookup,
		SelectedMode: selectedMode,
		Scales:       scales,
		Csrf:         csrf.Token(r),
	}

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Type", "text/html")
	hxRequest := htmx.Request(r)
	if hxRequest == nil || hxRequest.Boosted {
		page := Page(s, ewspages.ScalePickerPage(s, props, allModes), "Scale Picker")
		if err := page.Render(r.Context(), w); err != nil {
			log.Default().Println(err)
		}
	} else if hxRequest.Target == "main-content" {
		component := ewspages.ScalePickerPage(s, props, allModes)
		if err := component.Render(r.Context(), w); err != nil {
			log.Default().Println(err)
		}
	} else if hxRequest.Target == "scale-picker" {
		component := ewspages.ScalePicker(props)
		if err := component.Render(r.Context(), w); err != nil {
			log.Default().Println(err)
		}
	}
}

func (s *Server) singleScale(w http.ResponseWriter, r *http.Request) {
	scaleID := chi.URLParam(r, "scaleID")
	user := r.Context().Value(ck.UserKey).(db.User)

	queries := db.New(s.DB)

	scale, err := queries.GetUserScale(r.Context(), db.GetUserScaleParams{
		UserID: user.ID,
		ID:     scaleID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Failed to load scale")
		return
	}
	s.HxRender(w, r, ewspages.SingleScale(s, scale), scale.KeyName+" "+scale.Mode+" Scale")
}

func (s *Server) autocreateScale(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	scaleID, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		log.Default().Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	queries := db.New(s.DB)

	var userScaleID string
	userScaleID, err = queries.CheckForUserScale(r.Context(), db.CheckForUserScaleParams{
		UserID:  user.ID,
		ScaleID: scaleID,
	})
	if err != nil || userScaleID == "" {
		userScale, err := queries.CreateUserScale(r.Context(), db.CreateUserScaleParams{
			ID:            cuid2.Generate(),
			UserID:        user.ID,
			ScaleID:       scaleID,
			PracticeNotes: "",
			Reference:     "",
		})
		if err != nil {
			s.DatabaseError(w, r, err, "Failed to create user scale")
			return
		}
		userScaleID = userScale.ID
	}

	hxRequest := htmx.Request(r)
	if hxRequest == nil || hxRequest.Boosted {
		http.Redirect(w, r, "/library/scales/"+userScaleID, http.StatusSeeOther)
	} else {
		htmx.PushURL(r, "/library/scales/"+userScaleID)

		scale, err := queries.GetUserScale(r.Context(), db.GetUserScaleParams{
			UserID: user.ID,
			ID:     userScaleID,
		})
		if err != nil {
			s.DatabaseError(w, r, err, "Failed to load scale")
			return
		}
		s.HxRender(w, r, ewspages.SingleScale(s, scale), scale.KeyName+" "+scale.Mode+" Scale")
	}

}

func (s *Server) getPracticeScale(w http.ResponseWriter, r *http.Request) {
	scaleID := chi.URLParam(r, "scaleID")
	user := r.Context().Value(ck.UserKey).(db.User)

	queries := db.New(s.DB)

	scale, err := queries.GetUserScale(r.Context(), db.GetUserScaleParams{
		UserID: user.ID,
		ID:     scaleID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Failed to load scale")
		return
	}
	token := csrf.Token(r)
	if err := ewspages.PracticeScaleDisplay(scale, token).Render(r.Context(), w); err != nil {
		log.Default().Println(err)
	}
}

func (s *Server) practiceScale(w http.ResponseWriter, r *http.Request) {
	scaleID := chi.URLParam(r, "scaleID")
	user := r.Context().Value(ck.UserKey).(db.User)

	// FIXME: update last practiced
	queries := db.New(s.DB)

	activePracticePlanID, ok := s.GetActivePracticePlanID(r.Context())
	if ok && activePracticePlanID != "" {
		if err := queries.CompletePracticePlanScale(r.Context(), db.CompletePracticePlanScaleParams{
			PlanID:      activePracticePlanID,
			UserID:      user.ID,
			UserScaleID: scaleID,
		}); err != nil {
			s.DatabaseError(w, r, err, "Could not complete practice plan scale")
			return
		}

		if err := queries.UpdatePlanLastPracticed(r.Context(), db.UpdatePlanLastPracticedParams{
			ID:     activePracticePlanID,
			UserID: user.ID,
		}); err != nil {
			s.DatabaseError(w, r, err, "Could not update plan last practiced")
			return
		}

	}

	_, err := queries.UpdateScalePracticed(r.Context(), db.UpdateScalePracticedParams{
		ID:     scaleID,
		UserID: user.ID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Failed to load scale")
		return
	}
	if activePracticePlanID != "" {
		scale, err := queries.GetUserScale(r.Context(), db.GetUserScaleParams{
			UserID: user.ID,
			ID:     scaleID,
		})
		if err != nil {
			s.DatabaseError(w, r, err, "Failed to load scale")
			return
		}
		scaleInfo := components.UserScaleInfo{
			UserScaleID:   scaleID,
			KeyName:       scale.KeyName,
			ModeName:      scale.Mode,
			PracticeNotes: scale.PracticeNotes,
			Reference:     scale.Reference,
		}
		if err := components.ScaleCardOOB(scaleInfo, true, true).Render(r.Context(), w); err != nil {
			log.Default().Println(err)
		}
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) editScale(w http.ResponseWriter, r *http.Request) {
	scaleID := chi.URLParam(r, "scaleID")
	user := r.Context().Value(ck.UserKey).(db.User)

	queries := db.New(s.DB)

	scale, err := queries.GetUserScale(r.Context(), db.GetUserScaleParams{
		UserID: user.ID,
		ID:     scaleID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Failed to load scale")
		return
	}
	token := csrf.Token(r)
	if err := ewspages.EditScaleDisplay(scale, token).Render(r.Context(), w); err != nil {
		log.Default().Println(err)
	}
}

func (s *Server) updateScale(w http.ResponseWriter, r *http.Request) {
	scaleID := chi.URLParam(r, "scaleID")
	user := r.Context().Value(ck.UserKey).(db.User)

	if err := r.ParseForm(); err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not parse form",
			Title:    "Form Error",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	queries := db.New(s.DB)

	_, err := queries.UpdateUserScale(r.Context(), db.UpdateUserScaleParams{
		PracticeNotes: r.Form.Get("practice-notes"),
		Reference:     r.Form.Get("reference"),
		Working:       r.Form.Get("working") == "on",
		ID:            scaleID,
		UserID:        user.ID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Failed to update scale")
		return
	}

	scale, err := queries.GetUserScale(r.Context(), db.GetUserScaleParams{
		UserID: user.ID,
		ID:     scaleID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Failed to load scale")
		return
	}
	token := csrf.Token(r)
	if err := ewspages.UpdatedScale(scale, token).Render(r.Context(), w); err != nil {
		log.Default().Println(err)
	}
}
