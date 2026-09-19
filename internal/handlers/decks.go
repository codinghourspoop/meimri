package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"gitlab.com/coding4hours/memri/components"
	"gitlab.com/coding4hours/memri/internal/auth"
)

type createDeckRequests struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Cards       []components.CardInput `json:"cards"`
}

type deckResponse struct {
	ID          int64                  `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Cards       []components.CardInput `json:"cards"`
}

func (a *App) SearchDecks(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserID(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	q := strings.TrimSpace(r.FormValue("q"))
	if q == "" {
		q = strings.TrimSpace(r.FormValue("query"))
	}
	if q == "" {
		q = strings.TrimSpace(r.FormValue("search"))
	}

	var (
		rows *sql.Rows
		err  error
	)

	if q == "" {
		// No query: list the user's decks, newest first.
		rows, err = a.DB.QueryContext(
			r.Context(),
			`SELECT id, title, description, cards
			 FROM deck
			 WHERE user_id = ?
			 ORDER BY created_at DESC, id DESC
			 LIMIT 50`,
			userID,
		)
	} else {
		like := "%" + escapeLike(q) + "%"
		rows, err = a.DB.QueryContext(
			r.Context(),
			`SELECT id, title, description, cards
			 FROM deck
			 WHERE user_id = ?
			   AND (title LIKE ? ESCAPE '\' OR description LIKE ? ESCAPE '\')
			 ORDER BY created_at DESC, id DESC
			 LIMIT 50`,
			userID,
			like,
			like,
		)
	}
	if err != nil {
		log.Printf("SearchDecks database error: %v", err)
		http.Error(w, "failed to search decks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	decks := make([]deckResponse, 0)

	for rows.Next() {
		var deck deckResponse
		var cardsJSON []byte

		if err := rows.Scan(
			&deck.ID,
			&deck.Title,
			&deck.Description,
			&cardsJSON,
		); err != nil {
			log.Printf("SearchDecks scan error: %v", err)
			http.Error(w, "failed to scan deck", http.StatusInternalServerError)
			return
		}

		if len(cardsJSON) > 0 {
			if err := json.Unmarshal(cardsJSON, &deck.Cards); err != nil {
				log.Printf("SearchDecks decode cards error: %v", err)
				http.Error(w, "failed to decode cards", http.StatusInternalServerError)
				return
			}
		}
		if deck.Cards == nil {
			deck.Cards = []components.CardInput{}
		}

		decks = append(decks, deck)
	}

	if err := rows.Err(); err != nil {
		log.Printf("SearchDecks rows error: %v", err)
		http.Error(w, "failed to read decks", http.StatusInternalServerError)
		return
	}

	if isHX(r) {
		// htmx search box: empty query restores the hint, no DB dump.
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if q == "" {
			_ = components.DeckSearchHint().Render(r.Context(), w)
			return
		}
		results := make([]components.DeckSearchResult, 0, len(decks))
		for _, d := range decks {
			results = append(results, components.DeckSearchResult{
				ID:          d.ID,
				Title:       d.Title,
				Description: d.Description,
				CardCount:   len(d.Cards),
			})
		}
		if err := components.DeckSearchResults(results).Render(r.Context(), w); err != nil {
			log.Printf("SearchDecks render error: %v", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(decks); err != nil {
		log.Printf("SearchDecks encode error: %v", err)
	}
}

// escapeLike escapes \, % and _ so user input is treated literally in LIKE.
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

func (a *App) NewDeck(w http.ResponseWriter, r *http.Request) {
	uid, _, _, unread := a.currentUser(r)
	if uid == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	_ = components.NewDeckPage(true, unread).Render(r.Context(), w)
}

func (a *App) CreateDeck(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserID(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var input createDeckRequests

	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
	} else {
		// htmx form post from NewDeckPage: repeated front/back pairs.
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		input.Title = strings.TrimSpace(r.FormValue("title"))
		input.Description = strings.TrimSpace(r.FormValue("description"))
		fronts := r.Form["front"]
		backs := r.Form["back"]
		n := len(fronts)
		if len(backs) > n {
			n = len(backs)
		}
		for i := 0; i < n; i++ {
			var front, back string
			if i < len(fronts) {
				front = strings.TrimSpace(fronts[i])
			}
			if i < len(backs) {
				back = strings.TrimSpace(backs[i])
			}
			if front == "" && back == "" {
				continue
			}
			input.Cards = append(input.Cards, components.CardInput{Front: front, Back: back})
		}
	}

	if input.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	cardsJSON, err := json.Marshal(input.Cards)
	if err != nil {
		http.Error(w, "failed to encode cards", http.StatusInternalServerError)
		return
	}

	result, err := a.DB.ExecContext(
		r.Context(),
		`INSERT INTO deck (user_id, title, description, cards)
		 VALUES (?, ?, ?,?)`,
		userID,
		input.Title,
		input.Description,
		string(cardsJSON),
	)
	if err != nil {
		http.Error(w, "failed to create deck", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "failed to get deck ID", http.StatusInternalServerError)
		return
	}

	response := deckResponse{
		ID:          id,
		Title:       input.Title,
		Description: input.Description,
		Cards:       input.Cards,
	}

	if isHX(r) {
		// htmx form submit: navigate to the new deck.
		w.Header().Set("HX-Redirect", fmt.Sprintf("/decks/%d", id))
		w.WriteHeader(http.StatusCreated)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

func (a *App) ReadDeck(w http.ResponseWriter, r *http.Request) {
	rows, err := a.DB.QueryContext(
		r.Context(),
		`SELECT id, title, description, cards
		 FROM deck
		 WHERE id =?`,
		r.FormValue("deck_id"),
	)
	if err != nil {
		log.Printf("ReadDeck database error: %v", err)
		http.Error(w, "failed to read deck", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var decks []deckResponse

	for rows.Next() {
		var deck deckResponse
		var cardsJSON []byte

		if err := rows.Scan(
			&deck.ID,
			&deck.Title,
			&deck.Description,
			&cardsJSON,
		); err != nil {
			http.Error(w, "failed to scan deck", http.StatusInternalServerError)
			return
		}

		if len(cardsJSON) > 0 {
			if err := json.Unmarshal(cardsJSON, &deck.Cards); err != nil {
				http.Error(w, "failed to decode cards", http.StatusInternalServerError)
				return
			}
		}

		decks = append(decks, deck)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "failed to read decks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if len(decks[0].Cards) == 0 {
		http.Error(w, "deck has no cards", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err := components.CardFragment(
		decks[0].Title,
		decks[0].Description,
		decks[0].Cards,
	).Render(r.Context(), w); err != nil {
		log.Printf("render card: %v", err)
	}

}
