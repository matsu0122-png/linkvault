package repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/matsu0122-png/linkvault/backend/config"
	"github.com/matsu0122-png/linkvault/backend/database"
	"github.com/matsu0122-png/linkvault/backend/model"
)

func setupTestRepo(t *testing.T) *LinkRepository {
	t.Helper()

	cfg := config.Load()
	cfg.DBName = "linkvault_test"

	db, err := database.Connect(cfg)
	if err != nil {
		t.Skipf("test database not available: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if _, err := db.Exec("TRUNCATE TABLE links, tags RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate links table: %v", err)
	}

	return NewLinkRepository(db)
}

func TestLinkRepositoryCreate(t *testing.T) {
	repo := setupTestRepo(t)
	now := time.Now().Truncate(time.Second)

	got, err := repo.Create(model.Link{
		URL:         "https://example.com",
		Title:       "Example",
		Memo:        "memo",
		Description: "an example page",
		ImageURL:    "https://example.com/og.png",
		FaviconURL:  "https://example.com/favicon.png",
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID == 0 {
		t.Errorf("expected ID to be assigned, got %d", got.ID)
	}
	if got.Description != "an example page" || got.ImageURL != "https://example.com/og.png" || got.FaviconURL != "https://example.com/favicon.png" {
		t.Errorf("unexpected metadata: %+v", got)
	}
}

func TestLinkRepositoryList(t *testing.T) {
	repo := setupTestRepo(t)
	now := time.Now().Truncate(time.Second)

	for _, url := range []string{"https://a.example.com", "https://b.example.com"} {
		if _, err := repo.Create(model.Link{URL: url, Title: "t", CreatedAt: now, UpdatedAt: now}); err != nil {
			t.Fatalf("setup: failed to create link: %v", err)
		}
	}

	links, err := repo.List("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(links) != 2 {
		t.Errorf("expected 2 links, got %d", len(links))
	}
}

func TestLinkRepositoryListFilter(t *testing.T) {
	repo := setupTestRepo(t)
	now := time.Now().Truncate(time.Second)

	if _, err := repo.Create(model.Link{URL: "https://go.dev", Title: "Go Concurrency", Tags: []string{"Go", "backend"}, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("setup: failed to create link: %v", err)
	}
	if _, err := repo.Create(model.Link{URL: "https://react.dev", Title: "React Docs", Tags: []string{"frontend"}, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("setup: failed to create link: %v", err)
	}

	t.Run("キーワードでtitleを部分一致検索できる", func(t *testing.T) {
		links, err := repo.List("concurrency", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(links) != 1 || links[0].Title != "Go Concurrency" {
			t.Errorf("unexpected links: %+v", links)
		}
	})

	t.Run("タグで絞り込める", func(t *testing.T) {
		links, err := repo.List("", "frontend")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(links) != 1 || links[0].Title != "React Docs" {
			t.Errorf("unexpected links: %+v", links)
		}
	})
}

func TestLinkRepositoryUpdate(t *testing.T) {
	repo := setupTestRepo(t)
	now := time.Now().Truncate(time.Second)

	created, err := repo.Create(model.Link{
		URL:         "https://example.com",
		Title:       "Example",
		Description: "an example page",
		ImageURL:    "https://example.com/og.png",
		FaviconURL:  "https://example.com/favicon.png",
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("setup: failed to create link: %v", err)
	}

	t.Run("既存のリンクを更新できる", func(t *testing.T) {
		updated, err := repo.Update(model.Link{
			ID:        created.ID,
			URL:       "https://updated.example.com",
			Title:     "Updated",
			Memo:      "updated memo",
			UpdatedAt: time.Now(),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.URL != "https://updated.example.com" || updated.Title != "Updated" {
			t.Errorf("unexpected link: %+v", updated)
		}
	})

	t.Run("description/image/faviconはUpdateで上書きされず保持される", func(t *testing.T) {
		updated, err := repo.Update(model.Link{
			ID:        created.ID,
			URL:       created.URL,
			Title:     "Updated Again",
			Memo:      "another memo",
			UpdatedAt: time.Now(),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Description != "an example page" || updated.ImageURL != "https://example.com/og.png" || updated.FaviconURL != "https://example.com/favicon.png" {
			t.Errorf("expected metadata to be preserved, got %+v", updated)
		}
	})

	t.Run("存在しないIDならsql.ErrNoRowsを返す", func(t *testing.T) {
		_, err := repo.Update(model.Link{ID: 999999, URL: "https://example.com", UpdatedAt: time.Now()})
		if err != sql.ErrNoRows {
			t.Errorf("expected sql.ErrNoRows, got %v", err)
		}
	})

	t.Run("タグを入れ替えられる", func(t *testing.T) {
		if _, err := repo.Update(model.Link{ID: created.ID, URL: created.URL, Title: created.Title, Tags: []string{"Go", "web"}, UpdatedAt: time.Now()}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		updated, err := repo.Update(model.Link{ID: created.ID, URL: created.URL, Title: created.Title, Tags: []string{"backend"}, UpdatedAt: time.Now()})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(updated.Tags) != 1 || updated.Tags[0] != "backend" {
			t.Errorf("expected tags to be replaced with [backend], got %+v", updated.Tags)
		}

		links, err := repo.List("", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, l := range links {
			if l.ID == created.ID && (len(l.Tags) != 1 || l.Tags[0] != "backend") {
				t.Errorf("expected stored tags to be [backend], got %+v", l.Tags)
			}
		}
	})
}

func TestLinkRepositoryDelete(t *testing.T) {
	repo := setupTestRepo(t)
	now := time.Now().Truncate(time.Second)

	created, err := repo.Create(model.Link{URL: "https://example.com", Title: "Example", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("setup: failed to create link: %v", err)
	}

	t.Run("既存のリンクを削除できる", func(t *testing.T) {
		if err := repo.Delete(created.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		links, err := repo.List("", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(links) != 0 {
			t.Errorf("expected 0 links after delete, got %d", len(links))
		}
	})

	t.Run("存在しないIDならsql.ErrNoRowsを返す", func(t *testing.T) {
		err := repo.Delete(999999)
		if err != sql.ErrNoRows {
			t.Errorf("expected sql.ErrNoRows, got %v", err)
		}
	})
}
