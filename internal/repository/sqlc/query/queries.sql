-- name: ListAuthors :many
SELECT * FROM authors ORDER BY id;

-- name: AddAuthor :one
INSERT INTO authors (full_name, pseudonym, specialty) VALUES ($1, $2, $3) RETURNING id;

-- name: GetAuthor :one
SELECT id, full_name, pseudonym, specialty FROM authors WHERE id = $1;

-- name: DeleteAuthor :exec
DELETE FROM authors WHERE id = $1;

-- name: UpdateAuthor :one
UPDATE authors
SET 
    full_name = $2,
    pseudonym = $3,
    specialty = $4
WHERE id = $1
RETURNING *;

-- name: ListBooks :many
SELECT * FROM books ORDER BY id;

-- name: AddBook :one
INSERT INTO books (name, genre, isbn, author_id) VALUES ($1, $2, $3, $4) returning id;

-- name: GetBook :one
SELECT * from books where id = $1;

-- name: UpdateBook :exec
UPDATE books
set name = $2, genre = $3, isbn = $4, author_id = $5
where id = $1;

-- name: DeleteBook :exec
DELETE FROM books WHERE id = $1;

-- name: ListMembers :many
SELECT * FROM members ORDER BY id;

-- name: AddMember :one
INSERT INTO members (id, full_name) VALUES ($1, $2) RETURNING id;

-- name: GetMember :one
SELECT id, full_name FROM members WHERE id = $1;

-- name: UpdateMember :one
UPDATE members
SET full_name = $2
WHERE id = $1
RETURNING *;

-- name: DeleteMember :exec
DELETE FROM members WHERE id = $1;

-- name: GetMemberBooks :many
SELECT book_id FROM members_and_books WHERE member_id = $1;

-- name: AddMemberBook :exec
INSERT INTO members_and_books (book_id, member_id) VALUES ($1, $2);

-- name: DeleteMemberBook :exec
DELETE FROM members_and_books WHERE book_id = $1 AND member_id = $2;

-- name: DeleteAllMemberBooks :exec
DELETE FROM members_and_books WHERE member_id = $1;

-- name: CreateUser :one
INSERT INTO users (id, email, password_hash, full_name, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, full_name, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, password_hash, full_name, created_at, updated_at
FROM users
WHERE id = $1;

