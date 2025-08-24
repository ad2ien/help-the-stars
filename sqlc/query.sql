-- name: GetIssue :one
SELECT
  *
FROM
  issues
WHERE
  id = ?
LIMIT
  1;

-- name: ListIssues :many
SELECT
  *
FROM
  issues
ORDER BY
  name;

-- name: CreateIssue :one
INSERT INTO
  issues (link, title, description, owner, creation_date)
VALUES
  (?, ?, ?, ?, ?) RETURNING *;

-- name: UpdateIssue :exec
UPDATE issues
set
  link = ?,
  title = ?,
  description = ?,
  owner = ?,
  creation_date = ?
WHERE
  id = ?;

-- name: DeleteIssue :exec
DELETE FROM issues
WHERE
  id = ?;