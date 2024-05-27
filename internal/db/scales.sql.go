// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: scales.sql

package db

import (
	"context"
	"database/sql"
)

const checkForUserScale = `-- name: CheckForUserScale :one
SELECT id FROM user_scales WHERE user_id = ?1 AND scale_id = ?2
`

type CheckForUserScaleParams struct {
	UserID  string `json:"userId"`
	ScaleID int64  `json:"scaleId"`
}

func (q *Queries) CheckForUserScale(ctx context.Context, arg CheckForUserScaleParams) (string, error) {
	row := q.db.QueryRowContext(ctx, checkForUserScale, arg.UserID, arg.ScaleID)
	var id string
	err := row.Scan(&id)
	return id, err
}

const createUserScale = `-- name: CreateUserScale :one
INSERT INTO user_scales (
    id,
    user_id,
    scale_id,
    practice_notes,
    reference
) VALUES (?, ?, ?, ?, ?)
RETURNING id, user_id, scale_id, practice_notes, last_practiced, reference, working
`

type CreateUserScaleParams struct {
	ID            string `json:"id"`
	UserID        string `json:"userId"`
	ScaleID       int64  `json:"scaleId"`
	PracticeNotes string `json:"practiceNotes"`
	Reference     string `json:"reference"`
}

func (q *Queries) CreateUserScale(ctx context.Context, arg CreateUserScaleParams) (UserScale, error) {
	row := q.db.QueryRowContext(ctx, createUserScale,
		arg.ID,
		arg.UserID,
		arg.ScaleID,
		arg.PracticeNotes,
		arg.Reference,
	)
	var i UserScale
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ScaleID,
		&i.PracticeNotes,
		&i.LastPracticed,
		&i.Reference,
		&i.Working,
	)
	return i, err
}

const getMode = `-- name: GetMode :one
SELECT id, name, basic, cof FROM scale_modes WHERE id = ?
`

func (q *Queries) GetMode(ctx context.Context, id int64) (ScaleMode, error) {
	row := q.db.QueryRowContext(ctx, getMode, id)
	var i ScaleMode
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Basic,
		&i.Cof,
	)
	return i, err
}

const getScale = `-- name: GetScale :one
SELECT
    scales.id,
    scales.key_id,
    scales.mode_id,
    scale_keys.name as key_name,
    scale_modes.name as mode
FROM scales
INNER JOIN scale_keys ON scale_keys.id = scales.key_id
INNER JOIN scale_modes on scale_modes.id = scales.mode_id
WHERE scales.id = ?1
`

type GetScaleRow struct {
	ID      int64  `json:"id"`
	KeyID   int64  `json:"keyId"`
	ModeID  int64  `json:"modeId"`
	KeyName string `json:"keyName"`
	Mode    string `json:"mode"`
}

func (q *Queries) GetScale(ctx context.Context, id int64) (GetScaleRow, error) {
	row := q.db.QueryRowContext(ctx, getScale, id)
	var i GetScaleRow
	err := row.Scan(
		&i.ID,
		&i.KeyID,
		&i.ModeID,
		&i.KeyName,
		&i.Mode,
	)
	return i, err
}

const getUserScale = `-- name: GetUserScale :one
SELECT
    user_scales.id,
    scale_keys.name AS key_name,
    scale_modes.name AS mode,
    user_scales.practice_notes AS practice_notes,
    user_scales.last_practiced AS last_practiced,
    user_scales.reference AS reference,
    user_scales.working AS working
FROM user_scales
INNER JOIN scales ON scales.id = user_scales.scale_id
INNER JOIN scale_keys ON scale_keys.id = scales.key_id
INNER JOIN scale_modes on scale_modes.id = scales.mode_id
WHERE user_scales.id = ?1 AND user_scales.user_id = ?2
`

type GetUserScaleParams struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
}

type GetUserScaleRow struct {
	ID            string        `json:"id"`
	KeyName       string        `json:"keyName"`
	Mode          string        `json:"mode"`
	PracticeNotes string        `json:"practiceNotes"`
	LastPracticed sql.NullInt64 `json:"lastPracticed"`
	Reference     string        `json:"reference"`
	Working       bool          `json:"working"`
}

func (q *Queries) GetUserScale(ctx context.Context, arg GetUserScaleParams) (GetUserScaleRow, error) {
	row := q.db.QueryRowContext(ctx, getUserScale, arg.ID, arg.UserID)
	var i GetUserScaleRow
	err := row.Scan(
		&i.ID,
		&i.KeyName,
		&i.Mode,
		&i.PracticeNotes,
		&i.LastPracticed,
		&i.Reference,
		&i.Working,
	)
	return i, err
}

const listBasicScales = `-- name: ListBasicScales :many
SELECT
    scales.id,
    scales.key_id,
    scales.mode_id,
    scale_keys.name as key_name,
    scale_modes.name as mode
FROM scales
INNER JOIN scale_keys ON scale_keys.id = scales.key_id
INNER JOIN scale_modes on scale_modes.id = scales.mode_id
WHERE scale_modes.basic = true
`

type ListBasicScalesRow struct {
	ID      int64  `json:"id"`
	KeyID   int64  `json:"keyId"`
	ModeID  int64  `json:"modeId"`
	KeyName string `json:"keyName"`
	Mode    string `json:"mode"`
}

func (q *Queries) ListBasicScales(ctx context.Context) ([]ListBasicScalesRow, error) {
	rows, err := q.db.QueryContext(ctx, listBasicScales)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListBasicScalesRow
	for rows.Next() {
		var i ListBasicScalesRow
		if err := rows.Scan(
			&i.ID,
			&i.KeyID,
			&i.ModeID,
			&i.KeyName,
			&i.Mode,
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

const listKeys = `-- name: ListKeys :many
SELECT id, name, cof FROM scale_keys ORDER BY id
`

func (q *Queries) ListKeys(ctx context.Context) ([]ScaleKey, error) {
	rows, err := q.db.QueryContext(ctx, listKeys)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ScaleKey
	for rows.Next() {
		var i ScaleKey
		if err := rows.Scan(&i.ID, &i.Name, &i.Cof); err != nil {
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

const listModes = `-- name: ListModes :many
SELECT id, name, basic, cof FROM scale_modes WHERE basic = ? ORDER BY id
`

func (q *Queries) ListModes(ctx context.Context, basic bool) ([]ScaleMode, error) {
	rows, err := q.db.QueryContext(ctx, listModes, basic)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ScaleMode
	for rows.Next() {
		var i ScaleMode
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Basic,
			&i.Cof,
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

const listScales = `-- name: ListScales :many
SELECT
    scales.id,
    scales.key_id,
    scales.mode_id,
    scale_keys.name as key_name,
    scale_modes.name as mode,
    scale_modes.basic as basic
FROM scales
INNER JOIN scale_keys ON scale_keys.id = scales.key_id
INNER JOIN scale_modes on scale_modes.id = scales.mode_id
`

type ListScalesRow struct {
	ID      int64  `json:"id"`
	KeyID   int64  `json:"keyId"`
	ModeID  int64  `json:"modeId"`
	KeyName string `json:"keyName"`
	Mode    string `json:"mode"`
	Basic   bool   `json:"basic"`
}

func (q *Queries) ListScales(ctx context.Context) ([]ListScalesRow, error) {
	rows, err := q.db.QueryContext(ctx, listScales)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListScalesRow
	for rows.Next() {
		var i ListScalesRow
		if err := rows.Scan(
			&i.ID,
			&i.KeyID,
			&i.ModeID,
			&i.KeyName,
			&i.Mode,
			&i.Basic,
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

const listScalesForMode = `-- name: ListScalesForMode :many
SELECT
    scales.id,
    scales.key_id,
    scales.mode_id,
    scale_keys.name as key_name,
    (SELECT (scale_keys.cof + scale_modes.cof) % 12) as scale_cof,
    scale_modes.name as mode,
    scale_modes.basic as basic
FROM scales
INNER JOIN scale_keys ON scale_keys.id = scales.key_id
INNER JOIN scale_modes on scale_modes.id = scales.mode_id
WHERE scales.mode_id = ?1
ORDER BY (scale_keys.cof + scale_modes.cof) % 12
`

type ListScalesForModeRow struct {
	ID       int64       `json:"id"`
	KeyID    int64       `json:"keyId"`
	ModeID   int64       `json:"modeId"`
	KeyName  string      `json:"keyName"`
	ScaleCof interface{} `json:"scaleCof"`
	Mode     string      `json:"mode"`
	Basic    bool        `json:"basic"`
}

func (q *Queries) ListScalesForMode(ctx context.Context, modeID int64) ([]ListScalesForModeRow, error) {
	rows, err := q.db.QueryContext(ctx, listScalesForMode, modeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListScalesForModeRow
	for rows.Next() {
		var i ListScalesForModeRow
		if err := rows.Scan(
			&i.ID,
			&i.KeyID,
			&i.ModeID,
			&i.KeyName,
			&i.ScaleCof,
			&i.Mode,
			&i.Basic,
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

const listUserScales = `-- name: ListUserScales :many
SELECT id, user_id, scale_id, practice_notes, last_practiced, reference, working FROM user_scales WHERE user_id = ?1
`

func (q *Queries) ListUserScales(ctx context.Context, userID string) ([]UserScale, error) {
	rows, err := q.db.QueryContext(ctx, listUserScales, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []UserScale
	for rows.Next() {
		var i UserScale
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.ScaleID,
			&i.PracticeNotes,
			&i.LastPracticed,
			&i.Reference,
			&i.Working,
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

const listWorkingScales = `-- name: ListWorkingScales :many
SELECT
    user_scales.id,
    scale_keys.name AS key_name,
    scale_modes.name AS mode,
    user_scales.practice_notes AS practice_notes,
    user_scales.last_practiced AS last_practiced,
    user_scales.reference AS reference
FROM user_scales
INNER JOIN scales ON scales.id = user_scales.scale_id
INNER JOIN scale_keys ON scale_keys.id = scales.key_id
INNER JOIN scale_modes on scale_modes.id = scales.mode_id
WHERE user_scales.working = true AND user_scales.user_id = ?1
`

type ListWorkingScalesRow struct {
	ID            string        `json:"id"`
	KeyName       string        `json:"keyName"`
	Mode          string        `json:"mode"`
	PracticeNotes string        `json:"practiceNotes"`
	LastPracticed sql.NullInt64 `json:"lastPracticed"`
	Reference     string        `json:"reference"`
}

func (q *Queries) ListWorkingScales(ctx context.Context, userID string) ([]ListWorkingScalesRow, error) {
	rows, err := q.db.QueryContext(ctx, listWorkingScales, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListWorkingScalesRow
	for rows.Next() {
		var i ListWorkingScalesRow
		if err := rows.Scan(
			&i.ID,
			&i.KeyName,
			&i.Mode,
			&i.PracticeNotes,
			&i.LastPracticed,
			&i.Reference,
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

const updateScalePracticed = `-- name: UpdateScalePracticed :one
UPDATE user_scales
SET
    last_practiced = unixepoch("now")
WHERE id = ?1 AND user_id = ?2
RETURNING id, user_id, scale_id, practice_notes, last_practiced, reference, working
`

type UpdateScalePracticedParams struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
}

func (q *Queries) UpdateScalePracticed(ctx context.Context, arg UpdateScalePracticedParams) (UserScale, error) {
	row := q.db.QueryRowContext(ctx, updateScalePracticed, arg.ID, arg.UserID)
	var i UserScale
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ScaleID,
		&i.PracticeNotes,
		&i.LastPracticed,
		&i.Reference,
		&i.Working,
	)
	return i, err
}

const updateUserScale = `-- name: UpdateUserScale :one
UPDATE user_scales
SET
    practice_notes = ?1,
    last_practiced = ?2,
    reference = ?3,
    working = ?4
WHERE id = ?5 AND user_id = ?6
RETURNING id, user_id, scale_id, practice_notes, last_practiced, reference, working
`

type UpdateUserScaleParams struct {
	PracticeNotes string        `json:"practiceNotes"`
	LastPracticed sql.NullInt64 `json:"lastPracticed"`
	Reference     string        `json:"reference"`
	Working       bool          `json:"working"`
	ID            string        `json:"id"`
	UserID        string        `json:"userId"`
}

func (q *Queries) UpdateUserScale(ctx context.Context, arg UpdateUserScaleParams) (UserScale, error) {
	row := q.db.QueryRowContext(ctx, updateUserScale,
		arg.PracticeNotes,
		arg.LastPracticed,
		arg.Reference,
		arg.Working,
		arg.ID,
		arg.UserID,
	)
	var i UserScale
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ScaleID,
		&i.PracticeNotes,
		&i.LastPracticed,
		&i.Reference,
		&i.Working,
	)
	return i, err
}
