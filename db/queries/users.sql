-- name: CreateUser :one
INSERT INTO users (
    email, full_name, phone, role, auth_provider
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC;

-- name: UpdateUserStatus :exec
UPDATE users
SET is_active = $2, updated_at = NOW()
WHERE id = $1;
