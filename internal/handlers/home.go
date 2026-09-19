package handlers

import (
	"net/http"
	"time"

	"gitlab.com/coding4hours/memri/components"

	"gitlab.com/coding4hours/memri/internal/auth"
)

func (a *App) Home(w http.ResponseWriter, r *http.Request) {
	uid, username, displayName, unread := a.currentUser(r)
	_ = components.HomePage(uid != 0, unread, username, displayName).Render(r.Context(), w)
}
func (a *App) Bob(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r)
	_, _, _, unread := a.currentUser(r)
	_ = components.FlashPage(uid != 0, unread).Render(r.Context(), w)
}

func (a *App) Dashboard(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) GreetingFragment(w http.ResponseWriter, r *http.Request) {
	_, username, displayName, _ := a.currentUser(r)
	name := displayName
	if name == "" {
		name = username
	}
	if name == "" {
		name = "guest"
	}
	if isHX(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = components.GreetingFragment(name+" at "+time.Now().Format(time.Kitchen)).Render(r.Context(), w)
		return
	}
	uid, _, _, unread := a.currentUser(r)
	_ = components.HomePage(uid != 0, unread, username, displayName).Render(r.Context(), w)
}

func (a *App) Notifications(w http.ResponseWriter, r *http.Request) {
	uid, _, _, unread := a.currentUser(r)
	if isHX(r) {
		a.renderNotificationBell(w, r)
		return
	}
	rows, _ := a.DB.Query(`SELECT body, is_read FROM notifications WHERE user_id = ? ORDER BY created_at DESC LIMIT 50`, uid)
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	var items []components.NotificationItem
	if rows != nil {
		for rows.Next() {
			var n components.NotificationItem
			_ = rows.Scan(&n.Body, &n.IsRead)
			items = append(items, n)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "failed to read notifications", http.StatusInternalServerError)
			return
		}
	}
	_ = components.NotificationsPage(items, unread).Render(r.Context(), w)
}

func (a *App) MarkNotificationsRead(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r)
	_, _ = a.DB.Exec(`UPDATE notifications SET is_read = 1 WHERE user_id = ? AND is_read = 0`, uid)
	http.Redirect(w, r, "/notifications", http.StatusSeeOther)
}
