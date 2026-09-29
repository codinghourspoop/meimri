package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	_ "turso.tech/database/tursogo"

	"gitlab.com/coding4hours/memri/internal/auth"
	"gitlab.com/coding4hours/memri/internal/handlers"
)

func main() {
	db := openTursoDB()
	defer func() { _ = db.Close() }()

	if err := goose.SetDialect("sqlite3"); err != nil {
		log.Fatal(err)
	}
	if err := goose.Up(db, "db/migrations"); err != nil {
		log.Printf("goose: %v", err)
	}

	store := auth.NewStore(db)
	app := &handlers.App{DB: db, AuthStore: store}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(store.LoginLoader)
	r.Use(auth.CSRF)

	// JSON API (consumed by the Svelte SPA).
	r.Get("/api/me", app.Me)
	r.Post("/api/auth/login", app.DoLogin)
	r.Post("/api/auth/register", app.DoRegister)
	r.Post("/api/auth/recover", app.DoRecover)
	r.Post("/api/auth/logout", app.DoLogout)
	r.Post("/api/auth/recovery-codes", app.RegenerateRecoveryCodes)
	r.Get("/api/settings", app.ReadSettings)
	r.Put("/api/settings", app.UpdateSettings)
	r.Get("/api/decks/search", app.SearchDecks)
	r.Post("/api/decks", app.CreateDeck)
	r.Get("/api/decks/{id}", app.ReadDeck)
	r.Get("/api/learn", app.ReadLearn)
	r.Get("/api/notifications", app.Notifications)
	r.Post("/api/notifications/read", app.MarkNotificationsRead)

	// Vite build output.
	r.Handle("/assets/*", http.StripPrefix("/", http.FileServer(http.Dir("static/dist"))))
	r.Get("/vite.svg", serveDistFile("vite.svg", "image/svg+xml"))

	// SPA fallback: every non-API GET serves the built index.html
	// (or the Vite dev placeholder when no build exists yet).
	r.Get("/*", spaHandler())

	addr := env("ADDR", ":8080")
	srv := &http.Server{Addr: addr, Handler: r}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server error: %v", err)
		}
	}()

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

func serveDistFile(name, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join("static", "dist", name)
		data, err := os.ReadFile(path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", contentType)
		_, _ = w.Write(data)
	}
}

func spaHandler() http.HandlerFunc {
	indexPath := filepath.Join("static", "dist", "index.html")
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		data, err := os.ReadFile(indexPath)
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<!doctype html><html><body style="font-family:sans-serif">` +
				`<h1>memri</h1><p>Frontend not built yet. Run <code>bun install &amp;&amp; bun run build</code> ` +
				`in <code>frontend/</code>, or <code>task dev:frontend</code> for the Vite dev server.</p>` +
				`</body></html>`))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
	}
}

func openTursoDB() *sql.DB {
	dsn := env("DATABASE_URL", "./data/app.db")
	if dsn == ":memory:" {
		db, err := sql.Open("turso", ":memory:")
		if err != nil {
			log.Fatal(err)
		}
		return db
	}
	dsn = strings.TrimPrefix(dsn, "file:")
	_ = os.MkdirAll("./data", 0755)
	db, err := sql.Open("turso", dsn)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func env(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
