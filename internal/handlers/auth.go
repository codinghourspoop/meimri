package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"gitlab.com/coding4hours/memri/internal/auth"
)

type userJSON struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

// GET /api/me — session probe for the SPA.
func (a *App) Me(w http.ResponseWriter, r *http.Request) {
	uid, username, displayName, unread := a.currentUser(r)
	if uid == 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user":   userJSON{ID: uid, Username: username, DisplayName: displayName},
		"unread": unread,
	})
}

func (a *App) DoLogin(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	var id int64
	var hash, username, displayName string
	err := a.DB.QueryRow(`SELECT id, password_hash, username, display_name FROM user WHERE username = ? COLLATE NOCASE`, input.Username).Scan(&id, &hash, &username, &displayName)
	if err == sql.ErrNoRows {
		_ = auth.VerifyPassword("$argon2id$v=19$m=19456,t=2,p=1$dummy$dummy", input.Password)
		writeError(w, http.StatusUnauthorized, "Invalid username or password.")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !auth.VerifyPassword(hash, input.Password) {
		writeError(w, http.StatusUnauthorized, "Invalid username or password.")
		return
	}
	token, err := a.AuthStore.CreateSession(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	auth.SetSessionCookie(w, r, token)
	writeJSON(w, http.StatusOK, map[string]any{
		"user":   userJSON{ID: id, Username: username, DisplayName: displayName},
		"unread": a.unreadCount(id),
	})
}

func (a *App) DoRegister(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Password    string `json:"password"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if !auth.ValidUsername(input.Username) {
		writeError(w, http.StatusBadRequest, "Invalid username. Use 3-32 chars: letters, numbers, _ or -.")
		return
	}
	if !auth.ValidPassword(input.Password) {
		writeError(w, http.StatusBadRequest, "Password must be at least 8 characters.")
		return
	}
	if input.DisplayName == "" {
		input.DisplayName = input.Username
	}
	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	var userID int64
	err = a.DB.QueryRow(
		`INSERT INTO user (username, password_hash, display_name, created_at) VALUES (?, ?, ?, ?) RETURNING id`,
		input.Username, hash, input.DisplayName, time.Now().Unix(),
	).Scan(&userID)
	if err != nil {
		writeError(w, http.StatusConflict, "Username already taken.")
		return
	}
	codes, _ := auth.GenerateRecoveryCodes(recoveryCodeCount)
	for _, c := range codes {
		h, _ := auth.HashPassword(c)
		_, _ = a.DB.Exec(`INSERT INTO recovery_codes (user_id, code_hash) VALUES (?, ?)`, userID, h)
	}
	token, _ := a.AuthStore.CreateSession(r.Context(), userID)
	auth.SetSessionCookie(w, r, token)
	writeJSON(w, http.StatusCreated, map[string]any{
		"user":           userJSON{ID: userID, Username: input.Username, DisplayName: input.DisplayName},
		"recovery_codes": codes,
	})
}

func (a *App) DoLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(auth.SessionCookieName); err == nil {
		if sess, err := a.AuthStore.ValidateSession(c.Value); err == nil {
			_ = a.AuthStore.DeleteSession(sess.ID)
		}
	}
	auth.ClearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *App) DoRecover(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Code     string `json:"code"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if !auth.ValidPassword(input.Password) {
		writeError(w, http.StatusBadRequest, "Password must be at least 8 characters.")
		return
	}
	var userID int64
	if err := a.DB.QueryRow(`SELECT id FROM user WHERE username = ? COLLATE NOCASE`, input.Username).Scan(&userID); err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid username or code.")
		return
	}
	rows, err := a.DB.Query(`SELECT id, code_hash, used_at FROM recovery_codes WHERE user_id = ?`, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer func() { _ = rows.Close() }()
	var matchedID *int64
	for rows.Next() {
		var id int64
		var hash string
		var usedAt sql.NullInt64
		if err := rows.Scan(&id, &hash, &usedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		if usedAt.Valid {
			continue
		}
		if auth.VerifyPassword(hash, input.Code) {
			matchedID = &id
			break
		}
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if matchedID == nil {
		writeError(w, http.StatusUnauthorized, "Invalid username or code.")
		return
	}
	hash, _ := auth.HashPassword(input.Password)
	_, _ = a.DB.Exec(`UPDATE user SET password_hash = ? WHERE id = ?`, hash, userID)
	_, _ = a.DB.Exec(`UPDATE recovery_codes SET used_at = ? WHERE id = ?`, time.Now().Unix(), *matchedID)
	_ = a.AuthStore.DeleteUserSessions(r.Context(), userID)
	token, _ := a.AuthStore.CreateSession(r.Context(), userID)
	auth.SetSessionCookie(w, r, token)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
