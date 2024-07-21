package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"practicebetter/internal/auth"
	"practicebetter/internal/ck"
	"practicebetter/internal/db"
	"practicebetter/internal/pages/authpages"
	"strconv"
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
	if err != nil {
		s.HxRender(w, r, authpages.StartLoginPage(csrfToken, nextLoc), "Login")
		return
	}
	queries := db.New(s.DB)
	user, err := queries.GetUserForLogin(r.Context(), cookie.Value)
	if err != nil {
		http.SetCookie(w, &http.Cookie{
			Name:     "rememberEmail",
			Value:    "",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			MaxAge:   -1,
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
	err := r.ParseForm()
	if err != nil {
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not login",
			Title:    "Invalid Input",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Invalid Input", http.StatusBadRequest)
		return
	}

	honeypot := r.Form.Get("name")
	if honeypot != "" {
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not login",
			Title:    "Invalid Input",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Invalid Input", http.StatusBadRequest)
		return
	}

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
	} else {
		s.SM.Put(r.Context(), "rememberMe", false)
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

	if err := htmx.TriggerAfterSettle(r, "ShowAlert", ShowAlertEvent{
		Message:  "Login Code sent to your email",
		Title:    "Code Sent!",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}
	s.HxRender(w, r, authpages.FinishCodeLoginPage(token, nextLoc), "Login")
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

	err := r.ParseForm()
	if err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not login",
			Title:    "Invalid Input",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Invalid Input", http.StatusBadRequest)
		return
	}
	nextLoc := r.Form.Get("next")
	if nextLoc == "" {
		nextLoc = "/library"
	}
	submittedCode := r.Form.Get("code")
	if s.CheckOTP(r.Context(), submittedCode) {
		queries := db.New(s.DB)
		user, err := queries.GetOrCreateUser(r.Context(), userEmail)
		go s.setEmailVerified(r.Context(), user.ID)
		if err != nil {
			log.Default().Printf("Database error: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			// TODO: re-render the form with an error
			return
		}
		if val, ok := s.SM.Get(r.Context(), "rememberMe").(bool); val && ok {
			cookie := http.Cookie{
				Name:     "rememberEmail",
				Path:     "/",
				Value:    user.Email,
				MaxAge:   60 * 60 * 24 * 7,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			}
			http.SetCookie(w, &cookie)
		}
		if err := s.LoginUser(r.Context(), user.ID); err != nil {
			log.Default().Printf("Could not login user: %v\n", err)
			if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
				Message:  "Could not log you in with that information.",
				Title:    "Login Failed",
				Variant:  "error",
				Duration: 3000}); err != nil {
				log.Default().Println(err)
			}
			http.Error(w, "Could not log you in with that information.", http.StatusUnauthorized)
			return
		}
		nextLocCookie := http.Cookie{
			Name:     "nextLoc",
			Path:     "/",
			Value:    nextLoc,
			MaxAge:   60 * 5,
			HttpOnly: false,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(w, &nextLocCookie)
		s.Redirect(w, r, "/auth/me?recommend=1")
	} else {
		log.Default().Println("Login Rejected for incorrect code")
		err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Incorrect Code.",
			Title:    "Login Failed",
			Variant:  "error",
			Duration: 3000,
		})
		if err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Login Rejected for incorrect code", http.StatusUnauthorized)
		// TODO: re-render the form with an error
	}

}

func (s *Server) setEmailVerified(ctx context.Context, userID string) {
	queries := db.New(s.DB)
	err := queries.SetEmailVerified(ctx, userID)
	if err != nil {
		log.Default().Printf("Could not set email verified, Database error: %v\n", err)
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
		err := json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Failed to log in"})
		if err != nil {
			log.Default().Println(err)
		}
		return
	} else {
		if val, ok := s.SM.Get(r.Context(), "rememberMe").(bool); val && ok {
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
		if err := s.LoginUser(r.Context(), user.ID); err != nil {
			err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
				Message:  "Something went wrong.",
				Title:    "Login Failed",
				Variant:  "error",
				Duration: 3000,
			})
			if err != nil {
				log.Default().Println(err)
			}
			http.Error(w, "Something went wrong.", http.StatusInternalServerError)
			// TODO: re-render the form with an error
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(map[string]string{"status": "ok", "redirect": "/auth/me"})
		if err != nil {
			log.Default().Println(err)
		}
	}
}

func (s *Server) logoutUserRoute(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "rememberEmail",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   -1,
	})
	err := s.LogoutUser(r.Context())
	if err != nil {
		log.Default().Println(err)
	}
	s.Redirect(w, r, "/")
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
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
	component := authpages.MePage(user, registrationOptions, token, fmt.Sprintf("%d", credentialCount), s)
	s.HxRender(w, r, component, "Account")
}

func (s *Server) editProfile(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	token := csrf.Token(r)
	component := authpages.UserForm(user, token, nil)
	w.Header().Set("Content-Type", "text/html")
	err := component.Render(r.Context(), w)
	if err != nil {
		log.Default().Println(err)
	}
}

func (s *Server) updateProfile(w http.ResponseWriter, r *http.Request) {
	errors := make(map[string]string)
	hasError := false

	user := r.Context().Value(ck.UserKey).(db.User)
	if err := r.ParseForm(); err != nil {
		log.Default().Println(err)
		if err := htmx.Trigger(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not update your profile",
			Title:    "Invalid Input",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		http.Error(w, "Invalid Input.", http.StatusBadRequest)
		return
	}

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
		if err := htmx.TriggerAfterSettle(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not update your profile",
			Title:    "Invalid Input",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "text/html")
		if err := component.Render(r.Context(), w); err != nil {
			log.Default().Println(err)
		}
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
		if err := htmx.TriggerAfterSettle(r, "ShowAlert", ShowAlertEvent{
			Message:  "Could not update your profiled",
			Title:    "Database Error",
			Variant:  "error",
			Duration: 3000,
		}); err != nil {
			log.Default().Println(err)
		}
		component := authpages.UserForm(user, token, nil)
		if err := component.Render(r.Context(), w); err != nil {
			log.Default().Println(err)
		}
		return
	}

	if err := htmx.TriggerAfterSwap(r, "ShowAlert", ShowAlertEvent{
		Message:  "Your profile has been updated",
		Title:    "Update Complete",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}
	if needsReverify {
		if err := htmx.TriggerAfterSettle(r, "ShowAlert", ShowAlertEvent{
			Message:  "Since you changed your email, you need to log out and back in to re-verify it.",
			Title:    "Verify Email",
			Variant:  "info",
			Duration: 5000,
		}); err != nil {
			log.Default().Println(err)
		}
	}
	w.WriteHeader(http.StatusOK)
	if err := authpages.UserInfo(user, token).Render(r.Context(), w); err != nil {
		log.Default().Println(err)
	}
}

func (s *Server) getProfile(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	token := csrf.Token(r)
	w.Header().Set("Content-Type", "text/html")
	if err := authpages.UserInfo(user, token).Render(r.Context(), w); err != nil {
		log.Default().Println(err)
	}
}

func (s *Server) registerPasskey(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	if err := s.FinishPasskeyRegistration(r, user); err != nil {
		log.Default().Printf("Error registering passkey for user %s: %v", user.Email, err)
		http.Error(w, "Could not register passkey", http.StatusInternalServerError)
		return
	}
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Default().Println(err)
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) deletePasskeys(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	queries := db.New(s.DB)
	if err := queries.DeleteUserCredentials(r.Context(), user.ID); err != nil {
		log.Default().Printf("Error deleting passkeys for user %s: %v", user.Email, err)
		http.Error(w, "Could not delete passkeys", http.StatusInternalServerError)
		return
	}
	if err := htmx.TriggerAfterSettle(r, "ShowAlert", ShowAlertEvent{
		Message:  "All your passkeys have been deleted. Consider registering a new one!",
		Title:    "Passkeys Deleted!",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}
	_, err := w.Write([]byte("OK"))
	if err != nil {
		log.Default().Println(err)
	}
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
}

func (s *Server) forgetUser(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "rememberEmail",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   -1,
	})
	s.Redirect(w, r, "/auth/login")
}

func (s *Server) updateSettings(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ck.UserKey).(db.User)
	queries := db.New(s.DB)

	if err := r.ParseForm(); err != nil {
		log.Default().Println(err)
		s.InvalidInputError(w, r, "Invalid input")
		return
	}

	timeBetweenBreaks, err := strconv.Atoi(r.Form.Get("config_time_between_breaks"))
	if err != nil {
		s.InvalidInputError(w, r, "Invalid time between breaks")
		return
	}
	practicePlanIntensity := r.Form.Get("config_default_plan_intensity")
	if practicePlanIntensity != "light" && practicePlanIntensity != "medium" && practicePlanIntensity != "heavy" {
		s.InvalidInputError(w, r, "Invalid plan intensity")
		return
	}

	user, err = queries.UpdateUserSettings(r.Context(), db.UpdateUserSettingsParams{
		ID:                         user.ID,
		ConfigTimeBetweenBreaks:    int64(timeBetweenBreaks),
		ConfigDefaultPlanIntensity: practicePlanIntensity,
	})
	if err != nil {
		log.Default().Println(err)
		s.DatabaseError(w, r, err, "Could not update settings.")
		return
	}

	if err := htmx.TriggerAfterSettle(r, "ShowAlert", ShowAlertEvent{
		Message:  "Successfully updated settings",
		Title:    "Settings Updated!",
		Variant:  "success",
		Duration: 3000,
	}); err != nil {
		log.Default().Println(err)
	}
	if err := authpages.UserSettingsForm(user, csrf.Token(r)).Render(r.Context(), w); err != nil {
		log.Default().Println(err)
		http.Error(w, "Render Error", http.StatusInternalServerError)
	}
}
