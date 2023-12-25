-- name: CreatePracticeSession :exec
INSERT INTO practice_sessions (
    id,
    duration_minutes,
    date,
    user_id
) VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetPracticeSession :one
SELECT practice_sessions.*,
    practice_piece.measures AS practice_piece_measures,
    pieces.title AS piece_title,
    pieces.id AS piece_id,
    pieces.composer AS piece_composer,
    spots.name AS spot_name,
    spots.id AS spot_id,
    spots.measures AS spot_measures,
    spots.piece_id AS spot_piece_id,
    (SELECT pieces.title FROM pieces WHERE pieces.id = spots.piece_id) AS spot_piece_title
FROM practice_sessions
LEFT JOIN practice_piece ON practice_sessions.id = practice_piece.practice_session_id
LEFT JOIN pieces ON practice_piece.piece_id = pieces.id
LEFT JOIN practice_spot ON practice_sessions.id = practice_spot.practice_session_id
LEFT JOIN spots ON practice_spot.spot_id = spots.id
WHERE practice_sessions.id = :practice_session_id AND practice_sessions.user_id = :user_id;


-- name: PracticePiece :exec
INSERT INTO practice_piece (
    piece_id,
    practice_session_id,
    measures
) VALUES (
    (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1),
    ?, ?
);

-- name: CreatePracticeSpot :exec
INSERT INTO practice_spot (
    spot_id,
    practice_session_id
) VALUES (
    (SELECT spots.id FROM spots WHERE spots.piece_id IN (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id) AND spots.id = :spot_id),
    ?
);

-- name: AddRepToPracticeSpot :exec
UPDATE practice_spot
SET
    reps = reps + 1
WHERE spot_id = (SELECT spots.id FROM spots WHERE spots.piece_id IN (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id) AND spots.id = :spot_id)
AND practice_session_id = ?;

-- name: ListRecentPracticeSessions :many
SELECT practice_sessions.*,
    practice_piece.measures AS practice_piece_measures,
    pieces.title AS piece_title,
    pieces.id AS piece_id,
    pieces.composer AS piece_composer,
    spots.name AS spot_name,
    spots.id AS spot_id,
    spots.measures AS spot_measures,
    spots.piece_id AS spot_piece_id,
    (SELECT pieces.title FROM pieces WHERE pieces.id = spots.piece_id) AS spot_piece_title
FROM practice_sessions
LEFT JOIN practice_piece ON practice_sessions.id = practice_piece.practice_session_id
LEFT JOIN pieces ON practice_piece.piece_id = pieces.id
LEFT JOIN practice_spot ON practice_sessions.id = practice_spot.practice_session_id
LEFT JOIN spots ON practice_spot.spot_id = spots.id
WHERE practice_sessions.user_id = :user_id
AND practice_sessions.date >= (unixepoch('now') - 4 * 24 * 60 * 60)
AND practice_sessions.duration_minutes > 0
ORDER BY date DESC;

-- name: ListPracticeSessions :many
SELECT practice_sessions.*,
    practice_piece.measures AS practice_piece_measures,
    pieces.title AS piece_title,
    pieces.id AS piece_id,
    pieces.composer AS piece_composer,
    spots.name AS spot_name,
    spots.id AS spot_id,
    spots.measures AS spot_measures,
    spots.piece_id AS spot_piece_id,
    (SELECT pieces.title FROM pieces WHERE pieces.id = spots.piece_id) AS spot_piece_title
FROM practice_sessions
LEFT JOIN practice_piece ON practice_sessions.id = practice_piece.practice_session_id
LEFT JOIN pieces ON practice_piece.piece_id = pieces.id
LEFT JOIN practice_spot ON practice_sessions.id = practice_spot.practice_session_id
LEFT JOIN spots ON practice_spot.spot_id = spots.id
WHERE practice_sessions.user_id = :user_id AND practice_sessions.date >= (unixepoch('now') - :page * 14 * 24 * 60 * 60)
AND practice_sessions.date < (unixepoch('now') - (:page - 1) * 14 * 24 * 60 * 60)
AND practice_sessions.duration_minutes > 0
ORDER BY date DESC;

-- name: HasMorePracticeSessions :one
SELECT COUNT(id) > 0
FROM practice_sessions
WHERE practice_sessions.user_id = :user_id AND practice_sessions.date >= (unixepoch('now') - (:page + 1) * 14 * 24 * 60 * 60)
AND practice_sessions.date < (unixepoch('now') - :page * 14 * 24 * 60 * 60)
ORDER BY date DESC;

-- name: ListRecentPracticeSessionsForPiece :many
SELECT practice_sessions.*,
    practice_piece.measures AS practice_piece_measures
FROM practice_sessions
LEFT JOIN practice_piece ON practice_sessions.id = practice_piece.practice_session_id
LEFT JOIN pieces ON practice_piece.piece_id = pieces.id
WHERE practice_sessions.user_id = :user_id AND practice_piece.piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1)
AND practice_sessions.duration_minutes > 0
ORDER BY date DESC
LIMIT 10;

-- name: ListRecentPracticeSessionsForPieceSpots :many
SELECT practice_sessions.*,
    spots.name as spot_name,
    spots.id as spot_id
FROM practice_sessions
LEFT JOIN practice_spot ON practice_sessions.id = practice_spot.practice_session_id
LEFT JOIN spots ON practice_spot.spot_id = spots.id
WHERE practice_sessions.user_id = :user_id AND spots.piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1)
AND practice_sessions.duration_minutes > 0
ORDER BY date DESC
LIMIT 10;

-- name: ExtendPracticeSessionToNow :exec
UPDATE practice_sessions
SET duration_minutes = (unixepoch('now') - practice_sessions.date) / 60
WHERE id = ? AND user_id = ?;
