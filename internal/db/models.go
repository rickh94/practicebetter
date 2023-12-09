// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.23.0

package db

import (
	"database/sql"
)

type Credential struct {
	CredentialID    []byte
	PublicKey       []byte
	Transport       []byte
	AttestationType string
	Flags           []byte
	Authenticator   []byte
	UserID          string
}

type Piece struct {
	ID              string
	Title           string
	Description     sql.NullString
	Composer        sql.NullString
	Measures        sql.NullInt64
	BeatsPerMeasure sql.NullInt64
	GoalTempo       sql.NullInt64
	UserID          string
	LastPracticed   sql.NullInt64
}

type PracticePiece struct {
	PracticeSessionID string
	PieceID           string
	Measures          string
}

type PracticeSession struct {
	ID              string
	DurationMinutes int64
	Date            int64
	UserID          string
}

type PracticeSpot struct {
	PracticeSessionID string
	SpotID            string
}

type Spot struct {
	ID             string
	PieceID        string
	Name           string
	Idx            int64
	Stage          string
	Measures       sql.NullString
	AudioPromptUrl string
	ImagePromptUrl string
	NotesPrompt    string
	TextPrompt     string
	CurrentTempo   sql.NullInt64
	LastPracticed  sql.NullInt64
	Priority       int64
}

type User struct {
	ID            string
	Fullname      string
	Email         string
	EmailVerified sql.NullBool
}
