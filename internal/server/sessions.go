package server

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"time"
)

func (s *Server) SaveEncToSession(c context.Context, key string, value string) {
	s.SM.Put(c, key, s.Encrypt(value))
}

func (s *Server) GetEncFromSession(c context.Context, key string) string {
	val := s.SM.GetString(c, key)
	if val == "" {
		return ""
	}
	return s.Decrypt(val)
}

func (s *Server) SaveOTP(c context.Context, code string) {
	s.SaveEncToSession(c, "code", code)
	s.SM.Put(c, "codeCreated", time.Now().Unix())
}

func (s *Server) CheckOTP(c context.Context, submittedCode string) bool {
	expectedCode := s.GetEncFromSession(c, "code")
	created := s.SM.GetInt64(c, "codeCreated")
	createdTime := time.Unix(created, 0)

	if time.Since(createdTime) > 5*time.Minute {
		return false
	}
	if submittedCode == expectedCode {
		s.SM.Remove(c, "code")
		s.SM.Remove(c, "codeCreated")
		return true
	}
	return false
}

func (s *Server) Encrypt(plaintext string) string {
	aes, err := aes.NewCipher([]byte(s.SecretKey[0:32]))
	if err != nil {
		panic(err)
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		panic(err)
	}

	// We need a 12-byte nonce for GCM (modifiable if you use cipher.NewGCMWithNonceSize())
	// A nonce should always be randomly generated for every encryption.
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		panic(err)
	}

	// ciphertext here is actually nonce+ciphertext
	// So that when we decrypt, just knowing the nonce size
	// is enough to separate it from the ciphertext.
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return string(ciphertext)
}

func (s *Server) Decrypt(ciphertext string) string {
	aes, err := aes.NewCipher([]byte(s.SecretKey[0:32]))
	if err != nil {
		panic(err)
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		panic(err)
	}

	// Since we know the ciphertext is actually nonce+ciphertext
	// And len(nonce) == NonceSize(). We can separate the two.
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, []byte(nonce), []byte(ciphertext), nil)
	if err != nil {
		panic(err)
	}

	return string(plaintext)
}

func (s *Server) LoginUser(ctx context.Context, userID string) error {
	err := s.SM.RenewToken(ctx)
	if err != nil {
		return err
	}
	s.SaveEncToSession(ctx, "userID", userID)
	return nil
}

func (s *Server) LogoutUser(ctx context.Context) error {
	err := s.SM.RenewToken(ctx)
	if err != nil {
		return err
	}
	s.SM.Remove(ctx, "userID")
	return nil
}

type PracticeBreak struct {
	TimeTaken time.Time
	PlanID    string
}

const LAST_BREAK_KEY = "lastBreak"

func (s *Server) SetLastBreak(ctx context.Context, planID string) {
	s.SM.Put(ctx, LAST_BREAK_KEY, PracticeBreak{
		time.Now(),
		planID,
	})
}

func (s *Server) GetLastBreak(ctx context.Context, planID string) (time.Time, bool) {
	b, ok := s.SM.Get(ctx, LAST_BREAK_KEY).(PracticeBreak)
	if !ok || b.PlanID != planID {
		s.ClearLastBreak(ctx)
		s.SetLastBreak(ctx, planID)
		return time.Now(), false
	}
	return b.TimeTaken, true
}

func (s *Server) ClearLastBreak(ctx context.Context) {
	s.SM.Remove(ctx, LAST_BREAK_KEY)

}
