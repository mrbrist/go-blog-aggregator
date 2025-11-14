-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id) VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
    )
    RETURNING *
)
SELECT
    inserted_feed_follow.*,
    users.name AS user_name,
    feeds.name AS feed_name
FROM inserted_feed_follow
INNER JOIN users ON users.id = inserted_feed_follow.user_id
INNER JOIN feeds ON feeds.id = inserted_feed_follow.feed_id;

-- name: GetFeedFollowsForUser :many
SELECT 
    feed_follows.id AS follow_id,
    feed_follows.created_at AS follow_created_at,
    feed_follows.updated_at AS follow_updated_at,
    feed_follows.user_id,
    feed_follows.feed_id,
    feeds.id AS feed_id,
    feeds.name AS feed_name,
    feeds.url AS feed_url,
    feeds.created_at AS feed_created_at,
    feeds.updated_at AS feed_updated_at
FROM feed_follows
JOIN feeds ON feeds.id = feed_follows.feed_id
WHERE feed_follows.user_id = $1;

-- name: UnfollowFeed :one
DELETE FROM feed_follows WHERE feed_id = $1 AND user_id = $2
RETURNING *;