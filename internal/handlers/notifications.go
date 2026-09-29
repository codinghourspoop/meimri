package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"gitlab.com/coding4hours/memri/internal/auth"
)

type notificationJSON struct {
	Body   string `json:"body"`
	IsRead bool   `json:"is_read"`
}

// CreateNotification inserts one notification for a single user.
func (a *App) CreateNotification(ctx context.Context, userID int64, body string) error {
	_, err := a.DB.ExecContext(ctx,
		`INSERT INTO notifications (user_id, body, created_at) VALUES (?, ?, ?)`,
		userID, body, time.Now().Unix(),
	)
	return err
}

// CreateNotificationForAll inserts the same notification for every registered
// user and returns how many were created. A user who already has an identical
// notification from today is skipped, so a caller that fires twice (a restart, a
// DST shift) cannot spam anyone.
func (a *App) CreateNotificationForAll(ctx context.Context, body string) (int64, error) {
	now := time.Now()
	res, err := a.DB.ExecContext(ctx,
		`INSERT INTO notifications (user_id, body, created_at)
		 SELECT u.id, ?, ? FROM user u
		 WHERE NOT EXISTS (
		   SELECT 1 FROM notifications n
		   WHERE n.user_id = u.id AND n.body = ? AND n.created_at >= ?
		 )`,
		body, now.Unix(), body, startOfDay(now),
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func startOfDay(t time.Time) int64 {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location()).Unix()
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

// POST /api/notifications — add a notification for the current user.
func (a *App) AddNotification(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r)
	if uid == 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var input struct {
		Body string `json:"body"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	input.Body = strings.TrimSpace(input.Body)
	if input.Body == "" {
		writeError(w, http.StatusBadRequest, "body is required")
		return
	}
	if err := a.CreateNotification(r.Context(), uid, input.Body); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add notification")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]bool{"ok": true})
}

// POST /api/notifications/read
func (a *App) MarkNotificationsRead(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r)
	if uid == 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	_, _ = a.DB.ExecContext(r.Context(), `UPDATE notifications SET is_read = 1 WHERE user_id = ? AND is_read = 0`, uid)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
