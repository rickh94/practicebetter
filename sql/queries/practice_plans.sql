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

-- name: CreatePracticePlanSpotWithIdx :one
INSERT INTO practice_plan_spots (
    practice_plan_id,
    spot_id,
    practice_type,
    idx
) VALUES (?, ?, ?, ?)
RETURNING *;

-- name: CreatePracticePlanPiece :one
INSERT INTO practice_plan_pieces (
    practice_plan_id,
    piece_id,
    practice_type
) VALUES (?, ?, ?)
RETURNING *;

-- name: CreatePracticePlanPieceWithIdx :one
INSERT INTO practice_plan_pieces (
    practice_plan_id,
    piece_id,
    practice_type,
    idx
) VALUES (?, ?, ?, ?)
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
WHERE practice_plans.id = ? AND practice_plans.user_id = ?
ORDER BY practice_plan_pieces.idx;

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
WHERE practice_plans.id = ? AND practice_plans.user_id = ?
ORDER BY practice_plan_spots.idx;

-- name: GetPracticePlanWithTodo :one
SELECT
    practice_plans.*,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND practice_plan_spots.completed = true) AS completed_spots_count,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id) AS spots_count,
    (SELECT COUNT(*) FROM practice_plan_pieces WHERE practice_plan_pieces.practice_plan_id = practice_plans.id AND practice_plan_pieces.completed = true) AS completed_pieces_count,
    (SELECT COUNT(*) FROM practice_plan_pieces WHERE practice_plan_pieces.practice_plan_id = practice_plans.id) AS pieces_count,
    IFNULL((SELECT GROUP_CONCAT(DISTINCT pieces.title||'@') FROM practice_plan_pieces INNER JOIN pieces ON practice_plan_pieces.piece_id = pieces.id WHERE practice_plan_pieces.practice_plan_id = practice_plans.id), '') AS piece_titles,
    IFNULL((SELECT GROUP_CONCAT(DISTINCT pieces.title||'@') FROM practice_plan_spots INNER JOIN spots ON spots.id = practice_plan_spots.spot_id INNER JOIN pieces ON pieces.id = spots.piece_id WHERE practice_plan_spots.practice_plan_id = practice_plans.id), '') AS spot_piece_titles
FROM practice_plans
WHERE practice_plans.id = ? AND practice_plans.user_id = ?
LIMIT 1;

-- name: ListRecentPracticePlans :many
SELECT
    practice_plans.*,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND practice_plan_spots.completed = true) AS completed_spots_count,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id) AS spots_count,
    (SELECT COUNT(*) FROM practice_plan_pieces WHERE practice_plan_pieces.practice_plan_id = practice_plans.id AND practice_plan_pieces.completed = true) AS completed_pieces_count,
    (SELECT COUNT(*) FROM practice_plan_pieces WHERE practice_plan_pieces.practice_plan_id = practice_plans.id) AS pieces_count,
    IFNULL((SELECT GROUP_CONCAT(DISTINCT pieces.title||'@') FROM practice_plan_pieces INNER JOIN pieces ON practice_plan_pieces.piece_id = pieces.id WHERE practice_plan_pieces.practice_plan_id = practice_plans.id), '') AS piece_titles,
    IFNULL((SELECT GROUP_CONCAT(DISTINCT pieces.title||'@') FROM practice_plan_spots INNER JOIN spots ON spots.id = practice_plan_spots.spot_id INNER JOIN pieces ON pieces.id = spots.piece_id WHERE practice_plan_spots.practice_plan_id = practice_plans.id), '') AS spot_piece_titles
FROM practice_plans
WHERE practice_plans.id != ? AND practice_plans.user_id = ?
ORDER BY practice_plans.date DESC
LIMIT 3;

-- name: ListPaginatedPracticePlans :many
SELECT
    practice_plans.*,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND practice_plan_spots.completed = true) AS completed_spots_count,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id) AS spots_count,
    (SELECT COUNT(*) FROM practice_plan_pieces WHERE practice_plan_pieces.practice_plan_id = practice_plans.id AND practice_plan_pieces.completed = true) AS completed_pieces_count,
    (SELECT COUNT(*) FROM practice_plan_pieces WHERE practice_plan_pieces.practice_plan_id = practice_plans.id) AS pieces_count,
    IFNULL((SELECT GROUP_CONCAT(DISTINCT pieces.title||'@') FROM practice_plan_pieces INNER JOIN pieces ON practice_plan_pieces.piece_id = pieces.id WHERE practice_plan_pieces.practice_plan_id = practice_plans.id), '') AS piece_titles,
    IFNULL((SELECT GROUP_CONCAT(DISTINCT pieces.title||'@') FROM practice_plan_spots INNER JOIN spots ON spots.id = practice_plan_spots.spot_id INNER JOIN pieces ON pieces.id = spots.piece_id WHERE practice_plan_spots.practice_plan_id = practice_plans.id), '') AS spot_piece_titles
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
WHERE practice_plan_spots.practice_type = 'interleave_days' AND practice_plan_spots.practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id)
ORDER BY practice_plan_spots.idx;

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
WHERE practice_plan_spots.practice_type = 'interleave' AND practice_plan_spots.practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id)
ORDER BY practice_plan_spots.idx;

-- name: ListPracticePlanSpotsInCategory :many
SELECT practice_plan_spots.completed,
    spots.name,
    spots.id,
    spots.measures,
    spots.piece_id,
    spots.stage,
    spots.stage_started,
    spots.skip_days,
    (SELECT pieces.title FROM pieces WHERE pieces.id = spots.piece_id LIMIT 1) AS piece_title
FROM practice_plan_spots
INNER JOIN spots ON practice_plan_spots.spot_id = spots.id
WHERE practice_plan_spots.practice_type = :practice_type
AND practice_plan_spots.practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id)
ORDER BY practice_plan_spots.idx;

-- name: ListPracticePlanPiecesInCategory :many
SELECT practice_plan_pieces.completed,
    practice_plan_pieces.completed AS piece_completed,
    pieces.title AS piece_title,
    pieces.id AS piece_id,
    pieces.composer AS piece_composer,
    (SELECT COUNT(id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage != 'completed') AS piece_active_spots,
    (SELECT COUNT(id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage == 'random') AS piece_random_spots,
    (SELECT COUNT(id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage == 'completed') AS piece_completed_spots
FROM practice_plan_pieces
INNER JOIN pieces ON practice_plan_pieces.piece_id = pieces.id
WHERE practice_plan_pieces.practice_type = :practice_type
AND practice_plan_pieces.practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id)
ORDER BY practice_plan_pieces.idx;

-- name: GetPracticePlanIncompleteExtraRepeatSpots :many
SELECT practice_plan_spots.*,
    spots.piece_id AS spot_piece_id
FROM practice_plan_spots
INNER JOIN spots ON practice_plan_spots.spot_id = spots.id
WHERE practice_plan_spots.practice_type = 'extra_repeat'
AND practice_plan_spots.completed = false
AND practice_plan_spots.practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id)
ORDER BY practice_plan_spots.idx;

-- name: GetPracticePlanIncompleteRandomPieces :many
SELECT *
FROM practice_plan_pieces
WHERE practice_plan_pieces.practice_type = 'random_spots'
AND practice_plan_pieces.completed = false
AND practice_plan_pieces.practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id)
ORDER BY practice_plan_pieces.idx;

-- name: GetPracticePlanIncompleteStartingPointPieces :many
SELECT *
FROM practice_plan_pieces
WHERE practice_plan_pieces.practice_type = 'starting_point'
AND practice_plan_pieces.completed = false
AND practice_plan_pieces.practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id)
ORDER BY practice_plan_pieces.idx;

-- name: GetPracticePlanIncompleteNewSpots :many
SELECT practice_plan_spots.*,
    spots.piece_id AS spot_piece_id
FROM practice_plan_spots
INNER JOIN spots ON practice_plan_spots.spot_id = spots.id
WHERE practice_plan_spots.practice_type = 'new'
AND practice_plan_spots.completed = false
AND practice_plan_spots.practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id)
ORDER BY practice_plan_spots.idx;

-- name: GetPracticePlan :one
SELECT *
FROM practice_plans
WHERE id = ? AND user_id = ?;

-- name: UpdatePlanPieceIdx :exec
UPDATE practice_plan_pieces
SET idx = ?
WHERE practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id)
AND piece_id = ?
AND practice_type = ?;

-- name: UpdatePlanSpotIdx :exec
UPDATE practice_plan_spots
SET idx = ?
WHERE practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id)
AND spot_id = ?;

-- name: GetMaxSpotIdx :one
SELECT MAX(idx) FROM practice_plan_spots
WHERE practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id);

-- name: GetMaxPieceIdx :one
SELECT MAX(idx) FROM practice_plan_pieces
WHERE practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id);

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

-- name: DeletePracticePlanSpot :exec
DELETE FROM practice_plan_spots
WHERE practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id)
AND spot_id = :spot_id
AND practice_type = :practice_type;

-- name: DeletePracticePlanPiece :exec
DELETE FROM practice_plan_pieces
WHERE practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id)
AND piece_id = :piece_id
AND practice_type = :practice_type;

-- name: DeletePracticePlan :exec
DELETE FROM practice_plans
WHERE id = ? AND user_id = ?;
