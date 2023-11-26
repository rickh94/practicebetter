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
	r.Use(csrf.Protect([]byte(s.SecretKey), csrf.Secure(true)))
	r.Use(middleware.Compress(5, "application/json", "text/html", "text/css", "application/javascript"))
	// Just long enough for preload to matter
	r.Use(middleware.SetHeader("Cache-Control", "max-age=3600"))
	r.Use(middleware.SetHeader("Vary", "HX-Request"))
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(s.SM.LoadAndSave)

	// staticfiles
	sf := http.NewServeMux()
	sf.Handle("/", http.StripPrefix("/static", hashfs.FileServer(static.HashStatic)))
	r.With(middleware.SetHeader("Cache-Control", "max-age=604800")).With(middleware.SetHeader("Access-Control-Allow-Origin", "*")).Mount("/static", sf)
	// r.Mount("/static", sf)

	r.Get("/", s.index)
	r.Get("/about", s.about)
	r.Route("/practice", func(r chi.Router) {
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
		r.With(s.LoginRequired).Get("/me", s.me)
		r.With(s.LoginRequired).Post("/me", s.updateProfile)
		r.With(s.LoginRequired).Get("/me/edit", s.editProfile)
		r.With(s.LoginRequired).Get("/me/reset", s.getProfile)
		r.With(s.LoginRequired).Post("/passkey/register", s.registerPasskey)
		r.With(s.LoginRequired).Post("/passkey/delete", s.deletePasskeys)
	})

	r.With(s.LoginRequired).Route("/library", func(r chi.Router) {
		r.Get("/", s.libraryDashboard)
		r.Get("/pieces/create", s.createPieceForm)
		r.Post("/pieces/create", s.createPiece)
		/*
			r.Get("/random-single", s.randomPractice)
			r.Get("/random-sequence", s.sequencePractice)
			r.Get("/repeat", s.repeatPractice)
			r.Get("/starting-point", s.startingPointPractice)
		*/
	})

	return r
}

func (s *Server) HxRender(w http.ResponseWriter, r *http.Request, component templ.Component) {
	hxRequest := htmx.Request(r)
	if hxRequest == nil {
		component = Page(s, component)
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
