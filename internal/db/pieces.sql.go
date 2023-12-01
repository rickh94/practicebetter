// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.23.0
// source: pieces.sql

package db

import (
	"context"
	"database/sql"
)

const countUserPieces = `-- name: CountUserPieces :one
SELECT COUNT(*) FROM pieces WHERE user_id = ?
`

func (q *Queries) CountUserPieces(ctx context.Context, userID string) (int64, error) {
	row := q.db.QueryRowContext(ctx, countUserPieces, userID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createPiece = `-- name: CreatePiece :one
INSERT INTO pieces (
    id,
    title,
    description,
    composer,
    measures,
    beats_per_measure,
    goal_tempo,
    user_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id, title, description, composer, measures, beats_per_measure, goal_tempo, user_id, last_practiced
`

type CreatePieceParams struct {
	ID              string
	Title           string
	Description     sql.NullString
	Composer        sql.NullString
	Measures        sql.NullInt64
	BeatsPerMeasure sql.NullInt64
	GoalTempo       sql.NullInt64
	UserID          string
}

func (q *Queries) CreatePiece(ctx context.Context, arg CreatePieceParams) (Piece, error) {
	row := q.db.QueryRowContext(ctx, createPiece,
		arg.ID,
		arg.Title,
		arg.Description,
		arg.Composer,
		arg.Measures,
		arg.BeatsPerMeasure,
		arg.GoalTempo,
		arg.UserID,
	)
	var i Piece
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Description,
		&i.Composer,
		&i.Measures,
		&i.BeatsPerMeasure,
		&i.GoalTempo,
		&i.UserID,
		&i.LastPracticed,
	)
	return i, err
}

const deletePiece = `-- name: DeletePiece :exec
DELETE FROM pieces
WHERE id = ? AND user_id = ?
`

type DeletePieceParams struct {
	ID     string
	UserID string
}

func (q *Queries) DeletePiece(ctx context.Context, arg DeletePieceParams) error {
	_, err := q.db.ExecContext(ctx, deletePiece, arg.ID, arg.UserID)
	return err
}

const getPieceByID = `-- name: GetPieceByID :many
SELECT
    pieces.id,
    pieces.title,
    pieces.description,
    pieces.composer,
    pieces.measures,
    pieces.beats_per_measure,
    pieces.goal_tempo,
    pieces.last_practiced,
    spots.id AS spot_id,
    spots.name AS spot_name,
    spots.idx AS spot_idx,
    spots.stage AS spot_stage,
    spots.audio_prompt_url AS spot_audio_prompt_url,
    spots.image_prompt_url AS spot_image_prompt_url,
    spots.notes_prompt AS spot_notes_prompt,
    spots.text_prompt AS spot_text_prompt,
    spots.current_tempo AS spot_current_tempo,
    spots.measures AS spot_measures
FROM pieces
LEFT JOIN spots ON pieces.id = spots.piece_id
WHERE pieces.id = ?1 AND pieces.user_id = ?2
`

type GetPieceByIDParams struct {
	PieceID string
	UserID  string
}

type GetPieceByIDRow struct {
	ID                 string
	Title              string
	Description        sql.NullString
	Composer           sql.NullString
	Measures           sql.NullInt64
	BeatsPerMeasure    sql.NullInt64
	GoalTempo          sql.NullInt64
	LastPracticed      sql.NullInt64
	SpotID             sql.NullString
	SpotName           sql.NullString
	SpotIdx            sql.NullInt64
	SpotStage          sql.NullString
	SpotAudioPromptUrl sql.NullString
	SpotImagePromptUrl sql.NullString
	SpotNotesPrompt    sql.NullString
	SpotTextPrompt     sql.NullString
	SpotCurrentTempo   sql.NullInt64
	SpotMeasures       sql.NullString
}

func (q *Queries) GetPieceByID(ctx context.Context, arg GetPieceByIDParams) ([]GetPieceByIDRow, error) {
	rows, err := q.db.QueryContext(ctx, getPieceByID, arg.PieceID, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPieceByIDRow
	for rows.Next() {
		var i GetPieceByIDRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Description,
			&i.Composer,
			&i.Measures,
			&i.BeatsPerMeasure,
			&i.GoalTempo,
			&i.LastPracticed,
			&i.SpotID,
			&i.SpotName,
			&i.SpotIdx,
			&i.SpotStage,
			&i.SpotAudioPromptUrl,
			&i.SpotImagePromptUrl,
			&i.SpotNotesPrompt,
			&i.SpotTextPrompt,
			&i.SpotCurrentTempo,
			&i.SpotMeasures,
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

const getPieceWithRandomSpots = `-- name: GetPieceWithRandomSpots :many
SELECT
    pieces.id,
    pieces.title,
    pieces.description,
    pieces.composer,
    pieces.measures,
    pieces.beats_per_measure,
    pieces.goal_tempo,
    pieces.last_practiced,
    spots.id AS spot_id,
    spots.name AS spot_name,
    spots.idx AS spot_idx,
    spots.stage AS spot_stage,
    spots.audio_prompt_url AS spot_audio_prompt_url,
    spots.image_prompt_url AS spot_image_prompt_url,
    spots.notes_prompt AS spot_notes_prompt,
    spots.text_prompt AS spot_text_prompt,
    spots.current_tempo AS spot_current_tempo,
    spots.measures AS spot_measures
FROM pieces
INNER JOIN spots ON pieces.id = spots.piece_id AND spots.stage != 'repeat' AND spots.stage != 'completed'
WHERE pieces.id = ?1 AND pieces.user_id = ?2
`

type GetPieceWithRandomSpotsParams struct {
	PieceID string
	UserID  string
}

type GetPieceWithRandomSpotsRow struct {
	ID                 string
	Title              string
	Description        sql.NullString
	Composer           sql.NullString
	Measures           sql.NullInt64
	BeatsPerMeasure    sql.NullInt64
	GoalTempo          sql.NullInt64
	LastPracticed      sql.NullInt64
	SpotID             string
	SpotName           string
	SpotIdx            int64
	SpotStage          string
	SpotAudioPromptUrl string
	SpotImagePromptUrl string
	SpotNotesPrompt    string
	SpotTextPrompt     string
	SpotCurrentTempo   sql.NullInt64
	SpotMeasures       sql.NullString
}

func (q *Queries) GetPieceWithRandomSpots(ctx context.Context, arg GetPieceWithRandomSpotsParams) ([]GetPieceWithRandomSpotsRow, error) {
	rows, err := q.db.QueryContext(ctx, getPieceWithRandomSpots, arg.PieceID, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPieceWithRandomSpotsRow
	for rows.Next() {
		var i GetPieceWithRandomSpotsRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Description,
			&i.Composer,
			&i.Measures,
			&i.BeatsPerMeasure,
			&i.GoalTempo,
			&i.LastPracticed,
			&i.SpotID,
			&i.SpotName,
			&i.SpotIdx,
			&i.SpotStage,
			&i.SpotAudioPromptUrl,
			&i.SpotImagePromptUrl,
			&i.SpotNotesPrompt,
			&i.SpotTextPrompt,
			&i.SpotCurrentTempo,
			&i.SpotMeasures,
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

const getPieceWithoutSpots = `-- name: GetPieceWithoutSpots :one
SELECT id, title, description, composer, measures, beats_per_measure, goal_tempo, user_id, last_practiced FROM pieces WHERE id = ? AND user_id = ?
`

type GetPieceWithoutSpotsParams struct {
	ID     string
	UserID string
}

func (q *Queries) GetPieceWithoutSpots(ctx context.Context, arg GetPieceWithoutSpotsParams) (Piece, error) {
	row := q.db.QueryRowContext(ctx, getPieceWithoutSpots, arg.ID, arg.UserID)
	var i Piece
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Description,
		&i.Composer,
		&i.Measures,
		&i.BeatsPerMeasure,
		&i.GoalTempo,
		&i.UserID,
		&i.LastPracticed,
	)
	return i, err
}

const listAllUserPieces = `-- name: ListAllUserPieces :many
SELECT
    id,
    title,
    composer,
    last_practiced,
    (SELECT COUNT(spots.id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage == 'completed') AS completed_spots,
    (SELECT COUNT(spots.id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage != 'completed') AS active_spots
FROM pieces
WHERE user_id = ?
`

type ListAllUserPiecesRow struct {
	ID             string
	Title          string
	Composer       sql.NullString
	LastPracticed  sql.NullInt64
	CompletedSpots int64
	ActiveSpots    int64
}

func (q *Queries) ListAllUserPieces(ctx context.Context, userID string) ([]ListAllUserPiecesRow, error) {
	rows, err := q.db.QueryContext(ctx, listAllUserPieces, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListAllUserPiecesRow
	for rows.Next() {
		var i ListAllUserPiecesRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Composer,
			&i.LastPracticed,
			&i.CompletedSpots,
			&i.ActiveSpots,
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

const listPaginatedUserPieces = `-- name: ListPaginatedUserPieces :many
SELECT
    id,
    title,
    composer,
    (SELECT COUNT(spots.id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage == 'completed') AS completed_spots,
    (SELECT COUNT(spots.id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage != 'completed') AS active_spots
FROM pieces
WHERE user_id = ?
ORDER BY last_practiced DESC
LIMIT ? OFFSET ?
`

type ListPaginatedUserPiecesParams struct {
	UserID string
	Limit  int64
	Offset int64
}

type ListPaginatedUserPiecesRow struct {
	ID             string
	Title          string
	Composer       sql.NullString
	CompletedSpots int64
	ActiveSpots    int64
}

func (q *Queries) ListPaginatedUserPieces(ctx context.Context, arg ListPaginatedUserPiecesParams) ([]ListPaginatedUserPiecesRow, error) {
	rows, err := q.db.QueryContext(ctx, listPaginatedUserPieces, arg.UserID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListPaginatedUserPiecesRow
	for rows.Next() {
		var i ListPaginatedUserPiecesRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Composer,
			&i.CompletedSpots,
			&i.ActiveSpots,
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

const listRecentlyPracticedPieces = `-- name: ListRecentlyPracticedPieces :many
SELECT
    id,
    title,
    composer,
    (SELECT COUNT(spots.id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage == 'completed') AS completed_spots,
    (SELECT COUNT(spots.id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage != 'completed') AS active_spots
FROM pieces
WHERE user_id = ?
ORDER BY last_practiced DESC
LIMIT 5
`

type ListRecentlyPracticedPiecesRow struct {
	ID             string
	Title          string
	Composer       sql.NullString
	CompletedSpots int64
	ActiveSpots    int64
}

func (q *Queries) ListRecentlyPracticedPieces(ctx context.Context, userID string) ([]ListRecentlyPracticedPiecesRow, error) {
	rows, err := q.db.QueryContext(ctx, listRecentlyPracticedPieces, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListRecentlyPracticedPiecesRow
	for rows.Next() {
		var i ListRecentlyPracticedPiecesRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Composer,
			&i.CompletedSpots,
			&i.ActiveSpots,
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

const updatePiece = `-- name: UpdatePiece :one
UPDATE pieces
SET
    title = ?,
    description = ?,
    composer = ?,
    measures = ?,
    beats_per_measure = ?,
    goal_tempo = ?
WHERE id = ? AND user_id = ?
RETURNING id, title, description, composer, measures, beats_per_measure, goal_tempo, user_id, last_practiced
`

type UpdatePieceParams struct {
	Title           string
	Description     sql.NullString
	Composer        sql.NullString
	Measures        sql.NullInt64
	BeatsPerMeasure sql.NullInt64
	GoalTempo       sql.NullInt64
	ID              string
	UserID          string
}

func (q *Queries) UpdatePiece(ctx context.Context, arg UpdatePieceParams) (Piece, error) {
	row := q.db.QueryRowContext(ctx, updatePiece,
		arg.Title,
		arg.Description,
		arg.Composer,
		arg.Measures,
		arg.BeatsPerMeasure,
		arg.GoalTempo,
		arg.ID,
		arg.UserID,
	)
	var i Piece
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Description,
		&i.Composer,
		&i.Measures,
		&i.BeatsPerMeasure,
		&i.GoalTempo,
		&i.UserID,
		&i.LastPracticed,
	)
	return i, err
}
