package repository

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/lib/pq"
	"github.com/matsu0122-png/linkvault/backend/model"
)

var (
	// ErrDuplicateName is returned by Create/Update when collections.name's
	// UNIQUE constraint rejects the write.
	ErrDuplicateName = errors.New("collection name already exists")

	// ErrCollectionNotFound and ErrLinkNotFound are returned by AddLink when
	// collection_links' foreign keys reject the write because the given
	// collection_id or link_id doesn't exist.
	ErrCollectionNotFound = errors.New("collection not found")
	ErrLinkNotFound       = errors.New("link not found")

	// ErrParentNotFound is returned by Create when the given parent_id
	// doesn't reference an existing collection.
	ErrParentNotFound = errors.New("parent collection not found")
)

type CollectionRepository struct {
	db *sql.DB
}

func NewCollectionRepository(db *sql.DB) *CollectionRepository {
	return &CollectionRepository{db: db}
}

// collectionSelectColumns is shared by every query that returns full
// Collection rows, so link_count is always computed the same way.
const collectionSelectColumns = `
	c.id, c.name, c.description, c.parent_id, c.created_at, c.updated_at, COUNT(cl.link_id)
`

func (r *CollectionRepository) Create(c model.Collection) (model.Collection, error) {
	query := `
		INSERT INTO collections (name, description, parent_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	if err := r.db.QueryRow(query, c.Name, c.Description, c.ParentID, c.CreatedAt, c.UpdatedAt).Scan(&c.ID); err != nil {
		if isUniqueViolation(err) {
			return model.Collection{}, ErrDuplicateName
		}
		if isForeignKeyViolation(err) {
			return model.Collection{}, ErrParentNotFound
		}
		return model.Collection{}, err
	}

	return c, nil
}

// List returns all collections ordered by name, each with its link_count.
// Passing 0 for linkID returns every collection; a positive linkID filters
// to only the collections that link belongs to.
func (r *CollectionRepository) List(linkID int) ([]model.Collection, error) {
	query := `
		SELECT ` + collectionSelectColumns + `
		FROM collections c
		LEFT JOIN collection_links cl ON cl.collection_id = c.id
		WHERE ($1 = 0 OR c.id IN (SELECT collection_id FROM collection_links WHERE link_id = $1))
		GROUP BY c.id
		ORDER BY c.name
	`

	rows, err := r.db.Query(query, linkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCollections(rows)
}

func (r *CollectionRepository) Get(id int) (model.Collection, error) {
	query := `
		SELECT ` + collectionSelectColumns + `
		FROM collections c
		LEFT JOIN collection_links cl ON cl.collection_id = c.id
		WHERE c.id = $1
		GROUP BY c.id
	`

	c, err := scanCollection(r.db.QueryRow(query, id))
	if err != nil {
		return model.Collection{}, err
	}

	return c, nil
}

func (r *CollectionRepository) Update(c model.Collection) (model.Collection, error) {
	query := `UPDATE collections SET name = $1, description = $2, updated_at = $3 WHERE id = $4`

	result, err := r.db.Exec(query, c.Name, c.Description, c.UpdatedAt, c.ID)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Collection{}, ErrDuplicateName
		}
		return model.Collection{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return model.Collection{}, err
	}
	if rowsAffected == 0 {
		return model.Collection{}, sql.ErrNoRows
	}

	return r.Get(c.ID)
}

func (r *CollectionRepository) Delete(id int) error {
	query := `DELETE FROM collections WHERE id = $1`

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

// AddLink associates linkID with collectionID. Associating an already-linked
// pair is a no-op (ON CONFLICT DO NOTHING), not an error: adding a link to a
// collection it's already in has nothing left to do.
func (r *CollectionRepository) AddLink(collectionID, linkID int) error {
	query := `
		INSERT INTO collection_links (collection_id, link_id)
		VALUES ($1, $2)
		ON CONFLICT (collection_id, link_id) DO NOTHING
	`

	_, err := r.db.Exec(query, collectionID, linkID)
	if err == nil {
		return nil
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23503" {
		if strings.Contains(pqErr.Constraint, "link_id") {
			return ErrLinkNotFound
		}
		return ErrCollectionNotFound
	}

	return err
}

func isForeignKeyViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23503"
}

func (r *CollectionRepository) RemoveLink(collectionID, linkID int) error {
	query := `DELETE FROM collection_links WHERE collection_id = $1 AND link_id = $2`

	result, err := r.db.Exec(query, collectionID, linkID)
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

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

// rowScanner is satisfied by both *sql.Row and *sql.Rows, so scanCollection
// can be reused for a single-row lookup (Get) and a multi-row loop (List).
type rowScanner interface {
	Scan(dest ...any) error
}

func scanCollection(row rowScanner) (model.Collection, error) {
	var c model.Collection
	var parentID sql.NullInt64

	if err := row.Scan(&c.ID, &c.Name, &c.Description, &parentID, &c.CreatedAt, &c.UpdatedAt, &c.LinkCount); err != nil {
		return model.Collection{}, err
	}
	if parentID.Valid {
		id := int(parentID.Int64)
		c.ParentID = &id
	}

	return c, nil
}

func scanCollections(rows *sql.Rows) ([]model.Collection, error) {
	collections := []model.Collection{}
	for rows.Next() {
		c, err := scanCollection(rows)
		if err != nil {
			return nil, err
		}
		collections = append(collections, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return collections, nil
}
