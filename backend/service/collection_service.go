package service

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/matsu0122-png/linkvault/backend/model"
	"github.com/matsu0122-png/linkvault/backend/repository"
)

var (
	ErrCollectionNotFound = errors.New("collection not found")
	ErrLinkNotFound       = errors.New("link not found")
	ErrDuplicateName      = errors.New("collection name already exists")
	ErrParentNotFound     = errors.New("parent collection not found")
)

const maxCollectionNameLength = 100

type collectionRepository interface {
	Create(c model.Collection) (model.Collection, error)
	List(linkID int) ([]model.Collection, error)
	Get(id int) (model.Collection, error)
	Update(c model.Collection) (model.Collection, error)
	Delete(id int) error
	AddLink(collectionID, linkID int) error
	RemoveLink(collectionID, linkID int) error
}

type CollectionService struct {
	repo collectionRepository
}

func NewCollectionService(repo collectionRepository) *CollectionService {
	return &CollectionService{repo: repo}
}

// CreateCollection creates a collection, optionally nested under parentID
// (nil for a top-level collection). Where a collection sits in the tree can
// only be set here, at creation time: there's no "move" operation that
// reparents an existing collection, so a collection can never reference one
// of its own descendants as a parent — a newly created collection has no
// descendants yet, and its parent chain never changes afterward, so cycles
// simply can't arise.
func (s *CollectionService) CreateCollection(name, description string, parentID *int) (model.Collection, error) {
	name = strings.TrimSpace(name)
	if err := validateCollectionName(name); err != nil {
		return model.Collection{}, err
	}

	now := time.Now()
	c, err := s.repo.Create(model.Collection{
		Name:        name,
		Description: description,
		ParentID:    parentID,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if errors.Is(err, repository.ErrDuplicateName) {
		return model.Collection{}, ErrDuplicateName
	}
	if errors.Is(err, repository.ErrParentNotFound) {
		return model.Collection{}, ErrParentNotFound
	}

	return c, err
}

// ListCollections returns every collection. If linkID is positive, it
// returns only the collections that link belongs to (used by the Link edit
// form to pre-check the link's current collections).
func (s *CollectionService) ListCollections(linkID int) ([]model.Collection, error) {
	return s.repo.List(linkID)
}

func (s *CollectionService) GetCollection(id int) (model.Collection, error) {
	c, err := s.repo.Get(id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Collection{}, ErrCollectionNotFound
	}

	return c, err
}

func (s *CollectionService) UpdateCollection(id int, name, description string) (model.Collection, error) {
	name = strings.TrimSpace(name)
	if err := validateCollectionName(name); err != nil {
		return model.Collection{}, err
	}

	c, err := s.repo.Update(model.Collection{
		ID:          id,
		Name:        name,
		Description: description,
		UpdatedAt:   time.Now(),
	})
	if errors.Is(err, repository.ErrDuplicateName) {
		return model.Collection{}, ErrDuplicateName
	}
	if errors.Is(err, sql.ErrNoRows) {
		return model.Collection{}, ErrCollectionNotFound
	}

	return c, err
}

func (s *CollectionService) DeleteCollection(id int) error {
	err := s.repo.Delete(id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrCollectionNotFound
	}

	return err
}

func (s *CollectionService) AddLink(collectionID, linkID int) error {
	err := s.repo.AddLink(collectionID, linkID)
	if errors.Is(err, repository.ErrCollectionNotFound) {
		return ErrCollectionNotFound
	}
	if errors.Is(err, repository.ErrLinkNotFound) {
		return ErrLinkNotFound
	}

	return err
}

func (s *CollectionService) RemoveLink(collectionID, linkID int) error {
	err := s.repo.RemoveLink(collectionID, linkID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrCollectionNotFound
	}

	return err
}

func validateCollectionName(name string) error {
	if name == "" {
		return newValidationError("name is required")
	}
	if len(name) > maxCollectionNameLength {
		return newValidationError("name is too long")
	}

	return nil
}
