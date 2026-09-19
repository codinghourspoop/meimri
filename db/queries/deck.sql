
-- Create a post
INSERT INTO posts (user_id, title, body)
VALUES (?, ?, ?);

-- Get posts for a user
SELECT id, user_id, title, body, created_at
FROM posts
WHERE user_id = ?
ORDER BY created_at DESC;

-- Get one post
SELECT id, user_id, title, body, created_at
FROM posts
WHERE id = ?;
