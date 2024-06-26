// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: pieces.sql

package db

import (
	"context"
	"database/sql"
)

const checkPieceForRandomSpots = `-- name: CheckPieceForRandomSpots :one
SELECT COUNT(*) FROM spots WHERE piece_id = ? AND stage = 'random'
`

func (q *Queries) CheckPieceForRandomSpots(ctx context.Context, pieceID string) (int64, error) {
	row := q.db.QueryRowContext(ctx, checkPieceForRandomSpots, pieceID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

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
    user_id,
    key_id,
    mode_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id, title, description, composer, measures, beats_per_measure, goal_tempo, user_id, last_practiced, stage, key_id, mode_id
`

type CreatePieceParams struct {
	ID              string         `json:"id"`
	Title           string         `json:"title"`
	Description     sql.NullString `json:"description"`
	Composer        sql.NullString `json:"composer"`
	Measures        sql.NullInt64  `json:"measures"`
	BeatsPerMeasure sql.NullInt64  `json:"beatsPerMeasure"`
	GoalTempo       sql.NullInt64  `json:"goalTempo"`
	UserID          string         `json:"userId"`
	KeyID           sql.NullInt64  `json:"keyId"`
	ModeID          sql.NullInt64  `json:"modeId"`
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
		arg.KeyID,
		arg.ModeID,
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
		&i.Stage,
		&i.KeyID,
		&i.ModeID,
	)
	return i, err
}

const deletePiece = `-- name: DeletePiece :exec
DELETE FROM pieces
WHERE id = ? AND user_id = ?
`

type DeletePieceParams struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
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
    pieces.stage,
    scale_keys.name as key_name,
    scale_modes.name as mode,
    spots.id AS spot_id,
    spots.name AS spot_name,
    spots.stage AS spot_stage,
    spots.audio_prompt_url AS spot_audio_prompt_url,
    spots.image_prompt_url AS spot_image_prompt_url,
    spots.notes_prompt AS spot_notes_prompt,
    spots.text_prompt AS spot_text_prompt,
    spots.current_tempo AS spot_current_tempo,
    spots.measures AS spot_measures,
    spots.last_practiced AS spot_last_practiced
FROM pieces
LEFT JOIN spots ON pieces.id = spots.piece_id
LEFT JOIN scale_keys ON pieces.key_id = scale_keys.id
LEFT JOIN scale_modes ON pieces.mode_id = scale_modes.id
WHERE pieces.id = ?1 AND pieces.user_id = ?2
ORDER BY spot_last_practiced DESC
`

type GetPieceByIDParams struct {
	PieceID string `json:"pieceId"`
	UserID  string `json:"userId"`
}

type GetPieceByIDRow struct {
	ID                 string         `json:"id"`
	Title              string         `json:"title"`
	Description        sql.NullString `json:"description"`
	Composer           sql.NullString `json:"composer"`
	Measures           sql.NullInt64  `json:"measures"`
	BeatsPerMeasure    sql.NullInt64  `json:"beatsPerMeasure"`
	GoalTempo          sql.NullInt64  `json:"goalTempo"`
	LastPracticed      sql.NullInt64  `json:"lastPracticed"`
	Stage              string         `json:"stage"`
	KeyName            sql.NullString `json:"keyName"`
	Mode               sql.NullString `json:"mode"`
	SpotID             sql.NullString `json:"spotId"`
	SpotName           sql.NullString `json:"spotName"`
	SpotStage          sql.NullString `json:"spotStage"`
	SpotAudioPromptUrl sql.NullString `json:"spotAudioPromptUrl"`
	SpotImagePromptUrl sql.NullString `json:"spotImagePromptUrl"`
	SpotNotesPrompt    sql.NullString `json:"spotNotesPrompt"`
	SpotTextPrompt     sql.NullString `json:"spotTextPrompt"`
	SpotCurrentTempo   sql.NullInt64  `json:"spotCurrentTempo"`
	SpotMeasures       sql.NullString `json:"spotMeasures"`
	SpotLastPracticed  sql.NullInt64  `json:"spotLastPracticed"`
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
			&i.Stage,
			&i.KeyName,
			&i.Mode,
			&i.SpotID,
			&i.SpotName,
			&i.SpotStage,
			&i.SpotAudioPromptUrl,
			&i.SpotImagePromptUrl,
			&i.SpotNotesPrompt,
			&i.SpotTextPrompt,
			&i.SpotCurrentTempo,
			&i.SpotMeasures,
			&i.SpotLastPracticed,
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

const getPieceForPlan = `-- name: GetPieceForPlan :many
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
    spots.stage AS spot_stage,
    spots.last_practiced AS spot_last_practiced,
    spots.skip_days AS spot_skip_days
FROM pieces
LEFT JOIN spots ON pieces.id = spots.piece_id
WHERE pieces.id = ?1 AND pieces.user_id = ?2
`

type GetPieceForPlanParams struct {
	PieceID string `json:"pieceId"`
	UserID  string `json:"userId"`
}

type GetPieceForPlanRow struct {
	ID                string         `json:"id"`
	Title             string         `json:"title"`
	Description       sql.NullString `json:"description"`
	Composer          sql.NullString `json:"composer"`
	Measures          sql.NullInt64  `json:"measures"`
	BeatsPerMeasure   sql.NullInt64  `json:"beatsPerMeasure"`
	GoalTempo         sql.NullInt64  `json:"goalTempo"`
	LastPracticed     sql.NullInt64  `json:"lastPracticed"`
	SpotID            sql.NullString `json:"spotId"`
	SpotName          sql.NullString `json:"spotName"`
	SpotStage         sql.NullString `json:"spotStage"`
	SpotLastPracticed sql.NullInt64  `json:"spotLastPracticed"`
	SpotSkipDays      sql.NullInt64  `json:"spotSkipDays"`
}

func (q *Queries) GetPieceForPlan(ctx context.Context, arg GetPieceForPlanParams) ([]GetPieceForPlanRow, error) {
	rows, err := q.db.QueryContext(ctx, getPieceForPlan, arg.PieceID, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPieceForPlanRow
	for rows.Next() {
		var i GetPieceForPlanRow
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
			&i.SpotStage,
			&i.SpotLastPracticed,
			&i.SpotSkipDays,
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

const getPieceWithIncompleteSpots = `-- name: GetPieceWithIncompleteSpots :many
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
    spots.stage AS spot_stage,
    spots.audio_prompt_url AS spot_audio_prompt_url,
    spots.image_prompt_url AS spot_image_prompt_url,
    spots.notes_prompt AS spot_notes_prompt,
    spots.text_prompt AS spot_text_prompt,
    spots.current_tempo AS spot_current_tempo,
    spots.measures AS spot_measures
FROM pieces
INNER JOIN spots ON pieces.id = spots.piece_id AND spots.stage != 'completed'
WHERE pieces.id = ?1 AND pieces.user_id = ?2
`

type GetPieceWithIncompleteSpotsParams struct {
	PieceID string `json:"pieceId"`
	UserID  string `json:"userId"`
}

type GetPieceWithIncompleteSpotsRow struct {
	ID                 string         `json:"id"`
	Title              string         `json:"title"`
	Description        sql.NullString `json:"description"`
	Composer           sql.NullString `json:"composer"`
	Measures           sql.NullInt64  `json:"measures"`
	BeatsPerMeasure    sql.NullInt64  `json:"beatsPerMeasure"`
	GoalTempo          sql.NullInt64  `json:"goalTempo"`
	LastPracticed      sql.NullInt64  `json:"lastPracticed"`
	SpotID             string         `json:"spotId"`
	SpotName           string         `json:"spotName"`
	SpotStage          string         `json:"spotStage"`
	SpotAudioPromptUrl string         `json:"spotAudioPromptUrl"`
	SpotImagePromptUrl string         `json:"spotImagePromptUrl"`
	SpotNotesPrompt    string         `json:"spotNotesPrompt"`
	SpotTextPrompt     string         `json:"spotTextPrompt"`
	SpotCurrentTempo   sql.NullInt64  `json:"spotCurrentTempo"`
	SpotMeasures       sql.NullString `json:"spotMeasures"`
}

func (q *Queries) GetPieceWithIncompleteSpots(ctx context.Context, arg GetPieceWithIncompleteSpotsParams) ([]GetPieceWithIncompleteSpotsRow, error) {
	rows, err := q.db.QueryContext(ctx, getPieceWithIncompleteSpots, arg.PieceID, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPieceWithIncompleteSpotsRow
	for rows.Next() {
		var i GetPieceWithIncompleteSpotsRow
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
    spots.stage AS spot_stage,
    spots.audio_prompt_url AS spot_audio_prompt_url,
    spots.image_prompt_url AS spot_image_prompt_url,
    spots.notes_prompt AS spot_notes_prompt,
    spots.text_prompt AS spot_text_prompt,
    spots.current_tempo AS spot_current_tempo,
    spots.stage_started AS spot_stage_started,
    spots.measures AS spot_measures
FROM pieces
INNER JOIN spots ON pieces.id = spots.piece_id AND spots.stage = 'random'
WHERE pieces.id = ?1 AND pieces.user_id = ?2
`

type GetPieceWithRandomSpotsParams struct {
	PieceID string `json:"pieceId"`
	UserID  string `json:"userId"`
}

type GetPieceWithRandomSpotsRow struct {
	ID                 string         `json:"id"`
	Title              string         `json:"title"`
	Description        sql.NullString `json:"description"`
	Composer           sql.NullString `json:"composer"`
	Measures           sql.NullInt64  `json:"measures"`
	BeatsPerMeasure    sql.NullInt64  `json:"beatsPerMeasure"`
	GoalTempo          sql.NullInt64  `json:"goalTempo"`
	LastPracticed      sql.NullInt64  `json:"lastPracticed"`
	SpotID             string         `json:"spotId"`
	SpotName           string         `json:"spotName"`
	SpotStage          string         `json:"spotStage"`
	SpotAudioPromptUrl string         `json:"spotAudioPromptUrl"`
	SpotImagePromptUrl string         `json:"spotImagePromptUrl"`
	SpotNotesPrompt    string         `json:"spotNotesPrompt"`
	SpotTextPrompt     string         `json:"spotTextPrompt"`
	SpotCurrentTempo   sql.NullInt64  `json:"spotCurrentTempo"`
	SpotStageStarted   sql.NullInt64  `json:"spotStageStarted"`
	SpotMeasures       sql.NullString `json:"spotMeasures"`
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
			&i.SpotStage,
			&i.SpotAudioPromptUrl,
			&i.SpotImagePromptUrl,
			&i.SpotNotesPrompt,
			&i.SpotTextPrompt,
			&i.SpotCurrentTempo,
			&i.SpotStageStarted,
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
SELECT id, title, description, composer, measures, beats_per_measure, goal_tempo, user_id, last_practiced, stage, key_id, mode_id FROM pieces WHERE id = ? AND user_id = ?
`

type GetPieceWithoutSpotsParams struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
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
		&i.Stage,
		&i.KeyID,
		&i.ModeID,
	)
	return i, err
}

const listActivePiecesWithCompletedSpotsForPlan = `-- name: ListActivePiecesWithCompletedSpotsForPlan :many
SELECT
    id,
    title,
    composer,
    last_practiced,
    (SELECT COUNT(spots.id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage == 'completed') AS completed_spot_count
FROM pieces
WHERE user_id = ?1
AND stage = 'active'
AND completed_spot_count > 5
AND id NOT IN (SELECT practice_plan_pieces.piece_id FROM practice_plan_pieces WHERE practice_plan_pieces.practice_plan_id = ?2)
ORDER BY last_practiced DESC
`

type ListActivePiecesWithCompletedSpotsForPlanParams struct {
	UserID string `json:"userId"`
	PlanID string `json:"planId"`
}

type ListActivePiecesWithCompletedSpotsForPlanRow struct {
	ID                 string         `json:"id"`
	Title              string         `json:"title"`
	Composer           sql.NullString `json:"composer"`
	LastPracticed      sql.NullInt64  `json:"lastPracticed"`
	CompletedSpotCount int64          `json:"completedSpotCount"`
}

func (q *Queries) ListActivePiecesWithCompletedSpotsForPlan(ctx context.Context, arg ListActivePiecesWithCompletedSpotsForPlanParams) ([]ListActivePiecesWithCompletedSpotsForPlanRow, error) {
	rows, err := q.db.QueryContext(ctx, listActivePiecesWithCompletedSpotsForPlan, arg.UserID, arg.PlanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListActivePiecesWithCompletedSpotsForPlanRow
	for rows.Next() {
		var i ListActivePiecesWithCompletedSpotsForPlanRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Composer,
			&i.LastPracticed,
			&i.CompletedSpotCount,
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

const listActiveUserPieces = `-- name: ListActiveUserPieces :many
SELECT
    id,
    title,
    composer,
    last_practiced,
    (SELECT COUNT(spots.id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage == 'completed') AS completed_spots,
    (SELECT COUNT(spots.id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage != 'completed') AS active_spots
FROM pieces
WHERE user_id = ? AND stage = 'active'
ORDER BY last_practiced DESC
`

type ListActiveUserPiecesRow struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Composer       sql.NullString `json:"composer"`
	LastPracticed  sql.NullInt64  `json:"lastPracticed"`
	CompletedSpots int64          `json:"completedSpots"`
	ActiveSpots    int64          `json:"activeSpots"`
}

func (q *Queries) ListActiveUserPieces(ctx context.Context, userID string) ([]ListActiveUserPiecesRow, error) {
	rows, err := q.db.QueryContext(ctx, listActiveUserPieces, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListActiveUserPiecesRow
	for rows.Next() {
		var i ListActiveUserPiecesRow
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
ORDER BY last_practiced DESC
`

type ListAllUserPiecesRow struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Composer       sql.NullString `json:"composer"`
	LastPracticed  sql.NullInt64  `json:"lastPracticed"`
	CompletedSpots int64          `json:"completedSpots"`
	ActiveSpots    int64          `json:"activeSpots"`
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
	UserID string `json:"userId"`
	Limit  int64  `json:"limit"`
	Offset int64  `json:"offset"`
}

type ListPaginatedUserPiecesRow struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Composer       sql.NullString `json:"composer"`
	CompletedSpots int64          `json:"completedSpots"`
	ActiveSpots    int64          `json:"activeSpots"`
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

const listPiecesWithNewSpotsForPlan = `-- name: ListPiecesWithNewSpotsForPlan :many
SELECT
    id,
    title,
    composer,
    (SELECT COUNT(spots.id)
        FROM spots
        WHERE spots.piece_id = pieces.id
        AND spots.stage == 'repeat'
    AND spots.id NOT IN (SELECT practice_plan_spots.spot_id
            FROM practice_plan_spots
            WHERE practice_plan_spots.practice_plan_id = ?1)) AS new_spots_count
FROM pieces
WHERE user_id = ?2 AND stage = 'active' AND new_spots_count > 0
ORDER BY pieces.last_practiced DESC
`

type ListPiecesWithNewSpotsForPlanParams struct {
	PlanID string `json:"planId"`
	UserID string `json:"userId"`
}

type ListPiecesWithNewSpotsForPlanRow struct {
	ID            string         `json:"id"`
	Title         string         `json:"title"`
	Composer      sql.NullString `json:"composer"`
	NewSpotsCount int64          `json:"newSpotsCount"`
}

func (q *Queries) ListPiecesWithNewSpotsForPlan(ctx context.Context, arg ListPiecesWithNewSpotsForPlanParams) ([]ListPiecesWithNewSpotsForPlanRow, error) {
	rows, err := q.db.QueryContext(ctx, listPiecesWithNewSpotsForPlan, arg.PlanID, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListPiecesWithNewSpotsForPlanRow
	for rows.Next() {
		var i ListPiecesWithNewSpotsForPlanRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Composer,
			&i.NewSpotsCount,
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

const listRandomSpotPiecesForPlan = `-- name: ListRandomSpotPiecesForPlan :many
SELECT
    pieces.id, pieces.title, pieces.description, pieces.composer, pieces.measures, pieces.beats_per_measure, pieces.goal_tempo, pieces.user_id, pieces.last_practiced, pieces.stage, pieces.key_id, pieces.mode_id,
    (SELECT COUNT(spots.id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage == 'random') AS random_spot_count
FROM pieces
WHERE user_id = ?1
AND random_spot_count > 0
AND pieces.id NOT IN (SELECT practice_plan_pieces.piece_id FROM practice_plan_pieces WHERE practice_plan_pieces.practice_plan_id = ?2)
ORDER BY last_practiced DESC
`

type ListRandomSpotPiecesForPlanParams struct {
	UserID string `json:"userId"`
	PlanID string `json:"planId"`
}

type ListRandomSpotPiecesForPlanRow struct {
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
	KeyID           sql.NullInt64  `json:"keyId"`
	ModeID          sql.NullInt64  `json:"modeId"`
	RandomSpotCount int64          `json:"randomSpotCount"`
}

func (q *Queries) ListRandomSpotPiecesForPlan(ctx context.Context, arg ListRandomSpotPiecesForPlanParams) ([]ListRandomSpotPiecesForPlanRow, error) {
	rows, err := q.db.QueryContext(ctx, listRandomSpotPiecesForPlan, arg.UserID, arg.PlanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListRandomSpotPiecesForPlanRow
	for rows.Next() {
		var i ListRandomSpotPiecesForPlanRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Description,
			&i.Composer,
			&i.Measures,
			&i.BeatsPerMeasure,
			&i.GoalTempo,
			&i.UserID,
			&i.LastPracticed,
			&i.Stage,
			&i.KeyID,
			&i.ModeID,
			&i.RandomSpotCount,
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
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Composer       sql.NullString `json:"composer"`
	CompletedSpots int64          `json:"completedSpots"`
	ActiveSpots    int64          `json:"activeSpots"`
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
    goal_tempo = ?,
    stage = ?
WHERE id = ? AND user_id = ?
RETURNING id, title, description, composer, measures, beats_per_measure, goal_tempo, user_id, last_practiced, stage, key_id, mode_id
`

type UpdatePieceParams struct {
	Title           string         `json:"title"`
	Description     sql.NullString `json:"description"`
	Composer        sql.NullString `json:"composer"`
	Measures        sql.NullInt64  `json:"measures"`
	BeatsPerMeasure sql.NullInt64  `json:"beatsPerMeasure"`
	GoalTempo       sql.NullInt64  `json:"goalTempo"`
	Stage           string         `json:"stage"`
	ID              string         `json:"id"`
	UserID          string         `json:"userId"`
}

func (q *Queries) UpdatePiece(ctx context.Context, arg UpdatePieceParams) (Piece, error) {
	row := q.db.QueryRowContext(ctx, updatePiece,
		arg.Title,
		arg.Description,
		arg.Composer,
		arg.Measures,
		arg.BeatsPerMeasure,
		arg.GoalTempo,
		arg.Stage,
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
		&i.Stage,
		&i.KeyID,
		&i.ModeID,
	)
	return i, err
}

const updatePiecePracticed = `-- name: UpdatePiecePracticed :exec
UPDATE pieces
SET last_practiced = unixepoch('now')
WHERE pieces.id = ?1 AND user_id = ?2
`

type UpdatePiecePracticedParams struct {
	PieceID string `json:"pieceId"`
	UserID  string `json:"userId"`
}

func (q *Queries) UpdatePiecePracticed(ctx context.Context, arg UpdatePiecePracticedParams) error {
	_, err := q.db.ExecContext(ctx, updatePiecePracticed, arg.PieceID, arg.UserID)
	return err
}
