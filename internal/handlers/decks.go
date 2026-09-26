package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"gitlab.com/coding4hours/memri/internal/auth"
)

type createDeckRequest struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Cards       []CardInput `json:"cards"`
}

type deckJSON struct {
	ID          int64       `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Cards       []CardInput `json:"cards"`
}

type deckSearchJSON struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CardCount   int    `json:"card_count"`
}

// GET /api/decks/search?q= — empty query returns no results;
// search only runs once the user types.
func (a *App) SearchDecks(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserID(r)
	if userID == 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeJSON(w, http.StatusOK, []deckSearchJSON{})
		return
	}

	like := "%" + escapeLike(q) + "%"
	rows, err := a.DB.QueryContext(
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
	if err != nil {
		log.Printf("SearchDecks database error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to search decks")
		return
	}
	defer rows.Close()

	results := make([]deckSearchJSON, 0)
	for rows.Next() {
		var id int64
		var title, description string
		var cardsJSON []byte
		if err := rows.Scan(&id, &title, &description, &cardsJSON); err != nil {
			log.Printf("SearchDecks scan error: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to scan deck")
			return
		}
		var cards []CardInput
		if len(cardsJSON) > 0 {
			if err := json.Unmarshal(cardsJSON, &cards); err != nil {
				log.Printf("SearchDecks decode cards error: %v", err)
				writeError(w, http.StatusInternalServerError, "failed to decode cards")
				return
			}
		}
		results = append(results, deckSearchJSON{
			ID:          id,
			Title:       title,
			Description: description,
			CardCount:   len(cards),
		})
	}
	if err := rows.Err(); err != nil {
		log.Printf("SearchDecks rows error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to read decks")
		return
	}
	writeJSON(w, http.StatusOK, results)
}

// escapeLike escapes \, % and _ so user input is treated literally in LIKE.
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// POST /api/decks — JSON body {title, description, cards}.
func (a *App) CreateDeck(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserID(r)
	if userID == 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input createDeckRequest
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	clean := make([]CardInput, 0, len(input.Cards))
	for _, c := range input.Cards {
		c.Front = strings.TrimSpace(c.Front)
		c.Back = strings.TrimSpace(c.Back)
		if c.Front == "" && c.Back == "" {
			continue
		}
		clean = append(clean, c)
	}
	input.Cards = clean

	if input.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	cardsJSON, err := json.Marshal(input.Cards)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to encode cards")
		return
	}

	result, err := a.DB.ExecContext(
		r.Context(),
		`INSERT INTO deck (user_id, title, description, cards)
		 VALUES (?, ?, ?, ?)`,
		userID,
		input.Title,
		input.Description,
		string(cardsJSON),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create deck")
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get deck ID")
		return
	}

	writeJSON(w, http.StatusCreated, deckJSON{
		ID:          id,
		Title:       input.Title,
		Description: input.Description,
		Cards:       input.Cards,
	})
}

// GET /api/decks/{id}
func (a *App) ReadDeck(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserID(r)
	if userID == 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	deckID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || deckID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid deck id")
		return
	}

	var deck deckJSON
	var cardsJSON []byte
	err = a.DB.QueryRowContext(
		r.Context(),
		`SELECT id, title, description, cards FROM deck WHERE id = ? AND user_id = ?`,
		deckID, userID,
	).Scan(&deck.ID, &deck.Title, &deck.Description, &cardsJSON)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "deck not found")
		return
	}
	if err != nil {
		log.Printf("ReadDeck database error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to read deck")
		return
	}
	if len(cardsJSON) > 0 {
		if err := json.Unmarshal(cardsJSON, &deck.Cards); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to decode cards")
			return
		}
	}
	if deck.Cards == nil {
		deck.Cards = []CardInput{}
	}
	writeJSON(w, http.StatusOK, deck)
}
