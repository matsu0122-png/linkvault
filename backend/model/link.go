package model

import "time"

type Link struct {
	ID          int        `json:"id"`
	URL         string     `json:"url"`
	Title       string     `json:"title"`
	Memo        string     `json:"memo"`
	Tags        []string   `json:"tags"`
	Description string     `json:"description"`
	ImageURL    string     `json:"image_url"`
	FaviconURL  string     `json:"favicon_url"`
	Status      string     `json:"status"`
	CheckedAt   *time.Time `json:"checked_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

const (
	StatusUnknown = "unknown"
	StatusOK      = "ok"
	StatusBroken  = "broken"
)

// Metadata holds page information auto-extracted from a fetched URL.
type Metadata struct {
	Title       string
	Description string
	ImageURL    string
	FaviconURL  string
}
