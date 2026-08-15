package model

import "time"

type Collection struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ParentID    *int      `json:"parent_id"`
	LinkCount   int       `json:"link_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
