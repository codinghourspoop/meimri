package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"gitlab.com/coding4hours/memri/internal/auth"
)

func parsePasses(r *http.Request) int {
	n, err := strconv.Atoi(r.URL.Query().Get("passes"))
	if err != nil || n < 1 || n > 10 {
		return 3
	}
	return n
}

// GET /api/learn?deck_id=&passes= — full deck payload; the Svelte
// Learn page runs the session algorithm client-side.
func (a *App) ReadLearn(w http.ResponseWriter, r *http.Request) {
	if auth.UserID(r) == 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	deckID, err := strconv.ParseInt(r.URL.Query().Get("deck_id"), 10, 64)
	if err != nil || deckID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid deck id")
		return
	}

	var deck deckJSON
	var cardsJSON []byte
	err = a.DB.QueryRowContext(
		r.Context(),
		`SELECT id, title, description, cards FROM deck WHERE id = ?`,
		deckID,
	).Scan(&deck.ID, &deck.Title, &deck.Description, &cardsJSON)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "deck not found")
		return
	}
	if err != nil {
		log.Printf("ReadLearn database error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to read deck")
		return
	}
	if len(cardsJSON) > 0 {
		if err := json.Unmarshal(cardsJSON, &deck.Cards); err != nil {
			log.Printf("ReadLearn decode cards error: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to decode cards")
			return
		}
	}
	if deck.Cards == nil {
		deck.Cards = []CardInput{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":          deck.ID,
		"title":       deck.Title,
		"description": deck.Description,
		"cards":       deck.Cards,
		"passes":      parsePasses(r),
	})
}
