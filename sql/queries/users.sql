-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (gen_random_uuid(), $1, $2, $3, $4)
RETURNING *;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUserEmailAndPassword :one
UPDATE users
SET updated_at = $1, email = $2, hashed_password = $3
WHERE id = $4
RETURNING *;
