package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"

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
	r.Use(auth.CSRF)
	r.Use(store.LoginLoader)

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	r.Get("/", app.Home)
	r.Get("/login", app.ShowLogin)
	r.Post("/login", app.DoLogin)
	r.Get("/register", app.ShowRegister)
	r.Post("/register", app.DoRegister)

	r.Get("/decks/new", app.NewDeck)
	r.Get("/decks/{id}/learn", app.LearnView)
	r.Get("/decks/{id}", app.Bob)
	r.Get("/api/learn", app.ReadLearn)
	r.Post("/api/learn", app.ReadLearn)
	r.Post("/api/decks", app.ReadDeck)
	r.Post("/api/decks/search", app.SearchDecks)
	r.Post("/api/decks/create", app.CreateDeck)
	r.Post("/a", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	r.Get("/recover", app.ShowRecover)
	r.Post("/recover", app.DoRecover)
	r.Post("/logout", app.DoLogout)
	r.Get("/notifications", app.Notifications)
	r.Post("/notifications/read", app.MarkNotificationsRead)
	r.Get("/fragment/greeting", app.GreetingFragment)

	r.Group(func(r chi.Router) {
		r.Use(auth.RequireUser)
		r.Get("/dashboard", app.Dashboard)
	})

	addr := env("ADDR", ":8080")
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
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
	if strings.HasPrefix(dsn, "file:") {
		dsn = strings.TrimPrefix(dsn, "file:")
	}
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
