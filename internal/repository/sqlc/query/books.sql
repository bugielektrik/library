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