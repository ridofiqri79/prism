-- name: GetUserByUsername :one
SELECT *
FROM app_user
WHERE username = $1
  AND is_active = true;

-- name: GetUserByID :one
SELECT *
FROM app_user
WHERE id = $1;

-- name: ListUsers :many
SELECT *
FROM app_user
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*)
FROM app_user;

-- name: CreateUser :one
INSERT INTO app_user (username, email, password_hash, role)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateUser :one
UPDATE app_user
SET username = $2,
    email = $3,
    role = $4,
    is_active = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetUserPermissions :many
SELECT *
FROM user_permission
WHERE user_id = $1
ORDER BY module ASC;

-- name: DeleteUserPermissions :exec
DELETE FROM user_permission
WHERE user_id = $1;

-- name: CreateUserPermission :one
INSERT INTO user_permission (user_id, module, can_create, can_read, can_update, can_delete)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserPermissionByModule :one
SELECT *
FROM user_permission
WHERE user_id = $1
  AND module = $2;
