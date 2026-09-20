package handlers

import (
	"net/http"

	"gitlab.com/coding4hours/memri/internal/auth"
)

type notificationJSON struct {
	Body   string `json:"body"`
	IsRead bool   `json:"is_read"`
}

// GET /api/notifications
func (a *App) Notifications(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r)
	if uid == 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	rows, err := a.DB.Query(`SELECT body, is_read FROM notifications WHERE user_id = ? ORDER BY created_at DESC LIMIT 50`, uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read notifications")
		return
	}
	defer func() { _ = rows.Close() }()

	items := make([]notificationJSON, 0)
	for rows.Next() {
		var body string
		var isRead int64
		if err := rows.Scan(&body, &isRead); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read notifications")
			return
		}
		items = append(items, notificationJSON{Body: body, IsRead: isRead != 0})
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read notifications")
		return
	}
	_, _, _, unread := a.currentUser(r)
	writeJSON(w, http.StatusOK, map[string]any{
		"notifications": items,
		"unread":        unread,
	})
}

// POST /api/notifications/read
func (a *App) MarkNotificationsRead(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r)
	if uid == 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	_, _ = a.DB.Exec(`UPDATE notifications SET is_read = 1 WHERE user_id = ? AND is_read = 0`, uid)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
