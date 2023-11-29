// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.23.0
// source: spots.sql

package db

import (
	"context"
	"database/sql"
	"strings"
)

const createSpot = `-- name: CreateSpot :one
INSERT INTO spots (
    piece_id,
    id,
    name,
    idx,
    stage,
    audio_prompt_url,
    image_prompt_url,
    notes_prompt,
    text_prompt,
    current_tempo,
    measures
) VALUES (
    (SELECT pieces.id FROM pieces WHERE pieces.user_id = ? AND pieces.id = ? LIMIT 1),
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING id, piece_id, name, idx, stage, measures, audio_prompt_url, image_prompt_url, notes_prompt, text_prompt, current_tempo
`

type CreateSpotParams struct {
	UserID         string
	PieceID        string
	ID             string
	Name           string
	Idx            int64
	Stage          string
	AudioPromptUrl string
	ImagePromptUrl string
	NotesPrompt    string
	TextPrompt     string
	CurrentTempo   sql.NullInt64
	Measures       sql.NullString
}

func (q *Queries) CreateSpot(ctx context.Context, arg CreateSpotParams) (Spot, error) {
	row := q.db.QueryRowContext(ctx, createSpot,
		arg.UserID,
		arg.PieceID,
		arg.ID,
		arg.Name,
		arg.Idx,
		arg.Stage,
		arg.AudioPromptUrl,
		arg.ImagePromptUrl,
		arg.NotesPrompt,
		arg.TextPrompt,
		arg.CurrentTempo,
		arg.Measures,
	)
	var i Spot
	err := row.Scan(
		&i.ID,
		&i.PieceID,
		&i.Name,
		&i.Idx,
		&i.Stage,
		&i.Measures,
		&i.AudioPromptUrl,
		&i.ImagePromptUrl,
		&i.NotesPrompt,
		&i.TextPrompt,
		&i.CurrentTempo,
	)
	return i, err
}

const deleteSpot = `-- name: DeleteSpot :exec
DELETE FROM spots
WHERE spots.id = ? AND spots.piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = ? AND pieces.id = ? LIMIT 1)
`

type DeleteSpotParams struct {
	ID     string
	UserID string
	ID_2   string
}

func (q *Queries) DeleteSpot(ctx context.Context, arg DeleteSpotParams) error {
	_, err := q.db.ExecContext(ctx, deleteSpot, arg.ID, arg.UserID, arg.ID_2)
	return err
}

const deleteSpotsExcept = `-- name: DeleteSpotsExcept :exec
DELETE FROM spots
WHERE
spots.piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = ?1 AND pieces.id = ?2 LIMIT 1)
AND spots.id NOT IN (/*SLICE:spotIDs*/?)
`

type DeleteSpotsExceptParams struct {
	UserID  string
	PieceID string
	SpotIDs []string
}

func (q *Queries) DeleteSpotsExcept(ctx context.Context, arg DeleteSpotsExceptParams) error {
	query := deleteSpotsExcept
	var queryParams []interface{}
	queryParams = append(queryParams, arg.UserID)
	queryParams = append(queryParams, arg.PieceID)
	if len(arg.SpotIDs) > 0 {
		for _, v := range arg.SpotIDs {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:spotIDs*/?", strings.Repeat(",?", len(arg.SpotIDs))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:spotIDs*/?", "NULL", 1)
	}
	_, err := q.db.ExecContext(ctx, query, queryParams...)
	return err
}

const getSpot = `-- name: GetSpot :one
SELECT
    spots.id, spots.piece_id, spots.name, spots.idx, spots.stage, spots.measures, spots.audio_prompt_url, spots.image_prompt_url, spots.notes_prompt, spots.text_prompt, spots.current_tempo,
    pieces.title as piece_title
FROM spots
INNER JOIN pieces ON pieces.id = spots.piece_id
WHERE spots.id = ?1 AND spots.piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = ?2 AND pieces.id = ?3 LIMIT 1)
`

type GetSpotParams struct {
	SpotID  string
	UserID  string
	PieceID string
}

type GetSpotRow struct {
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
	PieceTitle     string
}

func (q *Queries) GetSpot(ctx context.Context, arg GetSpotParams) (GetSpotRow, error) {
	row := q.db.QueryRowContext(ctx, getSpot, arg.SpotID, arg.UserID, arg.PieceID)
	var i GetSpotRow
	err := row.Scan(
		&i.ID,
		&i.PieceID,
		&i.Name,
		&i.Idx,
		&i.Stage,
		&i.Measures,
		&i.AudioPromptUrl,
		&i.ImagePromptUrl,
		&i.NotesPrompt,
		&i.TextPrompt,
		&i.CurrentTempo,
		&i.PieceTitle,
	)
	return i, err
}

const listPieceSpots = `-- name: ListPieceSpots :many
SELECT
    id,
    name,
    idx,
    stage,
    audio_prompt_url,
    image_prompt_url,
    notes_prompt,
    text_prompt,
    current_tempo,
    measures
FROM spots
WHERE piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = ? AND pieces.id = ? LIMIT 1)
`

type ListPieceSpotsParams struct {
	UserID string
	ID     string
}

type ListPieceSpotsRow struct {
	ID             string
	Name           string
	Idx            int64
	Stage          string
	AudioPromptUrl string
	ImagePromptUrl string
	NotesPrompt    string
	TextPrompt     string
	CurrentTempo   sql.NullInt64
	Measures       sql.NullString
}

func (q *Queries) ListPieceSpots(ctx context.Context, arg ListPieceSpotsParams) ([]ListPieceSpotsRow, error) {
	rows, err := q.db.QueryContext(ctx, listPieceSpots, arg.UserID, arg.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListPieceSpotsRow
	for rows.Next() {
		var i ListPieceSpotsRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Idx,
			&i.Stage,
			&i.AudioPromptUrl,
			&i.ImagePromptUrl,
			&i.NotesPrompt,
			&i.TextPrompt,
			&i.CurrentTempo,
			&i.Measures,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateSpot = `-- name: UpdateSpot :one
UPDATE spots
SET
    name = ?,
    idx = ?,
    stage = ?,
    audio_prompt_url = ?,
    image_prompt_url = ?,
    notes_prompt = ?,
    text_prompt = ?,
    current_tempo = ?,
    measures = ?
WHERE spots.id = ? AND piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = ? AND pieces.id = ? LIMIT 1)
RETURNING id, piece_id, name, idx, stage, measures, audio_prompt_url, image_prompt_url, notes_prompt, text_prompt, current_tempo
`

type UpdateSpotParams struct {
	Name           string
	Idx            int64
	Stage          string
	AudioPromptUrl string
	ImagePromptUrl string
	NotesPrompt    string
	TextPrompt     string
	CurrentTempo   sql.NullInt64
	Measures       sql.NullString
	SpotID         string
	UserID         string
	PieceID        string
}

func (q *Queries) UpdateSpot(ctx context.Context, arg UpdateSpotParams) (Spot, error) {
	row := q.db.QueryRowContext(ctx, updateSpot,
		arg.Name,
		arg.Idx,
		arg.Stage,
		arg.AudioPromptUrl,
		arg.ImagePromptUrl,
		arg.NotesPrompt,
		arg.TextPrompt,
		arg.CurrentTempo,
		arg.Measures,
		arg.SpotID,
		arg.UserID,
		arg.PieceID,
	)
	var i Spot
	err := row.Scan(
		&i.ID,
		&i.PieceID,
		&i.Name,
		&i.Idx,
		&i.Stage,
		&i.Measures,
		&i.AudioPromptUrl,
		&i.ImagePromptUrl,
		&i.NotesPrompt,
		&i.TextPrompt,
		&i.CurrentTempo,
	)
	return i, err
}
