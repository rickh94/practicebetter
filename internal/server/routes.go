package server

import (
	"log"
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
	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/img/favicon.ico", http.StatusMovedPermanently)
	})
	r.Get("/about", s.about)
	r.Get("/browserconfig.xml", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<browserconfig>
    <msapplication>
        <tile>
            <square150x150logo src="/static/img/mstile-150x150.png"/>
            <TileColor>#ffefce</TileColor>
        </tile>
    </msapplication>
</browserconfig>
`)); err != nil {
			log.Default().Println(err)
		}
	})
	r.With(s.MaybeUser).With(s.MaybePracticePlan).Route("/practice", s.practiceRouter)

	r.Route("/auth", s.authRouter)

	r.With(s.LoginRequired).With(s.MaybePracticePlan).Route("/library", s.libraryRouter)

	return r
}

func (s *Server) HxRender(w http.ResponseWriter, r *http.Request, mainContent templ.Component, title string) {
	hxRequest := htmx.Request(r)
	if hxRequest == nil || hxRequest.Boosted {
		mainContent = Page(s, mainContent, title)
	}
	w.Header().Set("Content-Type", "text/html")
	if err := mainContent.Render(r.Context(), w); err != nil {
		log.Default().Println(err)
		http.Error(w, "Render Error", http.StatusInternalServerError)
	}
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

func (s *Server) pieceRouter(r chi.Router) {
	r.Get("/", s.pieces)
	r.Post("/", s.createPiece)
	r.Get("/create", s.createPieceForm)
	r.Get("/import", s.importPiece)
	r.Get("/import-file", s.uploadPieceFile)
	r.Post("/import-file", s.importPieceFromFile)
	r.Get("/{pieceID}", s.singlePiece)
	r.Get("/{pieceID}/edit", s.editPiece)
	r.Put("/{pieceID}", s.updatePiece)
	r.Delete("/{pieceID}", s.deletePiece)
	r.Get("/{pieceID}/export.json", s.exportPiece)

	r.Route("/{pieceID}/spots", s.spotsRouter)

	r.Get("/{pieceID}/practice/random-single", s.piecePracticeRandomSpotsPage)
	r.Post("/{pieceID}/practice/random-single", s.finishPracticePieceSpots)

	r.Get("/{pieceID}/practice/random", s.piecePracticeRandomSpotsPage)
	r.Post("/{pieceID}/practice/random", s.finishPracticePieceSpots)

	r.Get("/{pieceID}/practice/starting-point", s.piecePracticeStartingPointPage)
	r.Post("/{pieceID}/practice/starting-point", s.piecePracticeStartingPointFinished)

	r.Get("/{pieceID}/practice/repeat", s.piecePracticeRepeatPage)
}

func (s *Server) spotsRouter(r chi.Router) {
	r.Get("/", s.pieceSpots)
	r.Post("/", s.addSpot)
	r.Get("/add-single", s.addSingleSpotPage)
	r.Get("/add", s.addSpotsFromPDFPage)
	r.Post("/pdf", s.addSpotsFromPDF)

	r.Route("/{spotID}", func(r chi.Router) {
		r.Get("/", s.singleSpot)
		r.Get("/edit", s.editSpot)
		r.Put("/", s.updateSpot)
		r.Patch("/", s.updatePartialSpot)

		r.Patch("/image", s.updateSpotImage)
		r.Patch("/audio", s.updateSpotAudio)
		r.Patch("/reminders", s.updateReminders)

		r.Delete("/", s.deleteSpot)
		r.Get("/practice/repeat", s.repeatPracticeSpot)
		r.Post("/practice/repeat", s.repeatPracticeSpotFinished)
		r.Get("/practice/display", s.getPracticeSpotDisplay)
	})
}

func (s *Server) planRouter(r chi.Router) {
	r.Get("/", s.planList)
	r.Post("/", s.createPracticePlan)
	r.Get("/create", s.createPracticePlanForm)

	r.Route("/{planID}", func(r chi.Router) {
		r.Get("/", s.singlePracticePlan)
		r.Delete("/", s.deletePracticePlan)

		r.Get("/next", s.redirectToNextPlanItem)
		r.Get("/interleave", s.getInterleaveList)
		r.Get("/interleave/start", s.startInterleavePracticing)
		r.Get("/interleave/{spotID}", s.interleavePracticeSpot)
		r.Post("/interleave/practice", s.saveInterleaveResult)
		r.Get("/infrequent/start", s.startInfrequentPracticing)
		r.Get("/interleave_days/{spotID}", s.infrequentPracticeSpot)
		r.Post("/infrequent/practice", s.saveInfrequentResult)

		r.Post("/resume", s.resumePracticePlan)
		r.Post("/stop", s.stopPracticePlan)
		r.Post("/duplicate", s.duplicatePracticePlan)
		r.Get("/edit", s.editPracticePlan)

		r.Delete("/spots/{practiceType}/{spotID}", s.deleteSpotFromPracticePlan)
		r.Delete("/pieces/{practiceType}/{pieceID}", s.deletePieceFromPracticePlan)
		r.Delete("/scales/{userScaleID}", s.deleteScaleFromPracticePlan)
		r.Get("/spots/{practiceType}/add", s.getSpotsForPracticePlan)
		r.Get("/spots/new/add/pieces", s.getNewSpotPiecesForPracticePlan)
		r.Get("/spots/new/add/pieces/{pieceID}", s.getNewSpotsForPracticePlan)
		r.Put("/spots/{practiceType}", s.addSpotsToPracticePlan)
		r.Get("/scales/add", s.getScalesForPracticePlan)
		r.Post("/scales/add", s.addScaleToPracticePlan)

		r.Get("/pieces/{practiceType}/add", s.getPiecesForPracticePlan)
		r.Put("/pieces/{practiceType}", s.addPiecesToPracticePlan)
	})
}

func (s *Server) scalesRouter(r chi.Router) {
	r.Get("/", s.scalePicker)
	r.Get("/{scaleID}", s.singleScale)
	r.Put("/{scaleID}", s.updateScale)
	r.Get("/{scaleID}/practice", s.getPracticeScale)
	r.Post("/{scaleID}/practice", s.practiceScale)
	r.Get("/{scaleID}/edit", s.editScale)
	r.Get("/autocreate", s.autocreateScale)
}

func (s *Server) libraryRouter(r chi.Router) {
	r.Get("/", s.libraryDashboard)

	r.Route("/pieces", s.pieceRouter)
	r.Route("/scales", s.scalesRouter)

	r.Get("/upload/audio", s.uploadAudioForm)
	r.Post("/upload/audio", s.uploadAudio)
	r.Get("/upload/images", s.uploadImageForm)
	r.Post("/upload/images", s.uploadImage)

	r.Get("/break", s.shouldRecommendBreak)
	r.Post("/break", s.takeABreak)
	r.Get("/break/last", s.lastBreak)

	r.Route("/plans", s.planRouter)

}

func (s *Server) practiceRouter(r chi.Router) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		s.Redirect(w, r, "/practice/random-single")
	})
	r.Get("/random-single", s.randomPractice)
	r.Get("/random-sequence", s.sequencePractice)
	r.Get("/repeat", s.repeatPractice)
	r.Get("/starting-point", s.startingPointPractice)
}

func (s *Server) authRouter(r chi.Router) {
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
}
