-- name: CreateUser :one
INSERT INTO users (id, fullname, email) VALUES (?, ?, ?)
RETURNING *;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = LOWER(:email);

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = ?;

-- name: GetUserForLogin :one
SELECT
    users.id,
    users.fullname,
    users.email,
    users.email_verified,
    COUNT(credentials.credential_id) AS credential_count
FROM users
LEFT JOIN credentials ON users.id = credentials.user_id
WHERE users.email = LOWER(:email);

-- name: UpdateUser :one
UPDATE users
SET fullname = COALESCE(?, fullname),
    email = COALESCE(?, email),
    email_verified = COALESCE(?, email_verified)
WHERE id = ?
RETURNING *;

-- name: SetEmailVerified :exec
UPDATE users SET email_verified = 1 WHERE id = ?
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: CreateCredential :one
INSERT INTO credentials (
    credential_id,
    public_key,
    transport,
    attestation_type,
    flags,
    authenticator,
    user_id
) VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetUserCredentials :many
SELECT *
FROM credentials
WHERE user_id = ?;

-- name: CountUserCredentials :one
SELECT COUNT(*) FROM credentials WHERE user_id = ?;

-- name: DeleteUserCredentials :exec
DELETE FROM credentials WHERE user_id = ?;
