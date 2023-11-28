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
    current_tempo
) VALUES (
    (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1),
    ?, ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: ListPieceSpots :many
SELECT
    id,
    name,
    idx,
    stage,
    audio_prompt_url,
    image_prompt_url,
    notes_prompt,
    text_prompt,
    current_tempo
FROM spots
WHERE piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = ? AND pieces.id = ? LIMIT 1);

-- name: GetSpot :one
SELECT spots.*
FROM spots
WHERE spots.id = ? AND spots.piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = ? AND pieces.id = ? LIMIT 1);

-- name: UpdateSpot :one
UPDATE spots
SET
    name = ?,
    idx = ?,
    stage = ?,
    audio_prompt_url = ?,
    image_prompt_url = ?,
    notes_prompt = ?,
    text_prompt = ?,
    current_tempo = ?
WHERE spots.id = :spot_id AND piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1)
RETURNING *;

-- name: DeleteSpot :exec
DELETE FROM spots
WHERE spots.id = ? AND spots.piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = ? AND pieces.id = ? LIMIT 1);

-- name: DeleteSpotsExcept :exec
DELETE FROM spots
WHERE
spots.piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1)
AND spots.id NOT IN (sqlc.slice('spotIDs'));
