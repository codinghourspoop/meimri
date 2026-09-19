package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"errors"
	"strings"
	"time"
)

const (
	SessionCookieName = "session"
	SessionExpiry     = 10 * 24 * time.Hour
	renewInterval     = time.Hour
	alphabet          = "abcdefghijkmnpqrstuvwxyz23456789"
)

type Session struct {
	ID     string
	UserID int64
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func generateID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, 16)
	for i, b := range buf {
		out[i] = alphabet[int(b>>3)]
	}
	return string(out), nil
}

func (s *Store) CreateSession(ctx context.Context, userID int64) (string, error) {
	id, err := generateID()
	if err != nil {
		return "", err
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	h := sha256.Sum256(secret)
	now := time.Now().Unix()
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO auth_session (id, user_id, secret_hash, token_last_verified_at, created_at) VALUES (?, ?, ?, ?, ?)`,
		id, userID, h[:], now, now)
	if err != nil {
		return "", err
	}
	return id + "." + base64.StdEncoding.EncodeToString(secret), nil
}

func (s *Store) ValidateSession(token string) (*Session, error) {
	idx := strings.IndexByte(token, '.')
	if idx < 0 {
		return nil, errors.New("invalid token format")
	}
	id := token[:idx]
	secretB64 := token[idx+1:]
	secret, err := base64.StdEncoding.DecodeString(secretB64)
	if err != nil || len(secret) != 32 {
		return nil, errors.New("invalid token secret")
	}
	h := sha256.Sum256(secret)
	var userID int64
	var storedHash []byte
	var lastVerified int64
	err = s.db.QueryRow(`SELECT user_id, secret_hash, token_last_verified_at FROM auth_session WHERE id = ?`, id).
		Scan(&userID, &storedHash, &lastVerified)
	if err != nil {
		return nil, err
	}
	if subtle.ConstantTimeCompare(h[:], storedHash) != 1 {
		return nil, errors.New("invalid token")
	}
	now := time.Now().Unix()
	if now-lastVerified > int64(SessionExpiry.Seconds()) {
		_, _ = s.db.Exec(`DELETE FROM auth_session WHERE id = ?`, id)
		return nil, errors.New("session expired")
	}
	if now-lastVerified > int64(renewInterval.Seconds()) {
		_, _ = s.db.Exec(`UPDATE auth_session SET token_last_verified_at = ? WHERE id = ?`, now, id)
	}
	return &Session{ID: id, UserID: userID}, nil
}

func (s *Store) DeleteSession(id string) error {
	_, err := s.db.Exec(`DELETE FROM auth_session WHERE id = ?`, id)
	return err
}

func (s *Store) DeleteUserSessions(ctx context.Context, userID int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM auth_session WHERE user_id = ?`, userID)
	return err
}

func GenerateRecoveryCodes(n int) ([]string, error) {
	codes := make([]string, n)
	for i := range codes {
		b := make([]byte, 12)
		if _, err := rand.Read(b); err != nil {
			return nil, err
		}
		codes[i] = base64.RawURLEncoding.EncodeToString(b)
	}
	return codes, nil
}
