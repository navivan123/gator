-- name: DeleteFeedFollows :exec
DELETE FROM feed_follows;

-- name: CreateFeedFollow :one
WITH feed_follow_insert AS (
    INSERT INTO feed_follows ( id, created_at, updated_at, user_id, feed_id )
                      VALUES ( $1, $2,         $3,         $4,      $5      )
    RETURNING * )

SELECT feed_follow_insert.*, feeds.name AS feed_name, users.name AS user_name
FROM   feed_follow_insert
INNER JOIN users ON users.id = feed_follow_insert.user_id
INNER JOIN feeds ON feeds.id = feed_follow_insert.feed_id;

-- name: GetFeedFollowsForUser :many
SELECT feed_follows.*, users.name AS user_name, feeds.name AS feed_name
FROM feed_follows
INNER JOIN feeds ON feed_follows.feed_id = feeds.id
INNER JOIN users ON feed_follows.user_id = users.id
WHERE users.name = $1;

-- name: DeleteFeedFollowsForUserUrl :exec
DELETE FROM feed_follows
WHERE feed_follows.user_id = $1 AND feed_follows.feed_id = $2;
