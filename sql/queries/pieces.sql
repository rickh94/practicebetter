-- name: CreatePiece :one
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
RETURNING *;

-- name: GetPieceByID :many
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
    spots.id AS spot_id,
    spots.name AS spot_name,
    spots.idx AS spot_idx,
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
WHERE pieces.id = :piece_id AND pieces.user_id = :user_id
ORDER BY spot_last_practiced DESC;


-- name: GetPieceWithRandomSpots :many
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
    spots.stage_started AS spot_stage_started,
    spots.measures AS spot_measures
FROM pieces
INNER JOIN spots ON pieces.id = spots.piece_id AND spots.stage = 'random'
WHERE pieces.id = :piece_id AND pieces.user_id = :user_id;

-- name: GetPieceWithIncompleteSpots :many
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
INNER JOIN spots ON pieces.id = spots.piece_id AND spots.stage != 'completed'
WHERE pieces.id = :piece_id AND pieces.user_id = :user_id;

-- name: GetPieceForPlan :many
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
    spots.last_practiced AS spot_last_practiced,
    spots.skip_days AS spot_skip_days
FROM pieces
LEFT JOIN spots ON pieces.id = spots.piece_id
WHERE pieces.id = :piece_id AND pieces.user_id = :user_id;

-- name: CheckPieceForRandomSpots :one
SELECT COUNT(*) FROM spots WHERE piece_id = ? AND stage = 'random';

-- name: GetPieceWithoutSpots :one
SELECT * FROM pieces WHERE id = ? AND user_id = ?;

-- name: ListRandomSpotPiecesForPlan :many
SELECT
    pieces.*,
    (SELECT COUNT(spots.id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage == 'random') AS random_spot_count
FROM pieces
WHERE user_id = :user_id
AND random_spot_count > 0
AND pieces.id NOT IN (SELECT practice_plan_pieces.piece_id FROM practice_plan_pieces WHERE practice_plan_pieces.practice_plan_id = :plan_id);

-- name: ListActivePiecesWithCompletedSpotsForPlan :many
SELECT
    id,
    title,
    composer,
    last_practiced,
    (SELECT COUNT(spots.id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage == 'completed') AS completed_spot_count
FROM pieces
WHERE user_id = :user_id
AND stage = 'active'
AND completed_spot_count > 5
AND id NOT IN (SELECT practice_plan_pieces.piece_id FROM practice_plan_pieces WHERE practice_plan_pieces.practice_plan_id = :plan_id);

-- name: ListRecentlyPracticedPieces :many
SELECT
    id,
    title,
    composer,
    (SELECT COUNT(spots.id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage == 'completed') AS completed_spots,
    (SELECT COUNT(spots.id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage != 'completed') AS active_spots
FROM pieces
WHERE user_id = ?
ORDER BY last_practiced DESC
LIMIT 5;

-- name: ListPaginatedUserPieces :many
SELECT
    id,
    title,
    composer,
    (SELECT COUNT(spots.id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage == 'completed') AS completed_spots,
    (SELECT COUNT(spots.id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage != 'completed') AS active_spots
FROM pieces
WHERE user_id = ?
ORDER BY last_practiced DESC
LIMIT ? OFFSET ?;

-- name: CountUserPieces :one
SELECT COUNT(*) FROM pieces WHERE user_id = ?;

-- name: ListAllUserPieces :many
SELECT
    id,
    title,
    composer,
    last_practiced,
    (SELECT COUNT(spots.id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage == 'completed') AS completed_spots,
    (SELECT COUNT(spots.id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage != 'completed') AS active_spots
FROM pieces
WHERE user_id = ?;

-- name: ListActiveUserPieces :many
SELECT
    id,
    title,
    composer,
    last_practiced,
    (SELECT COUNT(spots.id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage == 'completed') AS completed_spots,
    (SELECT COUNT(spots.id) FROM spots WHERE spots.piece_id = pieces.id AND spots.stage != 'completed') AS active_spots
FROM pieces
WHERE user_id = ? AND stage = 'active';

-- name: ListPiecesWithNewSpotsForPlan :many
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
            WHERE practice_plan_spots.practice_plan_id = :plan_id)) AS new_spots_count
FROM pieces
WHERE user_id = :user_id AND stage = 'active' AND new_spots_count > 0;


-- name: UpdatePiece :one
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
RETURNING *;

-- name: UpdatePiecePracticed :exec
UPDATE pieces
SET last_practiced = unixepoch('now')
WHERE pieces.id = :piece_id AND user_id = :user_id;

-- name: DeletePiece :exec
DELETE FROM pieces
WHERE id = ? AND user_id = ?
