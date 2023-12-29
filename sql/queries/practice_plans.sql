-- name: CreatePracticePlan :one
INSERT INTO practice_plans (
    id,
    user_id,
    intensity,
    practice_session_id,
    date
) VALUES (?, ?, ?, ?, unixepoch('now'))
RETURNING *;

-- name: CreatePracticePlanSpot :one
INSERT INTO practice_plan_spots (
    practice_plan_id,
    spot_id,
    practice_type
) VALUES (?, ?, ?)
RETURNING *;

-- name: CreatePracticePlanPiece :one
INSERT INTO practice_plan_pieces (
    practice_plan_id,
    piece_id,
    practice_type
) VALUES (?, ?, ?)
RETURNING *;

-- name: GetPracticePlanWithPieces :many
SELECT
    practice_plans.*,
    practice_plan_pieces.practice_type as piece_practice_type,
    practice_plan_pieces.completed AS piece_completed,
    pieces.title AS piece_title,
    pieces.id AS piece_id,
    pieces.composer AS piece_composer,
    (SELECT COUNT(id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage != 'completed') AS piece_active_spots,
    (SELECT COUNT(id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage == 'random') AS piece_random_spots,
    (SELECT COUNT(id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage == 'completed') AS piece_completed_spots
FROM practice_plans
INNER JOIN practice_plan_pieces ON practice_plans.id = practice_plan_pieces.practice_plan_id
LEFT JOIN pieces ON practice_plan_pieces.piece_id = pieces.id
WHERE practice_plans.id = ? AND practice_plans.user_id = ?;

-- name: GetPracticePlanWithSpots :many
SELECT
    practice_plans.*,
    practice_plan_spots.practice_type as spot_practice_type,
    practice_plan_spots.completed as spot_completed,
    spots.name AS spot_name,
    spots.id AS spot_id,
    spots.measures AS spot_measures,
    spots.piece_id AS spot_piece_id,
    spots.stage AS spot_stage,
    spots.skip_days AS spot_skip_days,
    spots.stage_started AS spot_stage_started,
    (SELECT pieces.title FROM pieces WHERE pieces.id = spots.piece_id LIMIT 1) AS spot_piece_title
FROM practice_plans
INNER JOIN practice_plan_spots ON practice_plans.id = practice_plan_spots.practice_plan_id
LEFT JOIN spots ON practice_plan_spots.spot_id = spots.id
WHERE practice_plans.id = ? AND practice_plans.user_id = ?;

-- name: GetPracticePlanWithTodo :one
SELECT
    practice_plans.*,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND NOT practice_plan_spots.completed) AS incomplete_spots_count,
    (SELECT COUNT(*) FROM practice_plan_pieces WHERE practice_plan_pieces.practice_plan_id = practice_plans.id AND NOT practice_plan_pieces.completed) AS incomplete_pieces_count,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND practice_plan_spots.completed) AS completed_spots_count,
    (SELECT COUNT(*) FROM practice_plan_pieces WHERE practice_plan_pieces.practice_plan_id = practice_plans.id AND practice_plan_pieces.completed) AS completed_pieces_count
FROM practice_plans
WHERE practice_plans.id = ? AND practice_plans.user_id = ?;

-- name: ListRecentPracticePlans :many
SELECT
    practice_plans.*,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND practice_plan_spots.practice_type = 'new') AS new_spots_count,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND practice_plan_spots.practice_type = 'extra_repeat') AS repeat_spots_count,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND practice_plan_spots.practice_type = 'interleave') AS interleave_spots_count,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND practice_plan_spots.practice_type = 'interleave_days') AS interleave_days_spots_count,
    (SELECT COUNT(*) FROM practice_plan_pieces WHERE practice_plan_pieces.practice_plan_id = practice_plans.id) AS pieces_count
FROM practice_plans
WHERE practice_plans.id != ? AND practice_plans.user_id = ?
ORDER BY practice_plans.date DESC
LIMIT 3;

-- name: ListPaginatedPracticePlans :many
SELECT
    practice_plans.*,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND practice_plan_spots.practice_type = 'new') AS new_spots_count,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND practice_plan_spots.practice_type = 'extra_repeat') AS repeat_spots_count,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND practice_plan_spots.practice_type = 'interleave') AS interleave_spots_count,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND practice_plan_spots.practice_type = 'interleave_days') AS interleave_days_spots_count,
    (SELECT COUNT(*) FROM practice_plan_pieces WHERE practice_plan_pieces.practice_plan_id = practice_plans.id) AS pieces_count
FROM practice_plans
WHERE practice_plans.user_id = ?
ORDER BY practice_plans.date DESC
LIMIT ? OFFSET ?;

-- name: CountUserPracticePlans :one
SELECT COUNT(*) FROM practice_plans WHERE user_id = ?;

-- name: GetPracticePlanInterleaveDaysSpots :many
SELECT practice_plan_spots.*,
    spots.name AS spot_name,
    spots.measures AS spot_measures,
    spots.piece_id AS spot_piece_id,
    spots.stage AS spot_stage,
    spots.stage_started AS spot_stage_started,
    spots.skip_days AS spot_skip_days,
    (SELECT pieces.title FROM pieces WHERE pieces.id = spots.piece_id LIMIT 1) AS spot_piece_title
FROM practice_plan_spots
LEFT JOIN spots ON practice_plan_spots.spot_id = spots.id
WHERE practice_plan_spots.practice_type = 'interleave_days' AND practice_plan_spots.practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id);

-- name: GetPracticePlanInterleaveSpots :many
SELECT practice_plan_spots.*,
    spots.name AS spot_name,
    spots.measures AS spot_measures,
    spots.piece_id AS spot_piece_id,
    spots.stage AS spot_stage,
    spots.stage_started AS spot_stage_started,
    (SELECT pieces.title FROM pieces WHERE pieces.id = spots.piece_id LIMIT 1) AS spot_piece_title
FROM practice_plan_spots
LEFT JOIN spots ON practice_plan_spots.spot_id = spots.id
WHERE practice_plan_spots.practice_type = 'interleave' AND practice_plan_spots.practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id);

-- name: GetPracticePlan :one
SELECT *
FROM practice_plans
WHERE id = ? AND user_id = ?;

-- name: CompletePracticePlanSpot :exec
UPDATE practice_plan_spots
SET completed = true
WHERE practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id) AND spot_id = ?;

-- name: CompletePracticePlanPiece :exec
UPDATE practice_plan_pieces
SET completed = true
WHERE practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id) AND piece_id = ? AND practice_type = ?;

-- name: AddPracticeSessionToPlan :exec
UPDATE practice_plans
SET practice_session_id = ?
WHERE id = ? AND user_id = ?;

-- name: DeletePracticePlan :exec
DELETE FROM practice_plans
WHERE id = ? AND user_id = ?;
