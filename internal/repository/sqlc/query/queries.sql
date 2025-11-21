-- name: ListAuthors :many
SELECT * FROM authors ORDER BY id;

-- name: AddAuthor :one
INSERT INTO authors (full_name, pseudonym, specialty) VALUES ($1, $2, $3) RETURNING id;

-- name: GetAuthor :one
SELECT id, full_name, pseudonym, specialty FROM authors WHERE id = $1;

-- name: DeleteAuthor :one
DELETE FROM authors WHERE id = $1 RETURNING id;

-- name: UpdateAuthor :one
UPDATE authors
SET 
    full_name = $2,
    pseudonym = $3,
    specialty = $4
WHERE id = $1
RETURNING *;