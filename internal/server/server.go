package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"practicebetter/internal/db"
	"practicebetter/internal/static"
	"strconv"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gomodule/redigo/redis"
	mail "github.com/xhit/go-simple-mail/v2"
)

var port = 8080

type Server struct {
	port           int
	DB             *sql.DB
	SM             *scs.SessionManager
	WebAuthn       *webauthn.WebAuthn
	EmailSender    *mail.SMTPServer
	EmailFrom      string
	SecretKey      string
	StaticHostname string
	UploadsPath    string
	Debug          bool
}

func getEnvOrPanic(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("Missing environment variable: " + key)
	}
	return value
}

func NewServer() *http.Server {

	// SETUP DATABASE
	dbPath := getEnvOrPanic("DB_PATH")
	dbPath = fmt.Sprintf("file:%s?_fk=1&_journal=WAL&_mode=rw", dbPath)
	log.Printf("connecting to %s", dbPath)

	pool, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}

	pool.SetConnMaxLifetime(0)
	pool.SetMaxOpenConns(4)
	pool.SetMaxIdleConns(4)

	// SETUP SESSIONS
	redisUri := getEnvOrPanic("REDIS_URI")
	redisPool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", redisUri)
		},
	}

	sm := scs.New()
	sm.Lifetime = 24 * time.Hour
	sm.Store = redisstore.New(redisPool)

	// SETUP WEBAUTHN
	hostname := getEnvOrPanic("HOSTNAME")
	wconfig := &webauthn.Config{
		RPDisplayName: getEnvOrPanic("DISPLAY_NAME"),
		RPID:          hostname,
		RPOrigins:     []string{fmt.Sprintf("https://%s", hostname)},
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			RequireResidentKey: protocol.ResidentKeyNotRequired(),
			UserVerification:   protocol.VerificationRequired,
		},
		Timeouts: webauthn.TimeoutsConfig{
			Login: webauthn.TimeoutConfig{
				Enforce:    true,                 // Require the response from the client comes before the end of the timeout.
				Timeout:    time.Second * 60 * 5, // Standard timeout for login sessions.
				TimeoutUVD: time.Second * 60 * 5, // Timeout for login sessions which have user verification set to discouraged.
			},
			Registration: webauthn.TimeoutConfig{
				Enforce:    true,                  // Require the response from the client comes before the end of the timeout.
				Timeout:    time.Second * 60 * 10, // Standard timeout for registration sessions.
				TimeoutUVD: time.Second * 60 * 10, // Timeout for login sessions which have user verification set to discouraged.
			},
		},
	}

	wm, err := webauthn.New(wconfig)
	if err != nil {
		panic(err)
	}

	// SETUP EMAIL
	es := mail.NewSMTPClient()
	es.Host = getEnvOrPanic("EMAIL_HOST")
	es.Port, err = strconv.Atoi(getEnvOrPanic("EMAIL_PORT"))
	if err != nil {
		panic(err)
	}
	es.Username = getEnvOrPanic("EMAIL_USERNAME")
	es.Password = getEnvOrPanic("EMAIL_PASSWORD")
	es.Encryption = mail.EncryptionSTARTTLS

	debug := false
	if debugValue := os.Getenv("DEBUG"); debugValue != "" {
		debug = true
	}

	NewServer := &Server{
		port:           port,
		DB:             pool,
		SM:             sm,
		WebAuthn:       wm,
		EmailSender:    es,
		EmailFrom:      getEnvOrPanic("EMAIL_FROM"),
		SecretKey:      getEnvOrPanic("SECRET_KEY"),
		StaticHostname: os.Getenv("STATIC_HOSTNAME"),
		UploadsPath:    getEnvOrPanic("UPLOADS_PATH"),
		Debug:          debug,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}

func (s *Server) SendEmail(to, subject, body string) error {

	email := mail.NewMSG()

	email.SetFrom(s.EmailFrom).
		AddTo(to).
		SetSubject(subject)

	email.SetBody(mail.TextPlain, body)

	client, err := s.EmailSender.Connect()
	if err != nil {
		log.Default().Printf("failed to connect to email server: %v\n", err)
		return err
	}
	err = email.Send(client)
	if err != nil {
		log.Default().Printf("failed to send email: %v\n", err)
		return err
	}
	return nil
}

func (s *Server) StaticUrl(name string) string {
	return s.StaticHostname + "/static/" + static.HashStatic.HashName(name)
}

func (s *Server) LoginRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := s.GetEncFromSession(r.Context(), "userID")
		queries := db.New(s.DB)
		user, err := queries.GetUserByID(r.Context(), userID)
		if err != nil {
			location := r.URL.Path
			location = url.QueryEscape(location)
			s.Redirect(w, r, "/auth/login?next="+location)
			return
		}
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) MaybeUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := s.GetEncFromSession(r.Context(), "userID")
		queries := db.New(s.DB)
		user, err := queries.GetUserByID(r.Context(), userID)
		var ctx context.Context
		if user.ID != "" && err == nil {
			ctx = context.WithValue(r.Context(), "user", user)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) SetActivePracticePlanID(ctx context.Context, planID string, userID string) error {
	queries := db.New(s.DB)
	return queries.SetActivePracticePlan(ctx, db.SetActivePracticePlanParams{
		ActivePracticePlanID: sql.NullString{Valid: true, String: planID},
		UserID:               userID,
	})

}

func (s *Server) GetActivePracticePlanID(ctx context.Context) (string, bool) {
	queries := db.New(s.DB)
	practicePlanID, ok := ctx.Value("activePracticePlanID").(string)
	if ok && practicePlanID != "" {
		return practicePlanID, true
	}
	user, ok := ctx.Value("user").(db.User)
	if !ok || user.ID == "" {
		return "", false
	}
	if user.ActivePracticePlanID.Valid {
		if !user.ActivePracticePlanStarted.Valid || time.Since(time.Unix(user.ActivePracticePlanStarted.Int64, 0)) > 5*time.Hour {
			err := queries.ClearActivePracticePlan(ctx, user.ID)
			if err != nil {
				log.Default().Printf("failed to clear active practice plan: %v\n", err)
				return "", false
			}
			return "", false
		}
		return user.ActivePracticePlanID.String, true
	} else {
		return "", false
	}
}

func (s *Server) MaybePracticePlan(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		practicePlanID, ok := s.GetActivePracticePlanID(r.Context())
		var ctx context.Context
		if ok {
			ctx = context.WithValue(r.Context(), "activePracticePlanID", practicePlanID)
		} else {
			ctx = context.WithValue(r.Context(), "activePracticePlanID", "")
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) ContextPath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "currentPath", r.URL.Path)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
