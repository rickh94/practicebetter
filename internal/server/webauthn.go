package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"practicebetter/internal/auth"
	"practicebetter/internal/db"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

func (s *Server) BeginPasskeyRegistration(c context.Context, userRow db.User) (*protocol.CredentialCreation, error) {
	user := auth.NewUser(userRow, s.DB, c)
	options, session, err := s.WebAuthn.BeginRegistration(user)
	if err != nil {
		return nil, err
	}
	// serialize the session to json and save it with the session manager
	sessionJson, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}
	s.SaveEncToSession(c, "webauthnRegistration", string(sessionJson))

	return options, nil
}

func (s *Server) FinishPasskeyRegistration(r *http.Request, userRow db.User) error {
	user := auth.NewUser(userRow, s.DB, r.Context())
	sessionJson := s.GetEncFromSession(r.Context(), "webauthnRegistration")
	var session webauthn.SessionData
	err := json.Unmarshal([]byte(sessionJson), &session)
	if err != nil {
		return err
	}

	credential, err := s.WebAuthn.FinishRegistration(user, session, r)
	if err != nil {
		return err
	}
	return user.AddCredential(credential)
}

func (s *Server) BeginPasskeyLogin(r *http.Request, userRow db.GetUserForLoginRow) (*protocol.CredentialAssertion, error) {
	user := auth.NewUserFromForLogin(userRow, s.DB, r.Context())
	options, session, err := s.WebAuthn.BeginLogin(user)
	if err != nil {
		return nil, err
	}

	sessionJson, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}
	s.SaveEncToSession(r.Context(), "webauthnLogin", string(sessionJson))

	return options, nil
}

func (s *Server) FinishPasskeyLogin(r *http.Request, userRow db.GetUserForLoginRow) error {
	user := auth.NewUserFromForLogin(userRow, s.DB, r.Context())
	sessionJson := s.GetEncFromSession(r.Context(), "webauthnLogin")
	var session webauthn.SessionData
	err := json.Unmarshal([]byte(sessionJson), &session)
	if err != nil {
		log.Default().Printf("failed to unmarshal session: %v\n", err)
		return err
	}

	_, err = s.WebAuthn.FinishLogin(user, session, r)
	if err != nil {
		log.Default().Printf("failed to login: %v\n", err)
		return err
	}

	return nil
}
