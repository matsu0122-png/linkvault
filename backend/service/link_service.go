package service

import (
	"database/sql"
	"errors"
	"time"

	"github.com/matsu0122-png/linkvault/backend/model"
	"github.com/matsu0122-png/linkvault/backend/repository"
)

var ErrNotFound = errors.New("link not found")

type LinkService struct {
	repo *repository.LinkRepository
}

func NewLinkService(repo *repository.LinkRepository) *LinkService {
	return &LinkService{repo: repo}
}

func (s *LinkService) CreateLink(url, title, memo string) (model.Link, error) {
	if url == "" {
		return model.Link{}, errors.New("url is required")
	}

	now := time.Now()
	link := model.Link{
		URL:       url,
		Title:     title,
		Memo:      memo,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return s.repo.Create(link)
}

func (s *LinkService) ListLinks() ([]model.Link, error) {
	return s.repo.List()
}

func (s *LinkService) UpdateLink(id int, url, title, memo string) (model.Link, error) {
	if url == "" {
		return model.Link{}, errors.New("url is required")
	}

	link := model.Link{
		ID:        id,
		URL:       url,
		Title:     title,
		Memo:      memo,
		UpdatedAt: time.Now(),
	}

	updated, err := s.repo.Update(link)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Link{}, ErrNotFound
	}
	if err != nil {
		return model.Link{}, err
	}

	return updated, nil
}

func (s *LinkService) DeleteLink(id int) error {
	err := s.repo.Delete(id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	return err
}
