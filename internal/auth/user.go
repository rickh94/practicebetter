package auth

import (
	"context"
	"database/sql"
	"log"
	"practicebetter/internal/db"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

type User struct {
	ID       string
	Email    string
	Fullname string
	DB       *sql.DB
	ctx      context.Context
}

func NewUserFromForLogin(user db.GetUserForLoginRow, sqlDB *sql.DB, ctx context.Context) *User {
	return &User{
		ID:       user.ID,
		Email:    user.Email,
		Fullname: user.Fullname,
		DB:       sqlDB,
		ctx:      ctx,
	}
}

func NewUser(user db.User, sqlDB *sql.DB, ctx context.Context) *User {
	return &User{
		ID:       user.ID,
		Email:    user.Email,
		Fullname: user.Fullname,
		DB:       sqlDB,
		ctx:      ctx,
	}
}

func (u *User) WebAuthnID() []byte {
	return []byte(u.ID)
}

func (u *User) WebAuthnName() string {
	if u.Fullname != "" {
		return u.Fullname
	} else {
		return u.Email
	}
}

func (u *User) WebAuthnDisplayName() string {
	return string(u.WebAuthnName())
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	queries := db.New(u.DB)
	var credentials []webauthn.Credential
	dbCredentials, err := queries.GetUserCredentials(u.ctx, u.ID)
	if err != nil {
		return credentials
	}
	for _, credential := range dbCredentials {
		var transport []protocol.AuthenticatorTransport
		err := cbor.Unmarshal(credential.Transport, &transport)
		if err != nil {
			log.Print("failed to unmarshal transport: ")
			log.Println(err)
			continue
		}

		var flags webauthn.CredentialFlags
		err = cbor.Unmarshal(credential.Flags, &flags)
		if err != nil {
			log.Print("failed to unmarshal flags: ")
			log.Println(err)
			continue
		}

		var authenticator webauthn.Authenticator
		err = cbor.Unmarshal(credential.Authenticator, &authenticator)
		if err != nil {
			log.Print("failed to unmarshal flags: ")
			log.Println(err)
			continue
		}

		credentials = append(credentials, webauthn.Credential{
			ID:              credential.CredentialID,
			PublicKey:       credential.PublicKey,
			AttestationType: credential.AttestationType,
			Transport:       transport,
			Flags:           flags,
			Authenticator:   authenticator,
		})
	}
	return credentials
}

func (u *User) WebAuthnIcon() string {
	return ""
}

func (u *User) AddCredential(newCredential *webauthn.Credential) error {
	transport, err := cbor.Marshal(newCredential.Transport)
	if err != nil {
		return err
	}
	flags, err := cbor.Marshal(newCredential.Flags)
	if err != nil {
		return err
	}
	authenticator, err := cbor.Marshal(newCredential.Authenticator)
	if err != nil {
		return err
	}
	queries := db.New(u.DB)
	_, err = queries.CreateCredential(u.ctx, db.CreateCredentialParams{
		UserID:          u.ID,
		CredentialID:    newCredential.ID,
		PublicKey:       newCredential.PublicKey,
		Transport:       transport,
		AttestationType: newCredential.AttestationType,
		Flags:           flags,
		Authenticator:   authenticator,
	})
	if err != nil {
		return err
	}
	return nil
}

/*
func DeletePasskeys(user *User) error {
	result := DB.Where("user_id = ?", user.ID).Delete(&Credential{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func CountUserCredentials(user *User) (int64, error) {
	var count int64
	result := DB.Model(&Credential{}).Where("user_id = ?", user.ID).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}
*/
