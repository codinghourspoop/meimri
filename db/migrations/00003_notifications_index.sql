-- +goose Up
CREATE INDEX notifications_user_created_idx ON notifications(user_id, created_at);

CREATE INDEX notifications_user_unread_idx ON notifications(user_id) WHERE is_read = 0;

-- +goose Down
DROP INDEX IF EXISTS notifications_user_unread_idx;
DROP INDEX IF EXISTS notifications_user_created_idx;
