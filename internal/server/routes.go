package server

import (
	"net/http"
	"practicebetter/internal/static"
	"time"

	"github.com/a-h/templ"
	"github.com/benbjohnson/hashfs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mavolin/go-htmx"
)

// TODO: stick something on the session after a login redirect so the user gets a notification

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(middleware.Logger)
	r.Use(htmx.NewMiddleware())
	r.Use(csrf.Protect([]byte(s.SecretKey), csrf.Secure(true), csrf.Path("/"), csrf.SameSite(csrf.SameSiteLaxMode)))
	r.Use(middleware.Compress(5, "application/json", "text/html", "text/css", "application/javascript"))
	// r.Use(middleware.SetHeader("Cache-Control", "max-age=5"))
	r.Use(middleware.SetHeader("Vary", "HX-Request"))
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(s.SM.LoadAndSave)
	r.Use(s.ContextPath)

	// staticfiles
	sf := http.NewServeMux()
	// TODO: improve cache control for fonts and such
	sf.Handle("/", http.StripPrefix("/static", hashfs.FileServer(static.HashStatic)))
	r.With(middleware.SetHeader("Access-Control-Allow-Origin", "*")).With(middleware.SetHeader("Cache-Control", "max-age=31536000")).Mount("/static", sf)
	// r.Mount("/static", sf)

	// uploaded files
	uf := http.NewServeMux()
	uf.Handle("/", http.StripPrefix("/uploads", http.FileServer(http.Dir(s.UploadsPath))))
	r.With(middleware.SetHeader("Cache-Control", "max-age=31536000")).Mount("/uploads", uf)

	r.Get("/", s.index)
	r.Get("/about", s.about)
	r.With(s.MaybeUser).With(s.MaybePracticePlan).Route("/practice", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			s.Redirect(w, r, "/practice/random-single")
		})
		r.Get("/random-single", s.randomPractice)
		r.Get("/random-sequence", s.sequencePractice)
		r.Get("/repeat", s.repeatPractice)
		r.Get("/starting-point", s.startingPointPractice)
	})

	r.Route("/auth", func(r chi.Router) {
		r.Get("/login", s.startLogin)
		r.Post("/login", s.continueLogin)
		r.Post("/code", s.completeCodeLogin)
		r.Get("/code", s.forceCodeLogin)
		r.Post("/passkey/login", s.completePasskeySignin)
		r.Get("/logout", s.logoutUserRoute)
		r.Get("/forget", s.forgetUser)
		r.With(s.LoginRequired).Get("/me", s.me)
		r.With(s.LoginRequired).Post("/me", s.updateProfile)
		r.With(s.LoginRequired).Get("/me/edit", s.editProfile)
		r.With(s.LoginRequired).Get("/me/reset", s.getProfile)
		r.With(s.LoginRequired).Post("/passkey/register", s.registerPasskey)
		r.With(s.LoginRequired).Post("/passkey/delete", s.deletePasskeys)
	})

	r.With(s.LoginRequired).With(s.MaybePracticePlan).Route("/library", func(r chi.Router) {
		r.Get("/", s.libraryDashboard)

		r.Get("/pieces", s.pieces)
		r.Post("/pieces", s.createPiece)
		r.Get("/pieces/create", s.createPieceForm)
		r.Get("/pieces/{pieceID}", s.singlePiece)
		r.Get("/pieces/{pieceID}/edit", s.editPiece)
		r.Put("/pieces/{pieceID}", s.updatePiece)
		r.Delete("/pieces/{pieceID}", s.deletePiece)

		r.Get("/pieces/{pieceID}/spots", s.pieceSpots)
		r.Post("/pieces/{pieceID}/spots", s.addSpot)
		r.Get("/pieces/{pieceID}/spots/add", s.addSpotPage)
		r.Get("/pieces/{pieceID}/spots/{spotID}/edit", s.editSpot)
		r.Get("/pieces/{pieceID}/spots/{spotID}", s.singleSpot)
		r.Put("/pieces/{pieceID}/spots/{spotID}", s.updateSpot)
		r.Delete("/pieces/{pieceID}/spots/{spotID}", s.deleteSpot)
		r.Get("/pieces/{pieceID}/spots/{spotID}/practice/repeat", s.repeatPracticeSpot)
		r.Post("/pieces/{pieceID}/spots/{spotID}/practice/repeat", s.repeatPracticeSpotFinished)
		r.Get("/pieces/{pieceID}/spots/{spotID}/reminders/edit", s.getEditRemindersForm)
		r.Get("/pieces/{pieceID}/spots/{spotID}/reminders", s.getReminders)
		r.Post("/pieces/{pieceID}/spots/{spotID}/reminders", s.updateReminders)

		r.Get("/pieces/{pieceID}/practice/random-single", s.piecePracticeRandomSpotsPage)
		r.Post("/pieces/{pieceID}/practice/random-single", s.finishPracticePieceSpots)

		r.Get("/pieces/{pieceID}/practice/starting-point", s.piecePracticeStartingPointPage)
		r.Post("/pieces/{pieceID}/practice/starting-point", s.piecePracticeStartingPointFinished)

		r.Get("/pieces/{pieceID}/practice/repeat", s.piecePracticeRepeatPage)

		r.Get("/upload/audio", s.uploadAudioForm)
		r.Post("/upload/audio", s.uploadAudio)
		r.Get("/upload/images", s.uploadImageForm)
		r.Post("/upload/images", s.uploadImage)

		r.Get("/practice-sessions", s.listPracticeSessions)

		r.Get("/plans/create", s.createPracticePlanForm)
		r.Get("/plans", s.planList)
		r.Post("/plans", s.createPracticePlan)
		r.Get("/plans/{planID}", s.singlePracticePlan)
		r.Delete("/plans/{planID}", s.deletePracticePlan)
		r.Post("/plans/{planID}/resume", s.resumePracticePlan)
		r.Post("/plans/{planID}/stop", s.stopPracticePlan)
		r.Post("/plans/{planID}/interleave-days-spots/complete-all", s.completeInterleaveDaysPlan)
		r.Post("/plans/{planID}/interleave-spots/complete-all", s.completeInterleavePlan)
	})

	return r
}

func (s *Server) HxRender(w http.ResponseWriter, r *http.Request, component templ.Component, title string) {
	hxRequest := htmx.Request(r)
	if hxRequest == nil || hxRequest.Boosted {
		component = Page(s, component, title)
	}
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

func (s *Server) Redirect(w http.ResponseWriter, r *http.Request, url string) {
	if htmx.Request(r) != nil {
		htmx.PushURL(r, url)
		htmx.Redirect(r, url)
		return
	} else {
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
}

type ShowAlertEvent struct {
	Message  string `json:"message"`
	Title    string `json:"title"`
	Variant  string `json:"variant"`
	Duration int    `json:"duration"`
}

type FocusInputEvent struct {
	ID string `json:"id"`
}
