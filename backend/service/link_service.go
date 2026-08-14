package service

import (
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/matsu0122-png/linkvault/backend/model"
)

var ErrNotFound = errors.New("link not found")

type linkRepository interface {
	Create(link model.Link) (model.Link, error)
	List(query, tag string) ([]model.Link, error)
	Update(link model.Link) (model.Link, error)
	Delete(id int) error
}

// metadataFetcher fetches page metadata (title, description, OGP image,
// favicon) for a URL. Failures are non-fatal: callers fall back to whatever
// fields the caller already has rather than rejecting the request.
type metadataFetcher interface {
	FetchMetadata(url string) (model.Metadata, error)
}

type LinkService struct {
	repo    linkRepository
	fetcher metadataFetcher
}

func NewLinkService(repo linkRepository, fetcher metadataFetcher) *LinkService {
	return &LinkService{repo: repo, fetcher: fetcher}
}

func (s *LinkService) CreateLink(url, title, memo string, tags []string) (model.Link, error) {
	if url == "" {
		return model.Link{}, errors.New("url is required")
	}

	link := model.Link{
		URL:  url,
		Memo: memo,
		Tags: normalizeTags(tags),
	}

	meta, err := s.fetcher.FetchMetadata(url)
	if err != nil {
		log.Printf("fetch metadata for %s: %v", url, err)
	} else {
		link.Description = meta.Description
		link.ImageURL = meta.ImageURL
		link.FaviconURL = meta.FaviconURL
	}

	if title != "" {
		link.Title = title
	} else {
		link.Title = meta.Title
	}

	now := time.Now()
	link.CreatedAt = now
	link.UpdatedAt = now

	return s.repo.Create(link)
}

// normalizeTags trims whitespace and drops empty or duplicate tag names.
func normalizeTags(tags []string) []string {
	seen := make(map[string]bool, len(tags))
	normalized := make([]string, 0, len(tags))

	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" || seen[trimmed] {
			continue
		}

		seen[trimmed] = true
		normalized = append(normalized, trimmed)
	}

	return normalized
}

func (s *LinkService) ListLinks(query, tag string) ([]model.Link, error) {
	return s.repo.List(query, tag)
}

func (s *LinkService) UpdateLink(id int, url, title, memo string, tags []string) (model.Link, error) {
	if url == "" {
		return model.Link{}, errors.New("url is required")
	}

	link := model.Link{
		ID:        id,
		URL:       url,
		Title:     title,
		Memo:      memo,
		Tags:      normalizeTags(tags),
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
