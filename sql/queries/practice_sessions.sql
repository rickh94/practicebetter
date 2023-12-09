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

-- name: PracticeSpot :exec
INSERT INTO practice_spot (
    spot_id,
    practice_session_id
) VALUES (
    (SELECT spots.id FROM spots WHERE spots.piece_id = (SELECT pieces.id FROM pieces WHERE pieces.user_id = :user_id AND pieces.id = :piece_id LIMIT 1) AND spots.id = :spot_id),
    ?
);


-- name: GetRecentPracticeSessions :many
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
WHERE practice_sessions.user_id = :user_id AND practice_sessions.date <= (unixepoch('now') - 7 * 24 * 60 * 60)
ORDER BY date DESC;
