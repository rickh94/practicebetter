-- name: CreateSpot :one
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
    (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1),
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: ListPieceSpots :many
SELECT
    spots.*,
    pieces.title AS piece_title
FROM spots
INNER JOIN pieces ON pieces.id = spots.piece_id
WHERE piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1)
ORDER BY spots.idx;

-- name: ListHighPrioritySpots :many
SELECT
    spots.*,
    pieces.title AS piece_title
FROM spots
INNER JOIN pieces ON pieces.id = spots.piece_id
WHERE pieces.user_id = :user_id AND spots.priority < 0
ORDER BY spots.priority;

-- name: GetSpot :one
SELECT
    spots.*,
    pieces.title AS piece_title
FROM spots
INNER JOIN pieces ON pieces.id = spots.piece_id
WHERE spots.id = :spot_id AND spots.piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1);


-- name: UpdateSpot :exec
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
WHERE spots.id = :spot_id AND piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1);

-- name: UpdateTextPrompt :one
UPDATE spots
SET
    text_prompt = ?
WHERE spots.id = :spot_id AND piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1)
RETURNING *;

-- name: UpdateSpotPriority :exec
UPDATE spots
SET
    priority = :priority
WHERE spots.id = :spot_id AND piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1);

-- name: PromoteToRandom :exec
UPDATE spots
SET
    stage = CASE WHEN stage = 'repeat' OR stage = 'extra_repeat' THEN 'random' ELSE stage END,
    last_practiced = unixepoch('now')
WHERE spots.id = :spot_id AND piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1);

-- name: PromoteToMoreRepeat :exec
UPDATE spots
SET
    stage = CASE WHEN stage = 'repeat' THEN 'extra_repeat' ELSE stage END,
    last_practiced = unixepoch('now')
WHERE spots.id = :spot_id AND piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1);

-- name: UpdateSpotPracticed :exec
UPDATE spots
SET last_practiced = unixepoch('now')
WHERE spots.id = :spot_id AND spots.piece_id IN (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id);

-- name: DeleteSpot :exec
DELETE FROM spots
WHERE spots.id = :spot_id AND spots.piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1);

-- name: DeleteSpotsExcept :exec
DELETE FROM spots
WHERE
spots.piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1)
AND spots.id NOT IN (sqlc.slice('spotIDs'));
