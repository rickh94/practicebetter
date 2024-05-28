package server

import (
	"database/sql"
	"log"
	"math"
	"net/http"
	"practicebetter/internal/ck"
	"practicebetter/internal/config"
	"practicebetter/internal/db"
	"practicebetter/internal/pages/readingpages"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/mavolin/go-htmx"
	"github.com/nrednav/cuid2"
)

func (s *Server) createSightReadingForm(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)
	s.HxRender(w, r, readingpages.CreateSightReadingPage(s, token), "New Sight Reading")
}

func (s *Server) createSightReading(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	queries := db.New(s.DB)

	if err := r.ParseForm(); err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Variant:  "error",
			Title:    "Invalid Data",
			Message:  "Your submission was invalid",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	title := r.Form.Get("title")
	composer := r.Form.Get("composer")
	info := r.Form.Get("info")

	itemID := cuid2.Generate()

	readingItem, err := queries.CreateSightReadingItem(r.Context(), db.CreateSightReadingItemParams{
		ID:       itemID,
		Title:    title,
		Info:     sql.NullString{String: info, Valid: info != ""},
		Composer: sql.NullString{String: composer, Valid: composer != ""},
		UserID:   user.ID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Failed to create reading item")
		return
	}

	token := csrf.Token(r)

	hxRequest := htmx.Request(r)
	if hxRequest == nil {
		s.Redirect(w, r, "/library/reading/"+itemID)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	htmx.PushURL(r, "/library/reading/"+itemID)
	if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Successfully added reading Item: " + readingItem.Title + ".",
		Title:    "Sight Reading Created",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}

	itemInfo := readingpages.SingleReadingItemInfo{
		ID:       itemID,
		Title:    readingItem.Title,
		Info:     readingItem.Info,
		Composer: readingItem.Composer,
	}

	w.WriteHeader(http.StatusCreated)
	s.HxRender(w, r, readingpages.SingleReadingItem(s, itemInfo, token), readingItem.Title)
}

func (s *Server) bulkCreateSightReading(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	queries := db.New(s.DB)

	if err := r.ParseForm(); err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Variant:  "error",
			Title:    "Invalid Data",
			Message:  "Your submission was invalid",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bulk := r.Form.Get("bulk")
	if bulk == "" {
		s.InvalidInputError(w, r, "You need at least one item")
		return
	}
	bulkItems := strings.Split(bulk, "\n")
	if len(bulkItems) < 1 {
		s.InvalidInputError(w, r, "You need at least one item")
		return
	}

	numCreated := 0
	for _, item := range bulkItems {
		title := strings.TrimSpace(item)
		if title == "" {
			continue
		}
		itemID := cuid2.Generate()

		_, err := queries.CreateSightReadingItem(r.Context(), db.CreateSightReadingItemParams{
			ID:       itemID,
			Title:    title,
			Info:     sql.NullString{String: "", Valid: false},
			Composer: sql.NullString{String: "", Valid: false},
			UserID:   user.ID,
		})
		if err != nil {
			log.Default().Println("Failed to create item")
			log.Default().Println(err)
			continue
		}
		numCreated++
	}

	hxRequest := htmx.Request(r)
	if hxRequest == nil {
		s.Redirect(w, r, "/library/reading")
		return
	}

	w.Header().Set("Content-Type", "text/html")
	htmx.PushURL(r, "/library/reading")
	if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
		Message:  "Successfully added " + strconv.Itoa(numCreated) + " reading items.",
		Title:    "Sight Reading Created",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}

	items, err := queries.ListPaginatedUserReadingItems(r.Context(), db.ListPaginatedUserReadingItemsParams{
		UserID: user.ID,
		Limit:  config.ItemsPerPage,
		Offset: 0,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not load sight reading items")
		return
	}
	totalPieces, err := queries.CountUserReadingItems(r.Context(), user.ID)
	if err != nil {
		log.Default().Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	totalPages := int(math.Ceil(float64(totalPieces) / float64(config.ItemsPerPage)))
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	s.HxRender(w, r, readingpages.ReadingList(items, 1, totalPages), "Sight Reading")
}

func (s *Server) sightReadingItems(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}
	pageNum, err := strconv.Atoi(page)
	if err != nil {
		pageNum = 1
	}
	queries := db.New(s.DB)
	items, err := queries.ListPaginatedUserReadingItems(r.Context(), db.ListPaginatedUserReadingItemsParams{
		UserID: user.ID,
		Limit:  config.ItemsPerPage,
		Offset: int64((pageNum - 1) * config.ItemsPerPage),
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not load sight reading items")
		return
	}
	totalPieces, err := queries.CountUserReadingItems(r.Context(), user.ID)
	if err != nil {
		log.Default().Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	totalPages := int(math.Ceil(float64(totalPieces) / float64(config.ItemsPerPage)))
	if err != nil {
		log.Default().Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	s.HxRender(w, r, readingpages.ReadingList(items, pageNum, totalPages), "Sight Reading")
}

func (s *Server) singleSightReadingItem(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemID")
	user := r.Context().Value(ck.UserKey).(db.User)

	queries := db.New(s.DB)
	item, err := queries.GetReadingByID(r.Context(), db.GetReadingByIDParams{
		ReadingID: itemID,
		UserID:    user.ID,
	})
	if err != nil {
		s.DatabaseError(w, r, err, "Could not load sight reading item")
		return
	}

	itemInfo := readingpages.SingleReadingItemInfo{
		ID:       itemID,
		Title:    item.Title,
		Info:     item.Info,
		Composer: item.Composer,
	}

	token := csrf.Token(r)

	w.WriteHeader(http.StatusCreated)
	s.HxRender(w, r, readingpages.SingleReadingItem(s, itemInfo, token), item.Title)
}
