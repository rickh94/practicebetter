-- name: CreatePracticePlan :one
INSERT INTO practice_plans (
    id,
    user_id,
    intensity,
    date
) VALUES (?, ?, ?, unixepoch('now'))
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

-- name: CreatePracticePlanScaleWithIdx :one
INSERT INTO practice_plan_scales (
    practice_plan_id,
    user_scale_id,
    idx
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

-- name: GetPracticePlanWithScales :many
SELECT
    practice_plans.*,
    practice_plan_scales.completed AS scale_completed,
    user_scales.id AS user_scale_id,
    user_scales.practice_notes AS scale_practice_notes,
    user_scales.last_practiced AS scale_last_practiced,
    user_scales.reference AS scale_reference,
    scale_keys.name AS scale_key_name,
    scale_modes.name AS scale_mode
FROM practice_plans
INNER JOIN practice_plan_scales ON practice_plans.id = practice_plan_scales.practice_plan_id
INNER JOIN user_scales ON practice_plan_scales.user_scale_id = user_scales.id
INNER JOIN scales ON user_scales.scale_id = scales.id
INNER JOIN scale_keys ON scale_keys.id = scales.key_id
INNER JOIN scale_modes ON scale_modes.id = scales.mode_id
WHERE practice_plans.id = :practice_plan_id AND practice_plans.user_id = :user_id AND user_scales.user_id = :user_id
ORDER BY practice_plan_scales.idx;

-- name: GetPracticePlanWithTodo :one
SELECT
    practice_plans.*,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id AND practice_plan_spots.completed = true) AS completed_spots_count,
    (SELECT COUNT(*) FROM practice_plan_spots WHERE practice_plan_spots.practice_plan_id = practice_plans.id) AS spots_count,
    (SELECT COUNT(*) FROM practice_plan_pieces WHERE practice_plan_pieces.practice_plan_id = practice_plans.id AND practice_plan_pieces.completed = true) AS completed_pieces_count,
    (SELECT COUNT(*) FROM practice_plan_pieces WHERE practice_plan_pieces.practice_plan_id = practice_plans.id) AS pieces_count,
    (SELECT COUNT(*) FROM practice_plan_scales WHERE practice_plan_scales.practice_plan_id = practice_plans.id AND practice_plan_scales.completed = true) AS completed_scales_count,
    (SELECT COUNT(*) FROM practice_plan_scales WHERE practice_plan_scales.practice_plan_id = practice_plans.id) AS scales_count,
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
    (SELECT COUNT(*) FROM practice_plan_scales WHERE practice_plan_scales.practice_plan_id = practice_plans.id AND practice_plan_scales.completed = true) AS completed_scales_count,
    (SELECT COUNT(*) FROM practice_plan_scales WHERE practice_plan_scales.practice_plan_id = practice_plans.id) AS scales_count,
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

-- name: GetNextInfrequentSpot :one
SELECT spots.*,
    (SELECT pieces.title FROM pieces WHERE pieces.id = spots.piece_id LIMIT 1) AS piece_title
FROM practice_plan_spots
INNER JOIN spots ON practice_plan_spots.spot_id = spots.id
WHERE practice_plan_spots.practice_type = 'interleave_days'
    AND practice_plan_spots.practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id)
    AND practice_plan_spots.completed = false
ORDER BY practice_plan_spots.idx
LIMIT 1;

-- name: HasIncompleteInfrequentSpots :one
SELECT  COUNT(*) > 0
FROM practice_plan_spots
WHERE practice_plan_spots.practice_type = 'interleave_days'
    AND practice_plan_spots.practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id)
    AND practice_plan_spots.completed = false;

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

-- name: GetPracticePlanInterleaveSpot :one
SELECT practice_plan_spots.*,
    spots.name AS spot_name,
    spots.measures AS spot_measures,
    spots.piece_id AS spot_piece_id,
    spots.stage AS spot_stage,
    spots.stage_started AS spot_stage_started,
    (SELECT pieces.title FROM pieces WHERE pieces.id = spots.piece_id LIMIT 1) AS spot_piece_title
FROM practice_plan_spots
LEFT JOIN spots ON practice_plan_spots.spot_id = spots.id
WHERE practice_plan_spots.practice_type = 'interleave' AND practice_plan_spots.practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id) AND spots.id = :spot_id
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

-- name: GetPracticePlanFailedNewSpots :many
SELECT practice_plan_spots.*
FROM practice_plan_spots
INNER JOIN spots ON practice_plan_spots.spot_id = spots.id
WHERE practice_plan_spots.practice_type = 'new'
AND practice_plan_spots.practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.user_id = :user_id ORDER BY date DESC LIMIT 1)
AND spots.stage = 'repeat'
AND spots.piece_id IN (sqlc.slice('pieceIDs'))
ORDER BY practice_plan_spots.idx;

-- name: GetPracticePlanEvaluatedInterleaveSpots :many
SELECT practice_plan_spots.*,
    spots.stage_started AS spot_stage_started
FROM practice_plan_spots
INNER JOIN spots ON practice_plan_spots.spot_id = spots.id
WHERE practice_plan_spots.practice_type = 'interleave'
AND practice_plan_spots.evaluation IS NOT NULL
AND practice_plan_spots.completed = false
AND practice_plan_spots.practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id)
ORDER BY practice_plan_spots.idx;

-- name: GetPracticePlan :one
SELECT *
FROM practice_plans
WHERE id = ? AND user_id = ?;

-- name: GetLatestPracticePlan :one
SELECT *
FROM practice_plans
WHERE user_id = ?
ORDER BY date DESC
LIMIT 1;

-- name: GetPreviousPlanNotes :one
SELECT practice_notes
FROM practice_plans
WHERE user_id = ? AND id != :plan_id
ORDER BY date DESC
LIMIT 1;

-- name: GetPlanLastPracticed :one
SELECT last_practiced
FROM practice_plans
WHERE id = ? AND user_id = ?;

-- name: UpdatePlanLastPracticed :exec
UPDATE practice_plans
SET last_practiced = unixepoch("now")
WHERE id = ? AND user_id = ?;

-- name: UpdatePlanPieceIdx :exec
UPDATE practice_plan_pieces
SET idx = ?
WHERE practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id)
AND piece_id = ?
AND practice_type = ?;

-- name: UpdateSpotEvaluation :exec
UPDATE practice_plan_spots
SET evaluation = CASE WHEN evaluation IS NULL OR evaluation <> 'poor' THEN :evaluation ELSE evaluation END
WHERE practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id)
AND spot_id = :spot_id;

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

-- name: GetMaxScaleIdx :one
SELECT MAX(idx) FROM practice_plan_scales
WHERE practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id);

-- name: CompletePracticePlanSpot :exec
UPDATE practice_plan_spots
SET completed = true,
    evaluation = NULL
WHERE practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id) AND spot_id = ?;

-- name: CompletePracticePlanPiece :exec
UPDATE practice_plan_pieces
SET completed = true
WHERE practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id) AND piece_id = ? AND practice_type = ?;

-- name: CompletePracticePlanScale :exec
UPDATE practice_plan_scales
SET completed = true
WHERE practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id) AND user_scale_id = ?;

-- name: CompletePracticePlan :exec
UPDATE practice_plans
SET completed = true,
    practice_notes = ?
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

-- name: DeletePracticePlanScale :one
DELETE FROM practice_plan_scales
WHERE practice_plan_id = (SELECT practice_plans.id FROM practice_plans WHERE practice_plans.id = :plan_id AND practice_plans.user_id = :user_id)
AND user_scale_id = :user_scale_id
RETURNING *;

-- name: DeletePracticePlan :exec
DELETE FROM practice_plans
WHERE id = ? AND user_id = ?;
