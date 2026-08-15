package model

// Pagination describes where a page of results sits within the full,
// filtered result set.
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// SortOption selects how ListPage orders results. The zero value is not a
// valid option; callers should go through a validating parse step (see
// service.parseSort) rather than constructing one directly from user input.
type SortOption string

const (
	SortCreatedDesc SortOption = "created_at_desc"
	SortCreatedAsc  SortOption = "created_at_asc"
	SortUpdatedDesc SortOption = "updated_at_desc"
	SortUpdatedAsc  SortOption = "updated_at_asc"
	SortTitleAsc    SortOption = "title_asc"
	SortTitleDesc   SortOption = "title_desc"
)
