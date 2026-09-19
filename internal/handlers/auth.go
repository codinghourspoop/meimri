package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"gitlab.com/coding4hours/memri/components"
	"gitlab.com/coding4hours/memri/internal/auth"
)

func (a *App) ShowLogin(w http.ResponseWriter, r *http.Request) {
	_ = components.LoginPage("").Render(r.Context(), w)
}

func (a *App) DoLogin(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	var id int64
	var hash string
	err := a.DB.QueryRow(`SELECT id, password_hash FROM user WHERE username = ? COLLATE NOCASE`, username).Scan(&id, &hash)
	if err == sql.ErrNoRows {
		_ = auth.VerifyPassword("$argon2id$v=19$m=19456,t=2,p=1$dummy$dummy", password)
		_ = components.LoginPage("Invalid username or password.").Render(r.Context(), w)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !auth.VerifyPassword(hash, password) {
		_ = components.LoginPage("Invalid username or password.").Render(r.Context(), w)
		return
	}
	token, err := a.AuthStore.CreateSession(r.Context(), id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	auth.SetSessionCookie(w, r, token)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (a *App) ShowRegister(w http.ResponseWriter, r *http.Request) {
	_ = components.RegisterPage("").Render(r.Context(), w)
}

func (a *App) DoRegister(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	username := r.FormValue("username")
	displayName := r.FormValue("display_name")
	password := r.FormValue("password")

	if !auth.ValidUsername(username) {
		_ = components.RegisterPage("Invalid username. Use 3-32 chars: letters, numbers, _ or -.").Render(r.Context(), w)
		return
	}
	if !auth.ValidPassword(password) {
		_ = components.RegisterPage("Password must be at least 8 characters.").Render(r.Context(), w)
		return
	}
	if displayName == "" {
		displayName = username
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	var userID int64
	err = a.DB.QueryRow(
		`INSERT INTO user (username, password_hash, display_name, created_at) VALUES (?, ?, ?, ?) RETURNING id`,
		username, hash, displayName, time.Now().Unix(),
	).Scan(&userID)
	if err != nil {
		_ = components.RegisterPage("Username already taken.").Render(r.Context(), w)
		return
	}
	codes, _ := auth.GenerateRecoveryCodes(8)
	for _, c := range codes {
		h, _ := auth.HashPassword(c)
		_, _ = a.DB.Exec(`INSERT INTO recovery_codes (user_id, code_hash) VALUES (?, ?)`, userID, h)
	}
	token, _ := a.AuthStore.CreateSession(r.Context(), userID)
	auth.SetSessionCookie(w, r, token)
	_ = components.RecoveryCodesPage(codes).Render(r.Context(), w)
}

func (a *App) DoLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(auth.SessionCookieName); err == nil {
		if sess, err := a.AuthStore.ValidateSession(c.Value); err == nil {
			_ = a.AuthStore.DeleteSession(sess.ID)
		}
	}
	auth.ClearSessionCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) ShowRecover(w http.ResponseWriter, r *http.Request) {
	_ = components.RecoverPage("").Render(r.Context(), w)
}

func (a *App) DoRecover(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	username := r.FormValue("username")
	code := r.FormValue("code")
	newPassword := r.FormValue("password")

	if !auth.ValidPassword(newPassword) {
		_ = components.RecoverPage("Password must be at least 8 characters.").Render(r.Context(), w)
		return
	}
	var userID int64
	if err := a.DB.QueryRow(`SELECT id FROM user WHERE username = ? COLLATE NOCASE`, username).Scan(&userID); err != nil {
		_ = components.RecoverPage("Invalid username or code.").Render(r.Context(), w)
		return
	}
	rows, err := a.DB.Query(`SELECT id, code_hash, used_at FROM recovery_codes WHERE user_id = ?`, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()
	var matchedID *int64
	for rows.Next() {
		var id int64
		var hash string
		var usedAt sql.NullInt64
		if err := rows.Scan(&id, &hash, &usedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		if usedAt.Valid {
			continue
		}
		if auth.VerifyPassword(hash, code) {
			matchedID = &id
			break
		}
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if matchedID == nil {
		_ = components.RecoverPage("Invalid username or code.").Render(r.Context(), w)
		return
	}
	hash, _ := auth.HashPassword(newPassword)
	_, _ = a.DB.Exec(`UPDATE user SET password_hash = ? WHERE id = ?`, hash, userID)
	_, _ = a.DB.Exec(`UPDATE recovery_codes SET used_at = ? WHERE id = ?`, time.Now().Unix(), *matchedID)
	_ = a.AuthStore.DeleteUserSessions(r.Context(), userID)
	token, _ := a.AuthStore.CreateSession(r.Context(), userID)
	auth.SetSessionCookie(w, r, token)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
