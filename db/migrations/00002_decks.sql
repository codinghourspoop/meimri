
-- +goose Up
CREATE TABLE deck (
    id          INTEGER PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    cards       TEXT NOT NULL DEFAULT '',
    created_at  INTEGER NOT NULL DEFAULT (unixepoch())
) STRICT;

CREATE INDEX deck_user_id_idx ON deck(user_id);

-- +goose Down
DROP INDEX IF EXISTS deck_user_id_idx;
DROP TABLE IF EXISTS deck;
