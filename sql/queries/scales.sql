-- name: CreateUserScale :one
INSERT INTO user_scales (
    id,
    user_id,
    scale_id,
    practice_notes,
    reference
) VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: ListScales :many
SELECT
    scales.id,
    scales.key_id,
    scales.mode_id,
    scale_keys.name as key_name,
    scale_modes.name as mode,
    scale_modes.basic as basic
FROM scales
INNER JOIN scale_keys ON scale_keys.id = scales.key_id
INNER JOIN scale_modes on scale_modes.id = scales.mode_id;

-- name: ListScalesForMode :many
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
WHERE scales.mode_id = :mode_id
ORDER BY (scale_keys.cof + scale_modes.cof) % 12;

-- name: ListBasicScales :many
SELECT
    scales.id,
    scales.key_id,
    scales.mode_id,
    scale_keys.name as key_name,
    scale_modes.name as mode
FROM scales
INNER JOIN scale_keys ON scale_keys.id = scales.key_id
INNER JOIN scale_modes on scale_modes.id = scales.mode_id
WHERE scale_modes.basic = true;

-- name: GetScale :one
SELECT
    scales.id,
    scales.key_id,
    scales.mode_id,
    scale_keys.name as key_name,
    scale_modes.name as mode
FROM scales
INNER JOIN scale_keys ON scale_keys.id = scales.key_id
INNER JOIN scale_modes on scale_modes.id = scales.mode_id
WHERE scales.id = :id;

-- name: ListWorkingScales :many
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
WHERE user_scales.working = true AND user_scales.user_id = :user_id;

-- name: GetUserScale :one
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
WHERE user_scales.id = :id AND user_scales.user_id = :user_id;

-- name: CheckForUserScale :one
SELECT id FROM user_scales WHERE user_id = :user_id AND scale_id = :scale_id;

-- name: ListUserScales :many
SELECT * FROM user_scales WHERE user_id = :user_id;

-- name: ListModes :many
SELECT * FROM scale_modes WHERE basic = ? ORDER BY id;

-- name: ListKeys :many
SELECT * FROM scale_keys ORDER BY id;

-- name: GetMode :one
SELECT * FROM scale_modes WHERE id = ?;

-- name: UpdateUserScale :one
UPDATE user_scales
SET
    practice_notes = :practice_notes,
    last_practiced = :last_practiced,
    reference = :reference,
    working = :working
WHERE id = :id AND user_id = :user_id
RETURNING *;

-- name: UpdateScalePracticed :one
UPDATE user_scales
SET
    last_practiced = unixepoch("now")
WHERE id = :id AND user_id = :user_id
RETURNING *;
