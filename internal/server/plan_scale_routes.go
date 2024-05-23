package server

import (
	"log"
	"net/http"
	"practicebetter/internal/ck"
	"practicebetter/internal/components"
	"practicebetter/internal/db"
	"practicebetter/internal/pages/ewspages"
	"practicebetter/internal/pages/planpages"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/mavolin/go-htmx"
	"github.com/nrednav/cuid2"
)

func (s *Server) deleteScaleFromPracticePlan(w http.ResponseWriter, r *http.Request) {
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
	userScaleID := chi.URLParam(r, "userScaleID")
	queries := db.New(s.DB)
	_, err := queries.DeletePracticePlanScale(r.Context(), db.DeletePracticePlanScaleParams{
		PlanID:      planID,
		UserID:      user.ID,
		UserScaleID: userScaleID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not remove scale from practice plan")
		return
	}
	if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Your scale has been removed from this practice plan.",
		Title:    "Removed",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}
	w.WriteHeader(http.StatusOK)

}

func (s *Server) getScalesForPracticePlan(w http.ResponseWriter, r *http.Request) {
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
	token := csrf.Token(r)

	component := ewspages.PlanScalePicker(props, planID, allModes, token)
	if err := component.Render(r.Context(), w); err != nil {
		log.Default().Println(err)
	}
}

func (s *Server) addScaleToPracticePlan(w http.ResponseWriter, r *http.Request) {
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

	scaleID, err := strconv.ParseInt(r.FormValue("scale"), 10, 64)
	if err != nil {
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Invalid Scale ID",
			Title:    "Invalid Input",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Invalid Form", http.StatusBadRequest)
		return
	}

	userScaleID, err := queries.CheckForUserScale(r.Context(), db.CheckForUserScaleParams{
		UserID:  user.ID,
		ScaleID: scaleID,
	})
	if err != nil {
		log.Default().Println(err)
		userScaleID = cuid2.Generate()
		_, err := queries.CreateUserScale(r.Context(), db.CreateUserScaleParams{
			ID:            userScaleID,
			UserID:        user.ID,
			ScaleID:       scaleID,
			PracticeNotes: "",
			Reference:     "",
		})
		if err != nil {
			s.DatabaseError(w, r, err, "Failed to create user scale")
			return
		}
	}

	idxQuery, err := queries.GetMaxScaleIdx(r.Context(), db.GetMaxScaleIdxParams{
		PlanID: planID,
		UserID: user.ID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Failed to load scales")
		return
	}
	idx, ok := idxQuery.(int64)
	if !ok {
		idx = 0
	}

	_, err = queries.CreatePracticePlanScaleWithIdx(r.Context(), db.CreatePracticePlanScaleWithIdxParams{
		PracticePlanID: planID,
		UserScaleID:    userScaleID,
		Idx:            idx + 1,
	})

	if err != nil {
		s.DatabaseError(w, r, err, "Failed to add scale to plan")
		return
	}
	if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Scale Added",
		Title:    "Success",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}

	scale, err := queries.GetUserScale(r.Context(), db.GetUserScaleParams{
		ID:     userScaleID,
		UserID: user.ID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Failed to load scale")
		return
	}

	scaleInfo := components.UserScaleInfo{
		UserScaleID:   userScaleID,
		KeyName:       scale.KeyName,
		ModeName:      scale.Mode,
		PracticeNotes: scale.PracticeNotes,
		Reference:     scale.Reference,
	}

	token := csrf.Token(r)
	if err := planpages.DeleteScaleCard(scaleInfo, false, token, activePlanID).Render(r.Context(), w); err != nil {
		log.Default().Println(err)
	}

}
