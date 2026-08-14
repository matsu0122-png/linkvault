package service

import (
	"database/sql"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/matsu0122-png/linkvault/backend/model"
)

var ErrNotFound = errors.New("link not found")

const (
	// MaxBulkURLs is the largest number of URLs accepted by a single
	// BulkCreateLinks call.
	MaxBulkURLs = 50

	// maxConcurrentFetches bounds how many URLs are fetched at once during a
	// bulk create, to avoid hammering either this server or the target sites.
	maxConcurrentFetches = 5
)

type linkRepository interface {
	Create(link model.Link) (model.Link, error)
	List(query, tag string) ([]model.Link, error)
	Update(link model.Link) (model.Link, error)
	Delete(id int) error
	UpdateStatus(id int, status string, checkedAt time.Time) error
}

// metadataFetcher fetches page metadata (title, description, OGP image,
// favicon) for a URL, and can check whether a URL is currently reachable.
// Failures are non-fatal: callers fall back to whatever fields the caller
// already has rather than rejecting the request.
type metadataFetcher interface {
	FetchMetadata(url string) (model.Metadata, error)
	CheckAlive(url string) (bool, error)
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

// BulkCreateResult is the outcome of a BulkCreateLinks call. Created and
// Failed preserve the order of the input urls slice.
type BulkCreateResult struct {
	Created []model.Link
	Failed  []BulkCreateFailure
}

type BulkCreateFailure struct {
	URL   string
	Error string
}

// BulkCreateLinks creates a link for each url, fetching metadata for all of
// them concurrently (bounded by maxConcurrentFetches). tags are applied to
// every created link. One url failing (e.g. it's empty) does not affect the
// others; each is created independently via CreateLink, so metadata-fetch
// failures are handled the same non-fatal way as a single create.
func (s *LinkService) BulkCreateLinks(urls []string, tags []string) BulkCreateResult {
	type outcome struct {
		link model.Link
		err  error
	}

	outcomes := make([]outcome, len(urls))
	sem := make(chan struct{}, maxConcurrentFetches)
	var wg sync.WaitGroup

	for i, url := range urls {
		wg.Add(1)
		go func(i int, url string) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			link, err := s.CreateLink(url, "", "", tags)
			outcomes[i] = outcome{link: link, err: err}
		}(i, url)
	}
	wg.Wait()

	result := BulkCreateResult{
		Created: make([]model.Link, 0, len(urls)),
		Failed:  make([]BulkCreateFailure, 0),
	}
	for i, o := range outcomes {
		if o.err != nil {
			result.Failed = append(result.Failed, BulkCreateFailure{URL: urls[i], Error: o.err.Error()})
			continue
		}
		result.Created = append(result.Created, o.link)
	}

	return result
}

// CheckLinks checks every saved link concurrently (bounded by
// maxConcurrentFetches) and updates its status to StatusOK or StatusBroken.
// It returns the refreshed list of all links. A single link's check failing
// doesn't stop the others.
func (s *LinkService) CheckLinks() ([]model.Link, error) {
	links, err := s.repo.List("", "")
	if err != nil {
		return nil, err
	}

	sem := make(chan struct{}, maxConcurrentFetches)
	var wg sync.WaitGroup

	for _, link := range links {
		wg.Add(1)
		go func(link model.Link) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			alive, err := s.fetcher.CheckAlive(link.URL)
			if err != nil {
				log.Printf("check alive for %s: %v", link.URL, err)
			}

			status := model.StatusBroken
			if alive {
				status = model.StatusOK
			}

			if err := s.repo.UpdateStatus(link.ID, status, time.Now()); err != nil {
				log.Printf("update status for link %d: %v", link.ID, err)
			}
		}(link)
	}
	wg.Wait()

	return s.repo.List("", "")
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
