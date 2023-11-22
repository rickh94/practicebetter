package server

import (
	"musiclib/internal/pages"
	"musiclib/internal/static"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/benbjohnson/hashfs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mavolin/go-htmx"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(middleware.Logger)
	r.Use(htmx.NewMiddleware())
	r.Use(csrf.Protect([]byte(s.SecretKey), csrf.Secure(true)))
	r.Use(middleware.Compress(5, "application/json", "text/html", "text/css", "application/javascript"))
	// Just long enough for preload to matter
	r.Use(middleware.SetHeader("Cache-Control", "max-age=300"))
	r.Use(middleware.SetHeader("Vary", "HX-Request"))
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(s.SM.LoadAndSave)

	// staticfiles
	sf := http.NewServeMux()
	sf.Handle("/", http.StripPrefix("/static", hashfs.FileServer(static.HashStatic)))
	r.Mount("/static", sf)

	r.Get("/", s.index)

	return r
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	s.HxRender(w, r, pages.Index())
}

func (s *Server) HxRender(w http.ResponseWriter, r *http.Request, component templ.Component) {
	hxRequest := htmx.Request(r)

	if hxRequest == nil {
		component = Page(s, component)
	}
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

func Redirect(w http.ResponseWriter, r *http.Request, url string) {
	if htmx.Request(r) != nil {
		htmx.PushURL(r, url)
		htmx.Redirect(r, url)
		return
	} else {
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
}

/*

func LoginRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := auth.GetUser(r.Context())
		if err != nil {
			location := r.URL.Path
			location = url.QueryEscape(location)
			Redirect(w, r, "/auth/login?next="+location)
			return
		}
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RedirectLoggedIn(w http.ResponseWriter, r *http.Request, url string) bool {
	if _, err := auth.GetUser(r.Context()); err == nil {
		Redirect(w, r, "/auth/me")
		return true
	}
	return false
}

func RedirectNotLoggedIn(w http.ResponseWriter, r *http.Request, url string) bool {
	if _, err := auth.GetUser(r.Context()); err != nil {
		log.Println(err)
		Redirect(w, r, "/auth/login")
		return true
	}
	return false
}
*/

type ShowAlertEvent struct {
	Message  string `json:"message"`
	Title    string `json:"title"`
	Variant  string `json:"variant"`
	Duration int    `json:"duration"`
}

type FocusInputEvent struct {
	ID string `json:"id"`
}
