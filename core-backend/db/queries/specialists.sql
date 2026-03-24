-- name: CreateSpecialist :one
INSERT INTO specialists (
    full_name,
    personnel_number
) VALUES (
    $1, $2
)
RETURNING id, full_name, personnel_number, created_at, updated_at;

-- name: ListSpecialists :many
SELECT id, full_name, personnel_number, created_at, updated_at
FROM specialists
ORDER BY id DESC;

-- name: GetSpecialistByID :one
SELECT id, full_name, personnel_number, created_at, updated_at
FROM specialists
WHERE id = $1;

-- name: UpdateSpecialist :one
UPDATE specialists
SET
    full_name = $2,
    personnel_number = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING id, full_name, personnel_number, created_at, updated_at;

-- name: DeleteSpecialist :execrows
DELETE FROM specialists
WHERE id = $1;

