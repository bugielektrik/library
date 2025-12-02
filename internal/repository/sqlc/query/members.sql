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