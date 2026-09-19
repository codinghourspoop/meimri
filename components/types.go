package components

type CardInput struct {
	Front string `json:"front"`
	Back  string `json:"back"`
}

type NotificationItem struct {
	Body   string
	IsRead int64
}

type DeckSearchResult struct {
	ID          int64
	Title       string
	Description string
	CardCount   int
}
