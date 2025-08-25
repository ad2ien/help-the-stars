-- name: GetIssue :one
SELECT
  *
FROM
  issues
WHERE
  url = ?
LIMIT
  1;

-- name: ListIssues :many
SELECT
  *
FROM
  issues
ORDER BY
  creation_date DESC;

-- name: CreateIssue :one
INSERT INTO
  issues (url, repo_with_owner, title, description, creation_date, repo_description, stargazers_count)
VALUES
  (?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: UpdateIssue :exec
UPDATE issues
set
  repo_with_owner = ?,
  title = ?,
  description = ?,
  creation_date = ?,
  repo_description = ?,
  stargazers_count = ?
WHERE
  url = ?;

-- name: DeleteIssue :exec
DELETE FROM issues
WHERE
  url = ?;

-- name: GetTaskData :one
SELECT
  *
FROM
  task_data
LIMIT
  1;

-- name: UpdateTaskData :exec
UPDATE task_data
SET
  last_run = ?
WHERE
  id = 1;

-- name: CreateTaskData :one
INSERT INTO
  task_data (id, last_run)
VALUES
  (1, ?) RETURNING *;

-- name: DeleteTaskData :exec
DELETE FROM task_data
WHERE
  id = 1;