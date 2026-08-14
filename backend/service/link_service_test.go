package service

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/matsu0122-png/linkvault/backend/model"
)

type mockLinkRepository struct {
	createFn       func(link model.Link) (model.Link, error)
	listFn         func(query, tag string) ([]model.Link, error)
	updateFn       func(link model.Link) (model.Link, error)
	deleteFn       func(id int) error
	updateStatusFn func(id int, status string, checkedAt time.Time) error
}

func (m *mockLinkRepository) Create(link model.Link) (model.Link, error) {
	return m.createFn(link)
}

func (m *mockLinkRepository) List(query, tag string) ([]model.Link, error) {
	return m.listFn(query, tag)
}

func (m *mockLinkRepository) Update(link model.Link) (model.Link, error) {
	return m.updateFn(link)
}

func (m *mockLinkRepository) Delete(id int) error {
	return m.deleteFn(id)
}

func (m *mockLinkRepository) UpdateStatus(id int, status string, checkedAt time.Time) error {
	return m.updateStatusFn(id, status, checkedAt)
}

type mockMetadataFetcher struct {
	fetchFn      func(url string) (model.Metadata, error)
	checkAliveFn func(url string) (bool, error)
}

func (m *mockMetadataFetcher) FetchMetadata(url string) (model.Metadata, error) {
	return m.fetchFn(url)
}

func (m *mockMetadataFetcher) CheckAlive(url string) (bool, error) {
	return m.checkAliveFn(url)
}

// unusedFetcher fails the test if FetchMetadata or CheckAlive is called, for
// methods that never fetch (List/Update/Delete) or that return before
// reaching the fetch.
func unusedFetcher(t *testing.T) *mockMetadataFetcher {
	t.Helper()
	return &mockMetadataFetcher{
		fetchFn: func(url string) (model.Metadata, error) {
			t.Fatal("fetcher should not be called")
			return model.Metadata{}, nil
		},
		checkAliveFn: func(url string) (bool, error) {
			t.Fatal("checker should not be called")
			return false, nil
		},
	}
}

// noopFetcher succeeds with no metadata, for CreateLink tests that don't
// care about the fetch result.
func noopFetcher() *mockMetadataFetcher {
	return &mockMetadataFetcher{
		fetchFn: func(url string) (model.Metadata, error) {
			return model.Metadata{}, nil
		},
	}
}

func TestCreateLink(t *testing.T) {
	t.Run("成功したらリポジトリに保存されたリンクを返す", func(t *testing.T) {
		repo := &mockLinkRepository{
			createFn: func(link model.Link) (model.Link, error) {
				link.ID = 1
				return link, nil
			},
		}
		s := NewLinkService(repo, noopFetcher())

		got, err := s.CreateLink("https://example.com", "Example", "memo", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != 1 || got.URL != "https://example.com" || got.Title != "Example" || got.Memo != "memo" {
			t.Errorf("unexpected link: %+v", got)
		}
	})

	t.Run("urlが空ならバリデーションエラーを返す", func(t *testing.T) {
		repo := &mockLinkRepository{
			createFn: func(link model.Link) (model.Link, error) {
				t.Fatal("repository should not be called")
				return model.Link{}, nil
			},
		}
		s := NewLinkService(repo, unusedFetcher(t))

		_, err := s.CreateLink("", "Example", "memo", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("タグの前後空白除去と重複・空文字の除外をしてリポジトリに渡す", func(t *testing.T) {
		var gotTags []string
		repo := &mockLinkRepository{
			createFn: func(link model.Link) (model.Link, error) {
				gotTags = link.Tags
				link.ID = 1
				return link, nil
			},
		}
		s := NewLinkService(repo, noopFetcher())

		_, err := s.CreateLink("https://example.com", "Example", "memo", []string{" Go ", "web", "Go", "  "})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []string{"Go", "web"}
		if len(gotTags) != len(want) || gotTags[0] != want[0] || gotTags[1] != want[1] {
			t.Errorf("unexpected tags: %+v", gotTags)
		}
	})

	t.Run("titleが空ならfetcherで取得したtitleを使う", func(t *testing.T) {
		repo := &mockLinkRepository{
			createFn: func(link model.Link) (model.Link, error) {
				link.ID = 1
				return link, nil
			},
		}
		fetcher := &mockMetadataFetcher{
			fetchFn: func(url string) (model.Metadata, error) {
				return model.Metadata{Title: "Fetched Title"}, nil
			},
		}
		s := NewLinkService(repo, fetcher)

		got, err := s.CreateLink("https://example.com", "", "memo", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Title != "Fetched Title" {
			t.Errorf("expected fetched title, got %q", got.Title)
		}
	})

	t.Run("titleが空でfetcherが失敗しても空のまま保存される", func(t *testing.T) {
		repo := &mockLinkRepository{
			createFn: func(link model.Link) (model.Link, error) {
				link.ID = 1
				return link, nil
			},
		}
		fetcher := &mockMetadataFetcher{
			fetchFn: func(url string) (model.Metadata, error) {
				return model.Metadata{}, errors.New("timeout")
			},
		}
		s := NewLinkService(repo, fetcher)

		got, err := s.CreateLink("https://example.com", "", "memo", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Title != "" {
			t.Errorf("expected empty title, got %q", got.Title)
		}
	})

	t.Run("titleが指定されていればfetcherのtitleより優先される", func(t *testing.T) {
		repo := &mockLinkRepository{
			createFn: func(link model.Link) (model.Link, error) {
				link.ID = 1
				return link, nil
			},
		}
		fetcher := &mockMetadataFetcher{
			fetchFn: func(url string) (model.Metadata, error) {
				return model.Metadata{Title: "Fetched Title"}, nil
			},
		}
		s := NewLinkService(repo, fetcher)

		got, err := s.CreateLink("https://example.com", "My Title", "memo", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Title != "My Title" {
			t.Errorf("expected user-supplied title to win, got %q", got.Title)
		}
	})

	t.Run("titleを指定していてもdescription/image/faviconはfetch結果を使う", func(t *testing.T) {
		repo := &mockLinkRepository{
			createFn: func(link model.Link) (model.Link, error) {
				link.ID = 1
				return link, nil
			},
		}
		fetcher := &mockMetadataFetcher{
			fetchFn: func(url string) (model.Metadata, error) {
				return model.Metadata{
					Title:       "ignored",
					Description: "desc",
					ImageURL:    "https://example.com/og.png",
					FaviconURL:  "https://example.com/favicon.png",
				}, nil
			},
		}
		s := NewLinkService(repo, fetcher)

		got, err := s.CreateLink("https://example.com", "My Title", "memo", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Title != "My Title" {
			t.Errorf("Title = %q", got.Title)
		}
		if got.Description != "desc" || got.ImageURL != "https://example.com/og.png" || got.FaviconURL != "https://example.com/favicon.png" {
			t.Errorf("unexpected metadata: %+v", got)
		}
	})
}

func TestBulkCreateLinks(t *testing.T) {
	t.Run("全て成功したら入力順でCreatedに並ぶ", func(t *testing.T) {
		var nextID int64
		repo := &mockLinkRepository{
			createFn: func(link model.Link) (model.Link, error) {
				link.ID = int(atomic.AddInt64(&nextID, 1))
				return link, nil
			},
		}
		fetcher := noopFetcher()
		s := NewLinkService(repo, fetcher)

		urls := []string{"https://a.example.com", "https://b.example.com", "https://c.example.com"}
		result := s.BulkCreateLinks(urls, []string{"Go"})

		if len(result.Failed) != 0 {
			t.Fatalf("expected no failures, got %+v", result.Failed)
		}
		if len(result.Created) != len(urls) {
			t.Fatalf("expected %d created, got %d", len(urls), len(result.Created))
		}
		for i, url := range urls {
			if result.Created[i].URL != url {
				t.Errorf("Created[%d].URL = %q, want %q (order not preserved)", i, result.Created[i].URL, url)
			}
			if len(result.Created[i].Tags) != 1 || result.Created[i].Tags[0] != "Go" {
				t.Errorf("Created[%d].Tags = %+v, want [Go]", i, result.Created[i].Tags)
			}
		}
	})

	t.Run("一部URLが不正でも他は成功する", func(t *testing.T) {
		repo := &mockLinkRepository{
			createFn: func(link model.Link) (model.Link, error) {
				return link, nil
			},
		}
		s := NewLinkService(repo, noopFetcher())

		urls := []string{"https://a.example.com", "", "https://c.example.com"}
		result := s.BulkCreateLinks(urls, nil)

		if len(result.Created) != 2 {
			t.Fatalf("expected 2 created, got %d: %+v", len(result.Created), result.Created)
		}
		if len(result.Failed) != 1 || result.Failed[0].URL != "" {
			t.Fatalf("expected the empty url to be recorded as a failure, got %+v", result.Failed)
		}
	})

	t.Run("同時実行数の上限を守りつつ並行実行される", func(t *testing.T) {
		var mu sync.Mutex
		current, max := 0, 0
		fetcher := &mockMetadataFetcher{
			fetchFn: func(url string) (model.Metadata, error) {
				mu.Lock()
				current++
				if current > max {
					max = current
				}
				mu.Unlock()

				time.Sleep(20 * time.Millisecond)

				mu.Lock()
				current--
				mu.Unlock()

				return model.Metadata{}, nil
			},
		}
		repo := &mockLinkRepository{
			createFn: func(link model.Link) (model.Link, error) {
				return link, nil
			},
		}
		s := NewLinkService(repo, fetcher)

		urls := make([]string, maxConcurrentFetches*2)
		for i := range urls {
			urls[i] = fmt.Sprintf("https://example.com/%d", i)
		}

		result := s.BulkCreateLinks(urls, nil)

		if len(result.Created) != len(urls) {
			t.Fatalf("expected %d created, got %d", len(urls), len(result.Created))
		}
		if max <= 1 {
			t.Error("expected fetches to run concurrently, but max observed concurrency was 1")
		}
		if max > maxConcurrentFetches {
			t.Errorf("exceeded concurrency limit: observed %d, want <= %d", max, maxConcurrentFetches)
		}
	})
}

func TestCheckLinks(t *testing.T) {
	links := []model.Link{
		{ID: 1, URL: "https://ok.example.com"},
		{ID: 2, URL: "https://broken.example.com"},
		{ID: 3, URL: "https://error.example.com"},
	}

	var mu sync.Mutex
	gotStatuses := make(map[int]string)

	repo := &mockLinkRepository{
		listFn: func(query, tag string) ([]model.Link, error) {
			return links, nil
		},
		updateStatusFn: func(id int, status string, checkedAt time.Time) error {
			mu.Lock()
			gotStatuses[id] = status
			mu.Unlock()
			return nil
		},
	}
	fetcher := &mockMetadataFetcher{
		checkAliveFn: func(url string) (bool, error) {
			switch url {
			case "https://ok.example.com":
				return true, nil
			case "https://broken.example.com":
				return false, nil
			default:
				return false, errors.New("timeout")
			}
		},
	}
	s := NewLinkService(repo, fetcher)

	got, err := s.CheckLinks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(links) {
		t.Fatalf("expected %d links returned, got %d", len(links), len(got))
	}

	want := map[int]string{1: model.StatusOK, 2: model.StatusBroken, 3: model.StatusBroken}
	mu.Lock()
	defer mu.Unlock()
	for id, wantStatus := range want {
		if gotStatuses[id] != wantStatus {
			t.Errorf("link %d status = %q, want %q", id, gotStatuses[id], wantStatus)
		}
	}
}

func TestListLinks(t *testing.T) {
	want := []model.Link{{ID: 1, URL: "https://example.com"}}
	var gotQuery, gotTag string
	repo := &mockLinkRepository{
		listFn: func(query, tag string) ([]model.Link, error) {
			gotQuery, gotTag = query, tag
			return want, nil
		},
	}
	s := NewLinkService(repo, unusedFetcher(t))

	got, err := s.ListLinks("example", "Go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ID != want[0].ID {
		t.Errorf("unexpected links: %+v", got)
	}
	if gotQuery != "example" || gotTag != "Go" {
		t.Errorf("expected query/tag to be passed through, got query=%q tag=%q", gotQuery, gotTag)
	}
}

func TestUpdateLink(t *testing.T) {
	t.Run("成功したら更新後のリンクを返す", func(t *testing.T) {
		repo := &mockLinkRepository{
			updateFn: func(link model.Link) (model.Link, error) {
				return link, nil
			},
		}
		s := NewLinkService(repo, unusedFetcher(t))

		got, err := s.UpdateLink(1, "https://example.com", "Example", "memo", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != 1 {
			t.Errorf("unexpected link: %+v", got)
		}
	})

	t.Run("urlが空ならバリデーションエラーを返す", func(t *testing.T) {
		repo := &mockLinkRepository{
			updateFn: func(link model.Link) (model.Link, error) {
				t.Fatal("repository should not be called")
				return model.Link{}, nil
			},
		}
		s := NewLinkService(repo, unusedFetcher(t))

		_, err := s.UpdateLink(1, "", "Example", "memo", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("対象が存在しなければErrNotFoundを返す", func(t *testing.T) {
		repo := &mockLinkRepository{
			updateFn: func(link model.Link) (model.Link, error) {
				return model.Link{}, sql.ErrNoRows
			},
		}
		s := NewLinkService(repo, unusedFetcher(t))

		_, err := s.UpdateLink(1, "https://example.com", "Example", "memo", nil)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestDeleteLink(t *testing.T) {
	t.Run("成功したらnilを返す", func(t *testing.T) {
		repo := &mockLinkRepository{
			deleteFn: func(id int) error {
				return nil
			},
		}
		s := NewLinkService(repo, unusedFetcher(t))

		if err := s.DeleteLink(1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("対象が存在しなければErrNotFoundを返す", func(t *testing.T) {
		repo := &mockLinkRepository{
			deleteFn: func(id int) error {
				return sql.ErrNoRows
			},
		}
		s := NewLinkService(repo, unusedFetcher(t))

		err := s.DeleteLink(1)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}
