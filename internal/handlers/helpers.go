package handlers

import (
	"database/sql"
	"net/http"

	"gitlab.com/coding4hours/memri/components"
	"gitlab.com/coding4hours/memri/internal/auth"
)

type App struct {
	DB        *sql.DB
	AuthStore *auth.Store
}

func isHX(r *http.Request) bool { return r.Header.Get("HX-Request") == "true" }

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
	var c int64
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = 0`, uid).Scan(&c)
	return uid, u.Username, u.DisplayName, c
}

func (a *App) renderNotificationBell(w http.ResponseWriter, r *http.Request) {
	_, _, _, unread := a.currentUser(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = components.NotificationsBellFragment(unread).Render(r.Context(), w)
}
