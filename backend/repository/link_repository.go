package repository

import (
	"database/sql"

	"github.com/matsu0122-png/linkvault/backend/model"
)

type LinkRepository struct {
	db *sql.DB
}

func NewLinkRepository(db *sql.DB) *LinkRepository {
	return &LinkRepository{db: db}
}

func (r *LinkRepository) Create(link model.Link) (model.Link, error) {
	query := `
		INSERT INTO links (url, title, memo, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := r.db.QueryRow(query, link.URL, link.Title, link.Memo, link.CreatedAt, link.UpdatedAt).Scan(&link.ID)
	if err != nil {
		return model.Link{}, err
	}

	return link, nil
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

func (r *LinkRepository) List() ([]model.Link, error) {
	query := `
		SELECT id, url, title, memo, created_at, updated_at
		FROM links
		ORDER BY id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := []model.Link{}
	for rows.Next() {
		var link model.Link

		err := rows.Scan(&link.ID, &link.URL, &link.Title, &link.Memo, &link.CreatedAt, &link.UpdatedAt)
		if err != nil {
			return nil, err
		}

		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}

func (r *LinkRepository) Update(link model.Link) (model.Link, error) {
	query := `
		UPDATE links
		SET url = $1, title = $2, memo = $3, updated_at = $4
		WHERE id = $5
		RETURNING created_at
	`

	err := r.db.QueryRow(query, link.URL, link.Title, link.Memo, link.UpdatedAt, link.ID).Scan(&link.CreatedAt)
	if err != nil {
		return model.Link{}, err
	}

	return link, nil
}
