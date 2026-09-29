-- +goose Up
CREATE TABLE user_settings (
    user_id        INTEGER PRIMARY KEY REFERENCES user(id) ON DELETE CASCADE,
    default_passes INTEGER NOT NULL DEFAULT 3,
    theme          TEXT NOT NULL DEFAULT 'system'
) STRICT;

-- +goose Down
DROP TABLE IF EXISTS user_settings;