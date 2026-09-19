package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"gitlab.com/coding4hours/memri/components"
)

func parsePasses(r *http.Request) int {
	n, err := strconv.Atoi(r.FormValue("passes"))
	if err != nil || n < 1 || n > 10 {
		return 3
	}
	return n
}

func (a *App) LearnView(w http.ResponseWriter, r *http.Request) {
	uid, _, _, unread := a.currentUser(r)
	deckID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || deckID <= 0 {
		http.NotFound(w, r)
		return
	}
	_ = components.LearnPage(uid != 0, unread, deckID, parsePasses(r)).Render(r.Context(), w)
}

func (a *App) ReadLearn(w http.ResponseWriter, r *http.Request) {
	idStr := r.FormValue("deck_id")
	if idStr == "" {
		idStr = chi.URLParam(r, "id")
	}
	deckID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || deckID <= 0 {
		http.Error(w, "invalid deck id", http.StatusBadRequest)
		return
	}

	var deck deckResponse
	var cardsJSON []byte
	err = a.DB.QueryRowContext(
		r.Context(),
		`SELECT id, title, description, cards FROM deck WHERE id = ?`,
		deckID,
	).Scan(&deck.ID, &deck.Title, &deck.Description, &cardsJSON)
	if err != nil {
		http.Error(w, "deck not found", http.StatusNotFound)
		return
	}

	if len(cardsJSON) > 0 {
		if err := json.Unmarshal(cardsJSON, &deck.Cards); err != nil {
			log.Printf("ReadLearn decode cards error: %v", err)
			http.Error(w, "failed to decode cards", http.StatusInternalServerError)
			return
		}
	}
	if deck.Cards == nil {
		deck.Cards = []components.CardInput{}
	}

	encoded, err := json.Marshal(deck.Cards)
	if err != nil {
		http.Error(w, "failed to encode cards", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := components.LearnFragment(
		deck.Title,
		deck.Description,
		deck.ID,
		parsePasses(r),
		string(encoded),
	).Render(r.Context(), w); err != nil {
		log.Printf("render learn: %v", err)
	}
}
