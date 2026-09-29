package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"gitlab.com/coding4hours/memri/internal/auth"
)

// Bounds on the number of correct answers a card needs before it counts as
// mastered. The upper bound keeps a session from turning into an endurance run.
const (
	minDefaultPasses = 1
	maxDefaultPasses = 10

	defaultPasses = 3
	defaultTheme  = "system"
)

// settingsJSON is the wire shape of a user's saved preferences. Username is
// carried along read-only so the settings page can render the whole account
// section from one request instead of also reading the session store.
type settingsJSON struct {
	Username      string `json:"username"`
	DisplayName   string `json:"display_name"`
	DefaultPasses int    `json:"default_passes"`
	Theme         string `json:"theme"`
}

func validTheme(theme string) bool {
	switch theme {
	case "light", "dark", "system":
		return true
	}
	return false
}

// readSettings loads a user's preferences. The LEFT JOIN means an account that
// has never saved any gets the column defaults rather than an empty result, so
// callers never have to distinguish "no row" from "no settings".
func (a *App) readSettings(userID int64) (settingsJSON, error) {
	var s settingsJSON
	err := a.DB.QueryRow(
		`SELECT u.username, u.display_name, COALESCE(s.default_passes, ?), COALESCE(s.theme, ?)
		   FROM user u
		   LEFT JOIN user_settings s ON s.user_id = u.id
		  WHERE u.id = ?`,
		defaultPasses, defaultTheme, userID,
	).Scan(&s.Username, &s.DisplayName, &s.DefaultPasses, &s.Theme)
	return s, err
}

// GET /api/settings
func (a *App) ReadSettings(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r)
	if uid == 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	s, err := a.readSettings(uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read settings")
		return
	}
	writeJSON(w, http.StatusOK, s)
}

// PUT /api/settings — replaces every preference at once. A full replace keeps
// the stored row and the response body trivially in sync: whatever the client
// sends is what comes back, clamped or rejected.
func (a *App) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r)
	if uid == 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var input struct {
		DisplayName   string `json:"display_name"`
		DefaultPasses int    `json:"default_passes"`
		Theme         string `json:"theme"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	input.DisplayName = strings.TrimSpace(input.DisplayName)
	if !auth.ValidDisplayName(input.DisplayName) {
		writeError(w, http.StatusBadRequest,
			"Display name must be 1-"+strconv.Itoa(auth.MaxDisplayNameLength)+" characters with no leading or trailing space.")
		return
	}
	if input.DefaultPasses < minDefaultPasses || input.DefaultPasses > maxDefaultPasses {
		writeError(w, http.StatusBadRequest, "Passes to master must be between 1 and 10.")
		return
	}
	if !validTheme(input.Theme) {
		writeError(w, http.StatusBadRequest, `Theme must be "light", "dark", or "system".`)
		return
	}

	if _, err := a.DB.ExecContext(r.Context(),
		`INSERT INTO user_settings (user_id, default_passes, theme) VALUES (?, ?, ?)
		 ON CONFLICT(user_id) DO UPDATE SET
		   default_passes = excluded.default_passes,
		   theme          = excluded.theme`,
		uid, input.DefaultPasses, input.Theme,
	); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save settings")
		return
	}
	if _, err := a.DB.ExecContext(r.Context(),
		`UPDATE user SET display_name = ? WHERE id = ?`, input.DisplayName, uid,
	); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save settings")
		return
	}

	s, err := a.readSettings(uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read settings")
		return
	}
	writeJSON(w, http.StatusOK, s)
}

// POST /api/auth/recovery-codes — issue a fresh set of one-time password-reset
// codes. Every previous code for the account is deleted, spent ones included:
// a code that has already been redeemed must not become redeemable again.
func (a *App) RegenerateRecoveryCodes(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r)
	if uid == 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	codes, err := auth.GenerateRecoveryCodes(recoveryCodeCount)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	tx, err := a.DB.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to regenerate recovery codes")
		return
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(r.Context(), `DELETE FROM recovery_codes WHERE user_id = ?`, uid); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to regenerate recovery codes")
		return
	}
	for _, c := range codes {
		h, err := auth.HashPassword(c)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		if _, err := tx.ExecContext(r.Context(),
			`INSERT INTO recovery_codes (user_id, code_hash) VALUES (?, ?)`, uid, h,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to regenerate recovery codes")
			return
		}
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to regenerate recovery codes")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"recovery_codes": codes})
}
