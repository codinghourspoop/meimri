-- +goose Up
CREATE TABLE user (
    id            INTEGER PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE COLLATE NOCASE,
    password_hash TEXT NOT NULL,
    display_name  TEXT NOT NULL,
    is_admin      INTEGER NOT NULL DEFAULT 0,
    created_at    INTEGER NOT NULL
) STRICT;

CREATE TABLE auth_session (
    id                     TEXT NOT NULL PRIMARY KEY,
    user_id                INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE,
    secret_hash            BLOB NOT NULL,
    token_last_verified_at INTEGER NOT NULL,
    created_at             INTEGER NOT NULL
) STRICT;

CREATE TABLE recovery_codes (
    id        INTEGER PRIMARY KEY,
    user_id   INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE,
    code_hash TEXT NOT NULL,
    used_at   INTEGER
) STRICT;

CREATE TABLE notifications (
    id         INTEGER PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE,
    body       TEXT NOT NULL,
    is_read    INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL
) STRICT;

-- +goose Down
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS recovery_codes;
DROP TABLE IF EXISTS auth_session;
DROP TABLE IF EXISTS user;
