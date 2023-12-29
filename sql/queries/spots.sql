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
    measures,
    stage_started
) VALUES (
    (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1),
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
    unixepoch('now')
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

-- name: GetSpotStageStarted :one
SELECT
    stage_started,
    stage
FROM spots
WHERE spots.id = :spot_id AND spots.piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1)
LIMIT 1;


-- name: UpdateSpot :exec
UPDATE spots
SET
    name = ?,
    idx = ?,
    stage = ?,
    stage_started = ?,
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

-- name: FixSpotStageStarted :exec
UPDATE spots
SET
    stage_started = unixepoch('now')
WHERE spots.id = :spot_id AND piece_id IN (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id);

-- name: PromoteSpotToRandom :exec
UPDATE spots
SET
    stage = CASE WHEN stage = 'repeat' OR stage = 'extra_repeat' THEN 'random' ELSE stage END,
    stage_started = CASE WHEN stage = 'repeat' OR stage = 'extra_repeat' THEN unixepoch('now') ELSE stage_started END,
    last_practiced = unixepoch('now')
WHERE spots.id = :spot_id AND piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1);

-- name: PromoteSpotToExtraRepeat :exec
UPDATE spots
SET
    stage = CASE WHEN stage = 'repeat' THEN 'extra_repeat' ELSE stage END,
    stage_started = CASE WHEN stage = 'repeat' THEN unixepoch('now') ELSE stage_started END,
    last_practiced = unixepoch('now')
WHERE spots.id = :spot_id AND piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1);

-- name: DemoteSpotToExtraRepeat :exec
UPDATE spots
SET
    stage = CASE WHEN stage != 'extra_repeat' AND stage != 'repeat' THEN 'extra_repeat' ELSE stage END,
    stage_started = CASE WHEN stage != 'extra_repeat' AND stage != 'repeat' THEN unixepoch('now') ELSE stage_started END,
    last_practiced = unixepoch('now')
WHERE spots.id = :spot_id AND piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1);

-- name: PromoteSpotToInterleave :exec
UPDATE spots
SET
    stage = CASE WHEN stage = 'random' THEN 'interleave' ELSE stage END,
    stage_started = CASE WHEN stage = 'random' THEN unixepoch('now') ELSE stage_started END,
    last_practiced = unixepoch('now')
WHERE spots.id = :spot_id AND piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1);

-- name: PromoteSpotToInterleaveDays :exec
UPDATE spots
SET
    stage = CASE WHEN stage = 'interleave' THEN 'interleave_days' ELSE stage END,
    stage_started = CASE WHEN stage = 'interleave' THEN unixepoch('now') ELSE stage_started END,
    skip_days = 1,
    last_practiced = unixepoch('now')
WHERE spots.id = :spot_id AND piece_id IN (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id);

-- name: DemoteSpotToRandom :exec
UPDATE spots
SET
    stage = CASE WHEN stage = 'interleave' THEN 'random' ELSE stage END,
    stage_started = CASE WHEN stage = 'interleave' THEN unixepoch('now') ELSE stage_started END,
    last_practiced = unixepoch('now')
WHERE spots.id = :spot_id AND piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id);

-- name: DemoteSpotToInterleave :exec
UPDATE spots
SET
    stage = CASE WHEN stage = 'interleave_days' THEN 'interleave' ELSE stage END,
    stage_started = CASE WHEN stage = 'interleave_days' THEN unixepoch('now') ELSE stage_started END,
    skip_days = 1,
    last_practiced = unixepoch('now')
WHERE spots.id = :spot_id AND piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id);

-- name: PromoteSpotToCompleted :exec
UPDATE spots
SET
    stage = CASE WHEN stage = 'interleave_days' THEN 'completed' ELSE stage END,
    stage_started = CASE WHEN stage = 'interleave_days' THEN unixepoch('now') ELSE stage_started END,
    skip_days = 1,
    last_practiced = unixepoch('now')
WHERE spots.id = :spot_id AND piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id);

-- name: UpdateSpotSkipDays :exec
UPDATE spots
SET
    skip_days = :skip_days
WHERE spots.id = :spot_id AND piece_id IN (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id);

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
