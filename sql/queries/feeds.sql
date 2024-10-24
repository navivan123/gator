-- name: CreateFeed :one
INSERT INTO feeds  ( id, created_at, updated_at, name, url, user_id )
            VALUES ( $1, $2,         $3,         $4,   $5,  $6      )
RETURNING *;

-- name: GetFeeds :many
SELECT name, url, user_id FROM feeds;

-- name: GetFeedsWithName :many
SELECT feeds.name, feeds.url, users.name AS username FROM feeds
INNER JOIN users
ON feeds.user_id = users.id;

-- name: GetFeedUrl :one
SELECT * FROM feeds
WHERE feeds.url = $1;


-- name: MarkFeedFetched :one
UPDATE feeds
SET last_fetched_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;

-- name: DeleteFeeds :exec
DELETE FROM feeds;
