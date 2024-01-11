// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package db

import (
	"database/sql"
)

type Credential struct {
	CredentialID    []byte `json:"credentialId"`
	PublicKey       []byte `json:"publicKey"`
	Transport       []byte `json:"transport"`
	AttestationType string `json:"attestationType"`
	Flags           []byte `json:"flags"`
	Authenticator   []byte `json:"authenticator"`
	UserID          string `json:"userId"`
}

type Piece struct {
	ID              string         `json:"id"`
	Title           string         `json:"title"`
	Description     sql.NullString `json:"description"`
	Composer        sql.NullString `json:"composer"`
	Measures        sql.NullInt64  `json:"measures"`
	BeatsPerMeasure sql.NullInt64  `json:"beatsPerMeasure"`
	GoalTempo       sql.NullInt64  `json:"goalTempo"`
	UserID          string         `json:"userId"`
	LastPracticed   sql.NullInt64  `json:"lastPracticed"`
	Stage           string         `json:"stage"`
}

type PracticePlan struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Intensity string `json:"intensity"`
	Date      int64  `json:"date"`
	Completed bool   `json:"completed"`
}

type PracticePlanPiece struct {
	PracticePlanID string `json:"practicePlanId"`
	PieceID        string `json:"pieceId"`
	PracticeType   string `json:"practiceType"`
	Completed      bool   `json:"completed"`
	Sessions       int64  `json:"sessions"`
	Idx            int64  `json:"idx"`
}

type PracticePlanSpot struct {
	PracticePlanID string `json:"practicePlanId"`
	SpotID         string `json:"spotId"`
	PracticeType   string `json:"practiceType"`
	Completed      bool   `json:"completed"`
	Idx            int64  `json:"idx"`
}

type Spot struct {
	ID             string         `json:"id"`
	PieceID        string         `json:"pieceId"`
	Name           string         `json:"name"`
	Stage          string         `json:"stage"`
	Measures       sql.NullString `json:"measures"`
	AudioPromptUrl string         `json:"audioPromptUrl"`
	ImagePromptUrl string         `json:"imagePromptUrl"`
	NotesPrompt    string         `json:"notesPrompt"`
	TextPrompt     string         `json:"textPrompt"`
	CurrentTempo   sql.NullInt64  `json:"currentTempo"`
	LastPracticed  sql.NullInt64  `json:"lastPracticed"`
	StageStarted   sql.NullInt64  `json:"stageStarted"`
	SkipDays       int64          `json:"skipDays"`
	Priority       int64          `json:"priority"`
}

type User struct {
	ID                        string         `json:"id"`
	Fullname                  string         `json:"fullname"`
	Email                     string         `json:"email"`
	EmailVerified             sql.NullBool   `json:"emailVerified"`
	ActivePracticePlanID      sql.NullString `json:"activePracticePlanId"`
	ActivePracticePlanStarted sql.NullInt64  `json:"activePracticePlanStarted"`
}
