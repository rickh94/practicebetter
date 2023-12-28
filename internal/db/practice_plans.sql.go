// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0
// source: practice_plans.sql

package db

import (
	"context"
	"database/sql"
)

const addPracticeSessionToPlan = `-- name: AddPracticeSessionToPlan :exec
UPDATE practice_plans
SET practice_session_id = ?
WHERE id = ? AND user_id = ?
`

type AddPracticeSessionToPlanParams struct {
	PracticeSessionID sql.NullString `json:"practiceSessionId"`
	ID                string         `json:"id"`
	UserID            string         `json:"userId"`
}

func (q *Queries) AddPracticeSessionToPlan(ctx context.Context, arg AddPracticeSessionToPlanParams) error {
	_, err := q.db.ExecContext(ctx, addPracticeSessionToPlan, arg.PracticeSessionID, arg.ID, arg.UserID)
	return err
}

const completePracticePlanPiece = `-- name: CompletePracticePlanPiece :exec
UPDATE practice_plan_pieces
SET completed = true
WHERE practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = ? AND practice_plans.user_id = ?) AND piece_id = ? AND practice_type = ?
`

type CompletePracticePlanPieceParams struct {
	PlanID       string `json:"planId"`
	UserID       string `json:"userId"`
	PieceID      string `json:"pieceId"`
	PracticeType string `json:"practiceType"`
}

func (q *Queries) CompletePracticePlanPiece(ctx context.Context, arg CompletePracticePlanPieceParams) error {
	_, err := q.db.ExecContext(ctx, completePracticePlanPiece,
		arg.PlanID,
		arg.UserID,
		arg.PieceID,
		arg.PracticeType,
	)
	return err
}

const completePracticePlanSpot = `-- name: CompletePracticePlanSpot :exec
UPDATE practice_plan_spots
SET completed = true
WHERE practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = ? AND practice_plans.user_id = ?) AND spot_id = ?
`

type CompletePracticePlanSpotParams struct {
	PlanID string `json:"planId"`
	UserID string `json:"userId"`
	SpotID string `json:"spotId"`
}

func (q *Queries) CompletePracticePlanSpot(ctx context.Context, arg CompletePracticePlanSpotParams) error {
	_, err := q.db.ExecContext(ctx, completePracticePlanSpot, arg.PlanID, arg.UserID, arg.SpotID)
	return err
}

const countUserPracticePlans = `-- name: CountUserPracticePlans :one
SELECT COUNT(*) FROM practice_plans WHERE user_id = ?
`

func (q *Queries) CountUserPracticePlans(ctx context.Context, userID string) (int64, error) {
	row := q.db.QueryRowContext(ctx, countUserPracticePlans, userID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createPracticePlan = `-- name: CreatePracticePlan :one
INSERT INTO practice_plans (
    id,
    user_id,
    intensity,
    practice_session_id,
    date
) VALUES (?, ?, ?, ?, unixepoch('now'))
RETURNING id, user_id, intensity, date, completed, practice_session_id
`

type CreatePracticePlanParams struct {
	ID                string         `json:"id"`
	UserID            string         `json:"userId"`
	Intensity         string         `json:"intensity"`
	PracticeSessionID sql.NullString `json:"practiceSessionId"`
}

func (q *Queries) CreatePracticePlan(ctx context.Context, arg CreatePracticePlanParams) (PracticePlan, error) {
	row := q.db.QueryRowContext(ctx, createPracticePlan,
		arg.ID,
		arg.UserID,
		arg.Intensity,
		arg.PracticeSessionID,
	)
	var i PracticePlan
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Intensity,
		&i.Date,
		&i.Completed,
		&i.PracticeSessionID,
	)
	return i, err
}

const createPracticePlanPiece = `-- name: CreatePracticePlanPiece :one
INSERT INTO practice_plan_pieces (
    practice_plan_id,
    piece_id,
    practice_type
) VALUES (?, ?, ?)
RETURNING practice_plan_id, piece_id, practice_type, completed
`

type CreatePracticePlanPieceParams struct {
	PracticePlanID string `json:"practicePlanId"`
	PieceID        string `json:"pieceId"`
	PracticeType   string `json:"practiceType"`
}

func (q *Queries) CreatePracticePlanPiece(ctx context.Context, arg CreatePracticePlanPieceParams) (PracticePlanPiece, error) {
	row := q.db.QueryRowContext(ctx, createPracticePlanPiece, arg.PracticePlanID, arg.PieceID, arg.PracticeType)
	var i PracticePlanPiece
	err := row.Scan(
		&i.PracticePlanID,
		&i.PieceID,
		&i.PracticeType,
		&i.Completed,
	)
	return i, err
}

const createPracticePlanSpot = `-- name: CreatePracticePlanSpot :one
INSERT INTO practice_plan_spots (
    practice_plan_id,
    spot_id,
    practice_type
) VALUES (?, ?, ?)
RETURNING practice_plan_id, spot_id, practice_type, completed
`

type CreatePracticePlanSpotParams struct {
	PracticePlanID string `json:"practicePlanId"`
	SpotID         string `json:"spotId"`
	PracticeType   string `json:"practiceType"`
}

func (q *Queries) CreatePracticePlanSpot(ctx context.Context, arg CreatePracticePlanSpotParams) (PracticePlanSpot, error) {
	row := q.db.QueryRowContext(ctx, createPracticePlanSpot, arg.PracticePlanID, arg.SpotID, arg.PracticeType)
	var i PracticePlanSpot
	err := row.Scan(
		&i.PracticePlanID,
		&i.SpotID,
		&i.PracticeType,
		&i.Completed,
	)
	return i, err
}

const deletePracticePlan = `-- name: DeletePracticePlan :exec
DELETE FROM practice_plans
WHERE id = ? AND user_id = ?
`

type DeletePracticePlanParams struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
}

func (q *Queries) DeletePracticePlan(ctx context.Context, arg DeletePracticePlanParams) error {
	_, err := q.db.ExecContext(ctx, deletePracticePlan, arg.ID, arg.UserID)
	return err
}

const getPracticePlan = `-- name: GetPracticePlan :one
SELECT id, user_id, intensity, date, completed, practice_session_id
FROM practice_plans
WHERE id = ? AND user_id = ?
`

type GetPracticePlanParams struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
}

func (q *Queries) GetPracticePlan(ctx context.Context, arg GetPracticePlanParams) (PracticePlan, error) {
	row := q.db.QueryRowContext(ctx, getPracticePlan, arg.ID, arg.UserID)
	var i PracticePlan
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Intensity,
		&i.Date,
		&i.Completed,
		&i.PracticeSessionID,
	)
	return i, err
}

const getPracticePlanInterleaveDaysSpots = `-- name: GetPracticePlanInterleaveDaysSpots :many
SELECT practice_plan_spots.practice_plan_id, practice_plan_spots.spot_id, practice_plan_spots.practice_type, practice_plan_spots.completed,
    spots.name AS spot_name,
    spots.measures AS spot_measures,
    spots.piece_id AS spot_piece_id,
    spots.stage AS spot_stage,
    (SELECT pieces.title FROM pieces WHERE pieces.id = spots.piece_id LIMIT 1) AS spot_piece_title
FROM practice_plan_spots
LEFT JOIN spots ON practice_plan_spots.spot_id = spots.id
WHERE practice_plan_spots.practice_type = 'interleave_days' AND practice_plan_spots.practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = ?1 AND practice_plans.user_id = ?2)
`

type GetPracticePlanInterleaveDaysSpotsParams struct {
	PlanID string `json:"planId"`
	UserID string `json:"userId"`
}

type GetPracticePlanInterleaveDaysSpotsRow struct {
	PracticePlanID string         `json:"practicePlanId"`
	SpotID         string         `json:"spotId"`
	PracticeType   string         `json:"practiceType"`
	Completed      bool           `json:"completed"`
	SpotName       sql.NullString `json:"spotName"`
	SpotMeasures   sql.NullString `json:"spotMeasures"`
	SpotPieceID    sql.NullString `json:"spotPieceId"`
	SpotStage      sql.NullString `json:"spotStage"`
	SpotPieceTitle string         `json:"spotPieceTitle"`
}

func (q *Queries) GetPracticePlanInterleaveDaysSpots(ctx context.Context, arg GetPracticePlanInterleaveDaysSpotsParams) ([]GetPracticePlanInterleaveDaysSpotsRow, error) {
	rows, err := q.db.QueryContext(ctx, getPracticePlanInterleaveDaysSpots, arg.PlanID, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPracticePlanInterleaveDaysSpotsRow
	for rows.Next() {
		var i GetPracticePlanInterleaveDaysSpotsRow
		if err := rows.Scan(
			&i.PracticePlanID,
			&i.SpotID,
			&i.PracticeType,
			&i.Completed,
			&i.SpotName,
			&i.SpotMeasures,
			&i.SpotPieceID,
			&i.SpotStage,
			&i.SpotPieceTitle,
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

const getPracticePlanInterleaveSpots = `-- name: GetPracticePlanInterleaveSpots :many
SELECT practice_plan_spots.practice_plan_id, practice_plan_spots.spot_id, practice_plan_spots.practice_type, practice_plan_spots.completed,
    spots.name AS spot_name,
    spots.measures AS spot_measures,
    spots.piece_id AS spot_piece_id,
    spots.stage AS spot_stage,
    spots.stage_started AS spot_stage_started,
    (SELECT pieces.title FROM pieces WHERE pieces.id = spots.piece_id LIMIT 1) AS spot_piece_title
FROM practice_plan_spots
LEFT JOIN spots ON practice_plan_spots.spot_id = spots.id
WHERE practice_plan_spots.practice_type = 'interleave' AND practice_plan_spots.practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = ?1 AND practice_plans.user_id = ?2)
`

type GetPracticePlanInterleaveSpotsParams struct {
	PlanID string `json:"planId"`
	UserID string `json:"userId"`
}

type GetPracticePlanInterleaveSpotsRow struct {
	PracticePlanID   string         `json:"practicePlanId"`
	SpotID           string         `json:"spotId"`
	PracticeType     string         `json:"practiceType"`
	Completed        bool           `json:"completed"`
	SpotName         sql.NullString `json:"spotName"`
	SpotMeasures     sql.NullString `json:"spotMeasures"`
	SpotPieceID      sql.NullString `json:"spotPieceId"`
	SpotStage        sql.NullString `json:"spotStage"`
	SpotStageStarted sql.NullInt64  `json:"spotStageStarted"`
	SpotPieceTitle   string         `json:"spotPieceTitle"`
}

func (q *Queries) GetPracticePlanInterleaveSpots(ctx context.Context, arg GetPracticePlanInterleaveSpotsParams) ([]GetPracticePlanInterleaveSpotsRow, error) {
	rows, err := q.db.QueryContext(ctx, getPracticePlanInterleaveSpots, arg.PlanID, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPracticePlanInterleaveSpotsRow
	for rows.Next() {
		var i GetPracticePlanInterleaveSpotsRow
		if err := rows.Scan(
			&i.PracticePlanID,
			&i.SpotID,
			&i.PracticeType,
			&i.Completed,
			&i.SpotName,
			&i.SpotMeasures,
			&i.SpotPieceID,
			&i.SpotStage,
			&i.SpotStageStarted,
			&i.SpotPieceTitle,
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

const getPracticePlanWithPieces = `-- name: GetPracticePlanWithPieces :many
SELECT
    practice_plans.id, practice_plans.user_id, practice_plans.intensity, practice_plans.date, practice_plans.completed, practice_plans.practice_session_id,
    practice_plan_pieces.practice_type as piece_practice_type,
    practice_plan_pieces.completed AS piece_completed,
    pieces.title AS piece_title,
    pieces.id AS piece_id,
    pieces.composer AS piece_composer,
    (SELECT COUNT(id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage != 'completed') AS piece_active_spots,
    (SELECT COUNT(id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage == 'completed') AS piece_completed_spots
FROM practice_plans
INNER JOIN practice_plan_pieces ON practice_plans.id = practice_plan_pieces.practice_plan_id
LEFT JOIN pieces ON practice_plan_pieces.piece_id = pieces.id
WHERE practice_plans.id = ? AND practice_plans.user_id = ?
`

type GetPracticePlanWithPiecesParams struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
}

type GetPracticePlanWithPiecesRow struct {
	ID                  string         `json:"id"`
	UserID              string         `json:"userId"`
	Intensity           string         `json:"intensity"`
	Date                int64          `json:"date"`
	Completed           bool           `json:"completed"`
	PracticeSessionID   sql.NullString `json:"practiceSessionId"`
	PiecePracticeType   string         `json:"piecePracticeType"`
	PieceCompleted      bool           `json:"pieceCompleted"`
	PieceTitle          sql.NullString `json:"pieceTitle"`
	PieceID             sql.NullString `json:"pieceId"`
	PieceComposer       sql.NullString `json:"pieceComposer"`
	PieceActiveSpots    int64          `json:"pieceActiveSpots"`
	PieceCompletedSpots int64          `json:"pieceCompletedSpots"`
}

func (q *Queries) GetPracticePlanWithPieces(ctx context.Context, arg GetPracticePlanWithPiecesParams) ([]GetPracticePlanWithPiecesRow, error) {
	rows, err := q.db.QueryContext(ctx, getPracticePlanWithPieces, arg.ID, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPracticePlanWithPiecesRow
	for rows.Next() {
		var i GetPracticePlanWithPiecesRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Intensity,
			&i.Date,
			&i.Completed,
			&i.PracticeSessionID,
			&i.PiecePracticeType,
			&i.PieceCompleted,
			&i.PieceTitle,
			&i.PieceID,
			&i.PieceComposer,
			&i.PieceActiveSpots,
			&i.PieceCompletedSpots,
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

const getPracticePlanWithSpots = `-- name: GetPracticePlanWithSpots :many
SELECT
    practice_plans.id, practice_plans.user_id, practice_plans.intensity, practice_plans.date, practice_plans.completed, practice_plans.practice_session_id,
    practice_plan_spots.practice_type as spot_practice_type,
    practice_plan_spots.completed as spot_completed,
    spots.name AS spot_name,
    spots.id AS spot_id,
    spots.measures AS spot_measures,
    spots.piece_id AS spot_piece_id,
    spots.stage AS spot_stage,
    (SELECT pieces.title FROM pieces WHERE pieces.id = spots.piece_id LIMIT 1) AS spot_piece_title
FROM practice_plans
INNER JOIN practice_plan_spots ON practice_plans.id = practice_plan_spots.practice_plan_id
LEFT JOIN spots ON practice_plan_spots.spot_id = spots.id
WHERE practice_plans.id = ? AND practice_plans.user_id = ?
`

type GetPracticePlanWithSpotsParams struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
}

type GetPracticePlanWithSpotsRow struct {
	ID                string         `json:"id"`
	UserID            string         `json:"userId"`
	Intensity         string         `json:"intensity"`
	Date              int64          `json:"date"`
	Completed         bool           `json:"completed"`
	PracticeSessionID sql.NullString `json:"practiceSessionId"`
	SpotPracticeType  string         `json:"spotPracticeType"`
	SpotCompleted     bool           `json:"spotCompleted"`
	SpotName          sql.NullString `json:"spotName"`
	SpotID            sql.NullString `json:"spotId"`
	SpotMeasures      sql.NullString `json:"spotMeasures"`
	SpotPieceID       sql.NullString `json:"spotPieceId"`
	SpotStage         sql.NullString `json:"spotStage"`
	SpotPieceTitle    string         `json:"spotPieceTitle"`
}

func (q *Queries) GetPracticePlanWithSpots(ctx context.Context, arg GetPracticePlanWithSpotsParams) ([]GetPracticePlanWithSpotsRow, error) {
	rows, err := q.db.QueryContext(ctx, getPracticePlanWithSpots, arg.ID, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPracticePlanWithSpotsRow
	for rows.Next() {
		var i GetPracticePlanWithSpotsRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Intensity,
			&i.Date,
			&i.Completed,
			&i.PracticeSessionID,
			&i.SpotPracticeType,
			&i.SpotCompleted,
			&i.SpotName,
			&i.SpotID,
			&i.SpotMeasures,
			&i.SpotPieceID,
			&i.SpotStage,
			&i.SpotPieceTitle,
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

const getPracticePlanWithTodo = `-- name: GetPracticePlanWithTodo :one
SELECT
    practice_plans.id, practice_plans.user_id, practice_plans.intensity, practice_plans.date, practice_plans.completed, practice_plans.practice_session_id,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND NOT practice_plan_spots.completed) AS incomplete_spots_count,
    (SELECT COUNT(*) FROM practice_plan_pieces WHERE practice_plan_pieces.practice_plan_id = practice_plans.id AND NOT practice_plan_pieces.completed) AS incomplete_pieces_count,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND practice_plan_spots.completed) AS completed_spots_count,
    (SELECT COUNT(*) FROM practice_plan_pieces WHERE practice_plan_pieces.practice_plan_id = practice_plans.id AND practice_plan_pieces.completed) AS completed_pieces_count
FROM practice_plans
WHERE practice_plans.id = ? AND practice_plans.user_id = ?
`

type GetPracticePlanWithTodoParams struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
}

type GetPracticePlanWithTodoRow struct {
	ID                    string         `json:"id"`
	UserID                string         `json:"userId"`
	Intensity             string         `json:"intensity"`
	Date                  int64          `json:"date"`
	Completed             bool           `json:"completed"`
	PracticeSessionID     sql.NullString `json:"practiceSessionId"`
	IncompleteSpotsCount  int64          `json:"incompleteSpotsCount"`
	IncompletePiecesCount int64          `json:"incompletePiecesCount"`
	CompletedSpotsCount   int64          `json:"completedSpotsCount"`
	CompletedPiecesCount  int64          `json:"completedPiecesCount"`
}

func (q *Queries) GetPracticePlanWithTodo(ctx context.Context, arg GetPracticePlanWithTodoParams) (GetPracticePlanWithTodoRow, error) {
	row := q.db.QueryRowContext(ctx, getPracticePlanWithTodo, arg.ID, arg.UserID)
	var i GetPracticePlanWithTodoRow
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Intensity,
		&i.Date,
		&i.Completed,
		&i.PracticeSessionID,
		&i.IncompleteSpotsCount,
		&i.IncompletePiecesCount,
		&i.CompletedSpotsCount,
		&i.CompletedPiecesCount,
	)
	return i, err
}

const listPaginatedPracticePlans = `-- name: ListPaginatedPracticePlans :many
SELECT
    practice_plans.id, practice_plans.user_id, practice_plans.intensity, practice_plans.date, practice_plans.completed, practice_plans.practice_session_id,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND practice_plan_spots.practice_type = 'new') AS new_spots_count,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND practice_plan_spots.practice_type = 'extra_repeat') AS repeat_spots_count,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND practice_plan_spots.practice_type = 'interleave') AS interleave_spots_count,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND practice_plan_spots.practice_type = 'interleave_days') AS interleave_days_spots_count,
    (SELECT COUNT(*) FROM practice_plan_pieces WHERE practice_plan_pieces.practice_plan_id = practice_plans.id) AS pieces_count
FROM practice_plans
WHERE practice_plans.user_id = ?
ORDER BY practice_plans.date DESC
LIMIT ? OFFSET ?
`

type ListPaginatedPracticePlansParams struct {
	UserID string `json:"userId"`
	Limit  int64  `json:"limit"`
	Offset int64  `json:"offset"`
}

type ListPaginatedPracticePlansRow struct {
	ID                       string         `json:"id"`
	UserID                   string         `json:"userId"`
	Intensity                string         `json:"intensity"`
	Date                     int64          `json:"date"`
	Completed                bool           `json:"completed"`
	PracticeSessionID        sql.NullString `json:"practiceSessionId"`
	NewSpotsCount            int64          `json:"newSpotsCount"`
	RepeatSpotsCount         int64          `json:"repeatSpotsCount"`
	InterleaveSpotsCount     int64          `json:"interleaveSpotsCount"`
	InterleaveDaysSpotsCount int64          `json:"interleaveDaysSpotsCount"`
	PiecesCount              int64          `json:"piecesCount"`
}

func (q *Queries) ListPaginatedPracticePlans(ctx context.Context, arg ListPaginatedPracticePlansParams) ([]ListPaginatedPracticePlansRow, error) {
	rows, err := q.db.QueryContext(ctx, listPaginatedPracticePlans, arg.UserID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListPaginatedPracticePlansRow
	for rows.Next() {
		var i ListPaginatedPracticePlansRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Intensity,
			&i.Date,
			&i.Completed,
			&i.PracticeSessionID,
			&i.NewSpotsCount,
			&i.RepeatSpotsCount,
			&i.InterleaveSpotsCount,
			&i.InterleaveDaysSpotsCount,
			&i.PiecesCount,
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

const listRecentPracticePlans = `-- name: ListRecentPracticePlans :many
SELECT
    practice_plans.id, practice_plans.user_id, practice_plans.intensity, practice_plans.date, practice_plans.completed, practice_plans.practice_session_id,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND practice_plan_spots.practice_type = 'new') AS new_spots_count,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND practice_plan_spots.practice_type = 'extra_repeat') AS repeat_spots_count,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND practice_plan_spots.practice_type = 'interleave') AS interleave_spots_count,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND practice_plan_spots.practice_type = 'interleave_days') AS interleave_days_spots_count,
    (SELECT COUNT(*) FROM practice_plan_pieces WHERE practice_plan_pieces.practice_plan_id = practice_plans.id) AS pieces_count
FROM practice_plans
WHERE practice_plans.id != ? AND practice_plans.user_id = ?
ORDER BY practice_plans.date DESC
LIMIT 3
`

type ListRecentPracticePlansParams struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
}

type ListRecentPracticePlansRow struct {
	ID                       string         `json:"id"`
	UserID                   string         `json:"userId"`
	Intensity                string         `json:"intensity"`
	Date                     int64          `json:"date"`
	Completed                bool           `json:"completed"`
	PracticeSessionID        sql.NullString `json:"practiceSessionId"`
	NewSpotsCount            int64          `json:"newSpotsCount"`
	RepeatSpotsCount         int64          `json:"repeatSpotsCount"`
	InterleaveSpotsCount     int64          `json:"interleaveSpotsCount"`
	InterleaveDaysSpotsCount int64          `json:"interleaveDaysSpotsCount"`
	PiecesCount              int64          `json:"piecesCount"`
}

func (q *Queries) ListRecentPracticePlans(ctx context.Context, arg ListRecentPracticePlansParams) ([]ListRecentPracticePlansRow, error) {
	rows, err := q.db.QueryContext(ctx, listRecentPracticePlans, arg.ID, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListRecentPracticePlansRow
	for rows.Next() {
		var i ListRecentPracticePlansRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Intensity,
			&i.Date,
			&i.Completed,
			&i.PracticeSessionID,
			&i.NewSpotsCount,
			&i.RepeatSpotsCount,
			&i.InterleaveSpotsCount,
			&i.InterleaveDaysSpotsCount,
			&i.PiecesCount,
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
