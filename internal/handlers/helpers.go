package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"gitlab.com/coding4hours/memri/internal/auth"
)

type App struct {
	DB        *sql.DB
	AuthStore *auth.Store
}

type CardInput struct {
	Front string `json:"front"`
	Back  string `json:"back"`
}

// recoveryCodeCount is how many one-time codes an account is issued at
// registration and again on every regeneration.
const recoveryCodeCount = 8

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func (a *App) currentUser(r *http.Request) (id int64, username, displayName string, unread int64) {
	uid := auth.UserID(r)
	if uid == 0 {
		return 0, "", "", 0
	}
	var u struct {
		Username    string
		DisplayName string
	}
	_ = a.DB.QueryRow(`SELECT username, display_name FROM user WHERE id = ?`, uid).Scan(&u.Username, &u.DisplayName)
	return uid, u.Username, u.DisplayName, a.unreadCount(uid)
}

// unreadCount reports a user's unread notifications. It is used directly by the
// login handler, where the request has no session loaded yet.
func (a *App) unreadCount(userID int64) int64 {
	var c int64
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = 0`, userID).Scan(&c)
	return c
}
