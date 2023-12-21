package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"practicebetter/internal/auth"
	"practicebetter/internal/db"
	"practicebetter/internal/pages/authpages"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/mavolin/go-htmx"
)

func (s *Server) startLogin(w http.ResponseWriter, r *http.Request) {

	/* TODO:
	if RedirectLoggedIn(w, r, "/auth/me") {
		return
	}
	*/
	csrfToken := csrf.Token(r)
	nextLoc := r.URL.Query().Get("next")
	if nextLoc == "" {
		nextLoc = "/auth/me"
	}
	cookie, err := r.Cookie("rememberEmail")
	fmt.Println(cookie)
	if err != nil {
		s.HxRender(w, r, authpages.StartLoginPage(csrfToken, nextLoc), "Login")
		return
	}
	queries := db.New(s.DB)
	user, err := queries.GetUserForLogin(r.Context(), cookie.Value)
	if err != nil {
		http.SetCookie(w, &http.Cookie{
			Name:   "rememberEmail",
			Value:  "",
			MaxAge: -1,
		})
		s.HxRender(w, r, authpages.StartLoginPage(csrfToken, nextLoc), "Login")
		return
	}

	s.SM.Put(r.Context(), "rememberMe", true)
	if user.CredentialCount == 0 || !user.EmailVerified.Bool {
		s.continueOtpSignIn(w, r, user.Email, nextLoc)
		return
	} else {
		s.continuePasskeySignIn(w, r, user, nextLoc)
		return
	}
}

// TODO: error message for code

func (s *Server) continueLogin(w http.ResponseWriter, r *http.Request) {
	/*
		if RedirectLoggedIn(w, r, "/auth/me") {
			return
		}
	*/
	r.ParseForm()

	userEmail := r.Form.Get("email")
	if userEmail == "" {
		w.WriteHeader(http.StatusBadRequest)
		// TODO: send back the form with an error
		return
	}
	userEmail = strings.ToLower(userEmail)
	remember := r.Form.Get("remember")
	if remember == "on" {
		s.SM.Put(r.Context(), "rememberMe", true)
	}

	nextLoc := r.Form.Get("next")
	if nextLoc == "" {
		nextLoc = "/nextLoc"
	}
	queries := db.New(s.DB)
	user, err := queries.GetUserForLogin(r.Context(), userEmail)
	if err != nil || user.CredentialCount == 0 || !user.EmailVerified.Bool {
		s.continueOtpSignIn(w, r, userEmail, nextLoc)
		return
	} else {
		s.continuePasskeySignIn(w, r, user, nextLoc)
		return
	}
}

func (s *Server) continueOtpSignIn(w http.ResponseWriter, r *http.Request, userEmail string, nextLoc string) {
	token := csrf.Token(r)

	code, err := auth.GenerateOTP(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		// TODO: send back the form with an error
		return
	}

	s.SaveOTP(r.Context(), code)
	message := fmt.Sprintf("Your one-time login code is %s. It will expire in 5 minutes.", code)
	go s.SendEmail(userEmail, "Practice Better: Login Code", message)
	s.SaveEncToSession(r.Context(), "email", userEmail)

	htmx.TriggerAfterSettle(r, "ShowAlert", ShowAlertEvent{
		Message:  "Login Code sent to your email",
		Title:    "Code Sent!",
		Variant:  "success",
		Duration: 3000,
	})
	s.HxRender(w, r, authpages.FinishCodeLoginPage(token, nextLoc), "Login")

	return
}

func (s *Server) continuePasskeySignIn(w http.ResponseWriter, r *http.Request, user db.GetUserForLoginRow, nextLoc string) {
	token := csrf.Token(r)
	s.SaveEncToSession(r.Context(), "email", user.Email)
	options, err := s.BeginPasskeyLogin(r, user)
	if err != nil {
		s.continueOtpSignIn(w, r, user.Email, nextLoc)
		return
	}
	s.HxRender(w, r, authpages.FinishPasskeyLoginPage(options, token, nextLoc), "Login")
	return
}

func (s *Server) completeCodeLogin(w http.ResponseWriter, r *http.Request) {
	/*
		if RedirectLoggedIn(w, r, "/") {
			return
		}
	*/
	userEmail := s.GetEncFromSession(r.Context(), "email")
	if userEmail == "" {
		log.Default().Printf("user email not found in session: %s\n", userEmail)
		w.WriteHeader(http.StatusBadRequest)
		// TODO: redirect back to the main form with an error
		return
	}

	r.ParseForm()
	nextLoc := r.Form.Get("next")
	if nextLoc == "" {
		nextLoc = "/library"
	}
	submittedCode := r.Form.Get("code")
	if s.CheckOTP(r.Context(), submittedCode) {
		queries := db.New(s.DB)
		user, err := queries.GetOrCreateUser(r.Context(), userEmail)
		go queries.SetEmailVerified(r.Context(), user.ID)
		if err != nil {
			log.Default().Printf("Database error: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			// TODO: re-render the form with an error
			return
		}
		if s.SM.Get(r.Context(), "rememberMe").(bool) {
			cookie := http.Cookie{
				Name:     "rememberEmail",
				Value:    user.Email,
				MaxAge:   60 * 60 * 24 * 7,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			}
			http.SetCookie(w, &cookie)
		}
		s.LoginUser(r.Context(), user.ID)
		s.Redirect(w, r, nextLoc)
		return
	} else {
		log.Default().Println("Login Rejected for incorrect code")
		w.WriteHeader(http.StatusBadRequest)
		// TODO: re-render the form with an error
		return
	}

}

func (s *Server) completePasskeySignin(w http.ResponseWriter, r *http.Request) {
	userEmail := s.GetEncFromSession(r.Context(), "email")
	queries := db.New(s.DB)
	user, err := queries.GetUserForLogin(r.Context(), userEmail)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		// TODO: re-render the form with an error
		return
	}

	if err := s.FinishPasskeyLogin(r, user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Failed to log in"})
		return
	} else {
		if s.SM.Get(r.Context(), "rememberMe").(bool) {
			cookie := http.Cookie{
				Name:     "rememberEmail",
				Value:    user.Email,
				Path:     "/",
				MaxAge:   60 * 60 * 24 * 7,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
			}
			http.SetCookie(w, &cookie)
		}
		s.LoginUser(r.Context(), user.ID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "redirect": "/auth/me"})
		return
	}
}

func (s *Server) logoutUserRoute(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "rememberEmail",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	})
	s.LogoutUser(r.Context())
	s.Redirect(w, r, "/")
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	queries := db.New(s.DB)
	credentialCount, err := queries.CountUserCredentials(r.Context(), user.ID)
	if err != nil {
		fmt.Println("Could not count credentials")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "text/html")

	registrationOptions, err := s.BeginPasskeyRegistration(r.Context(), user)
	if err != nil {
		fmt.Println("Could not find registration options")
	}

	token := csrf.Token(r)
	component := authpages.MePage(user, registrationOptions, token, fmt.Sprintf("%d", credentialCount))
	s.HxRender(w, r, component, "Account")
}

func (s *Server) editProfile(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	token := csrf.Token(r)
	component := authpages.UserForm(user, token, nil)
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

func (s *Server) updateProfile(w http.ResponseWriter, r *http.Request) {
	errors := make(map[string]string)
	hasError := false

	user := r.Context().Value("user").(db.User)
	r.ParseForm()

	email := r.Form.Get("email")
	_, err := mail.ParseAddress(email)
	if err != nil {
		errors["email"] = "Invalid email"
		hasError = true
		email = user.Email
	}

	fullname := r.Form.Get("name")
	token := csrf.Token(r)

	if hasError {
		component := authpages.UserForm(user, token, errors)
		htmx.TriggerAfterSettle(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not update your profile",
			Title:    "Invalid Input",
			Variant:  "error",
			Duration: 3000,
		})
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "text/html")
		component.Render(r.Context(), w)
		return
	}

	queries := db.New(s.DB)
	needsReverify := strings.ToLower(email) != user.Email
	user, err = queries.UpdateUser(r.Context(), db.UpdateUserParams{
		Fullname: fullname,
		Email:    strings.ToLower(email),
		EmailVerified: sql.NullBool{
			Bool:  !needsReverify,
			Valid: true,
		},
		ID: user.ID,
	})
	if err != nil {
		log.Default().Printf("Database error: %v\n", err)
		htmx.TriggerAfterSettle(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not update your profiled",
			Title:    "Database Error",
			Variant:  "error",
			Duration: 3000,
		})
		w.Header().Set("Content-Type", "text/html")
		component := authpages.UserForm(user, token, nil)
		component.Render(r.Context(), w)
		return
	}

	htmx.TriggerAfterSwap(r, "ShowAlert", ShowAlertEvent{
		Message:  "Your profile has been updated",
		Title:    "Update Complete",
		Variant:  "success",
		Duration: 3000,
	})
	if needsReverify {
		htmx.TriggerAfterSettle(r, "ShowAlert", ShowAlertEvent{
			Message:  "Since you changed your email, you need to log out and back in to re-verify it.",
			Title:    "Verify Email",
			Variant:  "info",
			Duration: 5000,
		})
	}
	w.WriteHeader(http.StatusOK)
	component := authpages.UserInfo(user, token)
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

func (s *Server) getProfile(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	token := csrf.Token(r)
	component := authpages.UserInfo(user, token)
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

func (s *Server) registerPasskey(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	if err := s.FinishPasskeyRegistration(r, user); err != nil {
		log.Default().Printf("Error registering passkey for user %s: %v", user.Email, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not register passkey"))
		return
	}
	w.Write([]byte("OK"))
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) deletePasskeys(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(db.User)
	queries := db.New(s.DB)
	if err := queries.DeleteUserCredentials(r.Context(), user.ID); err != nil {
		log.Default().Printf("Error deleting passkeys for user %s: %v", user.Email, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not delete passkeys"))
		return
	}
	htmx.TriggerAfterSettle(r, "ShowAlert", ShowAlertEvent{
		Message:  "All your passkeys have been deleted. Consider registering a new one!",
		Title:    "Passkeys Deleted!",
		Variant:  "success",
		Duration: 3000,
	})
	w.Write([]byte("OK"))
	w.WriteHeader(http.StatusOK)
}

func (s *Server) forceCodeLogin(w http.ResponseWriter, r *http.Request) {
	userEmail := s.GetEncFromSession(r.Context(), "email")
	if userEmail == "" {
		w.WriteHeader(http.StatusBadRequest)
		// TODO: send back the form with an error
		return
	}
	nextLoc := r.Form.Get("next")
	if nextLoc == "" {
		nextLoc = "/auth/me"
	}
	s.continueOtpSignIn(w, r, userEmail, nextLoc)
	return
}
