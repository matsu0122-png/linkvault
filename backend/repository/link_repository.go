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

// List returns links, optionally filtered by a keyword (matched against url/title/memo)
// and/or an exact tag name. Passing "" for either skips that filter.
func (r *LinkRepository) List(query, tag string) ([]model.Link, error) {
	sqlQuery := `
		SELECT l.id, l.url, l.title, l.memo, l.description, l.image_url, l.favicon_url,
		       l.status, l.checked_at, l.created_at, l.updated_at,
		       COALESCE(array_agg(t.name ORDER BY t.name) FILTER (WHERE t.name IS NOT NULL), '{}')
		FROM links l
		LEFT JOIN link_tags lt ON lt.link_id = l.id
		LEFT JOIN tags t ON t.id = lt.tag_id
		WHERE ($1 = '' OR l.title ILIKE '%' || $1 || '%' OR l.url ILIKE '%' || $1 || '%' OR l.memo ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR l.id IN (
		        SELECT lt2.link_id
		        FROM link_tags lt2
		        JOIN tags t2 ON t2.id = lt2.tag_id
		        WHERE t2.name = $2
		      ))
		GROUP BY l.id
		ORDER BY l.id
	`

	rows, err := r.db.Query(sqlQuery, query, tag)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
