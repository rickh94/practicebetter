-- name: CreateSightReadingItem :one
INSERT INTO reading (
    id,
    title,
    info,
    composer,
    user_id
) VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetReadingByID :one
SELECT *
FROM reading
WHERE id = :reading_id AND user_id = :user_id
LIMIT 1;


-- name: CountUserReadingItems :one
SELECT COUNT(*) FROM reading WHERE user_id = ? LIMIT 1;

-- name: ListAllUserReadingItems :many
SELECT *
FROM reading
WHERE user_id = ?
ORDER BY title;

-- name: ListPaginatedUserReadingItems :many
SELECT *
FROM reading
WHERE user_id = ?
ORDER BY title
LIMIT ? OFFSET ?;

-- name: ListIncompleteUserReadingItems :many
SELECT *
FROM reading
WHERE user_id = ? AND completed = false
ORDER BY title;

-- name: UpdateReading :one
UPDATE reading
SET
    title = ?,
    info = ?,
    composer = ?
WHERE id = ? AND user_id = ?
RETURNING *;

-- name: CompleteReading :one
UPDATE reading
SET completed = true
WHERE id = ? AND user_id = ?
RETURNING *;

-- name: DeleteReadingItem :exec
DELETE FROM reading
WHERE id = ? AND user_id = ?;
