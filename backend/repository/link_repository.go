package repository

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/matsu0122-png/linkvault/backend/model"
)

type LinkRepository struct {
	db *sql.DB
}

func NewLinkRepository(db *sql.DB) *LinkRepository {
	return &LinkRepository{db: db}
}

func (r *LinkRepository) Create(link model.Link) (model.Link, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return model.Link{}, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO links (url, title, memo, description, image_url, favicon_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	if err := tx.QueryRow(
		query,
		link.URL, link.Title, link.Memo, link.Description, link.ImageURL, link.FaviconURL,
		link.CreatedAt, link.UpdatedAt,
	).Scan(&link.ID); err != nil {
		return model.Link{}, err
	}
	link.Status = model.StatusUnknown

	tagIDs, err := upsertTags(tx, link.Tags)
	if err != nil {
		return model.Link{}, err
	}

	if err := linkTags(tx, link.ID, tagIDs); err != nil {
		return model.Link{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.Link{}, err
	}

	return link, nil
}

// upsertTags ensures each tag name exists in the tags table and returns their ids.
// ON CONFLICT ... DO UPDATE (instead of DO NOTHING) is used so RETURNING id always
// yields a row, even when the tag already existed.
func upsertTags(tx *sql.Tx, names []string) ([]int64, error) {
	if len(names) == 0 {
		return nil, nil
	}

	query := `
		INSERT INTO tags (name)
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`

	ids := make([]int64, 0, len(names))
	for _, name := range names {
		var id int64
		if err := tx.QueryRow(query, name).Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func linkTags(tx *sql.Tx, linkID int, tagIDs []int64) error {
	query := `INSERT INTO link_tags (link_id, tag_id) VALUES ($1, $2)`

	for _, tagID := range tagIDs {
		if _, err := tx.Exec(query, linkID, tagID); err != nil {
			return err
		}
	}

	return nil
}

func (r *LinkRepository) UpdateStatus(id int, status string, checkedAt time.Time) error {
	query := `UPDATE links SET status = $1, checked_at = $2 WHERE id = $3`

	result, err := r.db.Exec(query, status, checkedAt, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *LinkRepository) Delete(id int) error {
	query := `DELETE FROM links WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// linkFilterWhere is the WHERE clause shared by every query that lists or
// counts links, so a filtered list and its total count can never drift
// apart. It only references l (the links table) and takes the same three
// params everywhere it's used: $1 = keyword, $2 = tag name, $3 = collection
// id (0 skips the filter, same convention as "" for $1/$2).
const linkFilterWhere = `
	WHERE ($1 = '' OR l.title ILIKE '%' || $1 || '%' OR l.url ILIKE '%' || $1 || '%' OR l.memo ILIKE '%' || $1 || '%')
	  AND ($2 = '' OR l.id IN (
	        SELECT lt2.link_id
	        FROM link_tags lt2
	        JOIN tags t2 ON t2.id = lt2.tag_id
	        WHERE t2.name = $2
	      ))
	  AND ($3 = 0 OR l.id IN (
	        SELECT cl2.link_id
	        FROM collection_links cl2
	        WHERE cl2.collection_id = $3
	      ))
`

// linkSelectColumns is the column list shared by every query that returns
// full link rows (as opposed to just a count).
const linkSelectColumns = `
	l.id, l.url, l.title, l.memo, l.description, l.image_url, l.favicon_url,
	l.status, l.checked_at, l.created_at, l.updated_at,
	COALESCE(array_agg(t.name ORDER BY t.name) FILTER (WHERE t.name IS NOT NULL), '{}')
`

// orderByClause maps a validated SortOption to a fixed ORDER BY fragment.
// It never interpolates the caller's raw string directly into SQL, so an
// unrecognized SortOption safely falls back to the default order instead of
// risking SQL injection through an ORDER BY built from user input.
func orderByClause(sort model.SortOption) string {
	switch sort {
	case model.SortCreatedAsc:
		return "l.created_at ASC, l.id ASC"
	case model.SortUpdatedDesc:
		return "l.updated_at DESC, l.id DESC"
	case model.SortUpdatedAsc:
		return "l.updated_at ASC, l.id ASC"
	case model.SortTitleAsc:
		return "l.title ASC, l.id ASC"
	case model.SortTitleDesc:
		return "l.title DESC, l.id DESC"
	default: // model.SortCreatedDesc, and any unrecognized value
		return "l.created_at DESC, l.id DESC"
	}
}

// List returns all links, optionally filtered by a keyword (matched against
// url/title/memo) and/or an exact tag name. Passing "" for either skips
// that filter. Unlike ListPage, it has no concept of paging: it's used by
// CheckLinks, which needs to sweep every saved link regardless of any
// page/sort the user currently has selected.
func (r *LinkRepository) List(query, tag string) ([]model.Link, error) {
	sqlQuery := `
		SELECT ` + linkSelectColumns + `
		FROM links l
		LEFT JOIN link_tags lt ON lt.link_id = l.id
		LEFT JOIN tags t ON t.id = lt.tag_id
		` + linkFilterWhere + `
		GROUP BY l.id
		ORDER BY ` + orderByClause(model.SortCreatedDesc)

	// List has no concept of a collection filter (0 = no filter); it's
	// only used by CheckLinks, which sweeps every saved link.
	rows, err := r.db.Query(sqlQuery, query, tag, 0)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanLinks(rows)
}

// ListPage returns one page of links matching query/tag/collection, ordered
// by sort, along with the total number of links that match (across all
// pages) so the caller can compute how many pages exist. collectionID = 0
// skips the collection filter.
func (r *LinkRepository) ListPage(query, tag string, collectionID int, sort model.SortOption, limit, offset int) ([]model.Link, int, error) {
	countQuery := `SELECT COUNT(*) FROM links l ` + linkFilterWhere

	var total int
	if err := r.db.QueryRow(countQuery, query, tag, collectionID).Scan(&total); err != nil {
		return nil, 0, err
	}

	sqlQuery := `
		SELECT ` + linkSelectColumns + `
		FROM links l
		LEFT JOIN link_tags lt ON lt.link_id = l.id
		LEFT JOIN tags t ON t.id = lt.tag_id
		` + linkFilterWhere + `
		GROUP BY l.id
		ORDER BY ` + orderByClause(sort) + `
		LIMIT $4 OFFSET $5
	`

	rows, err := r.db.Query(sqlQuery, query, tag, collectionID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	links, err := scanLinks(rows)
	if err != nil {
		return nil, 0, err
	}

	return links, total, nil
}

func scanLinks(rows *sql.Rows) ([]model.Link, error) {
	links := []model.Link{}
	for rows.Next() {
		var link model.Link
		var checkedAt sql.NullTime

		err := rows.Scan(
			&link.ID, &link.URL, &link.Title, &link.Memo,
			&link.Description, &link.ImageURL, &link.FaviconURL,
			&link.Status, &checkedAt,
			&link.CreatedAt, &link.UpdatedAt, pq.Array(&link.Tags),
		)
		if err != nil {
			return nil, err
		}
		if checkedAt.Valid {
			link.CheckedAt = &checkedAt.Time
		}

		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}

func (r *LinkRepository) Update(link model.Link) (model.Link, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return model.Link{}, err
	}
	defer tx.Rollback()

	// description/image_url/favicon_url/status/checked_at are intentionally left
	// untouched here: there's no way to edit them from the UI, and Update never
	// re-fetches or re-checks, so overwriting them would wipe out whatever
	// Create/CheckLinks originally captured.
	query := `
		UPDATE links
		SET url = $1, title = $2, memo = $3, updated_at = $4
		WHERE id = $5
		RETURNING created_at, description, image_url, favicon_url, status, checked_at
	`

	var checkedAt sql.NullTime
	if err := tx.QueryRow(query, link.URL, link.Title, link.Memo, link.UpdatedAt, link.ID).Scan(
		&link.CreatedAt, &link.Description, &link.ImageURL, &link.FaviconURL, &link.Status, &checkedAt,
	); err != nil {
		return model.Link{}, err
	}
	if checkedAt.Valid {
		link.CheckedAt = &checkedAt.Time
	}

	if _, err := tx.Exec(`DELETE FROM link_tags WHERE link_id = $1`, link.ID); err != nil {
		return model.Link{}, err
	}

	tagIDs, err := upsertTags(tx, link.Tags)
	if err != nil {
		return model.Link{}, err
	}

	if err := linkTags(tx, link.ID, tagIDs); err != nil {
		return model.Link{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.Link{}, err
	}

	return link, nil
}
