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

	if _, err := db.Exec("TRUNCATE TABLE links RESTART IDENTITY"); err != nil {
		t.Fatalf("failed to truncate links table: %v", err)
	}

	return NewLinkRepository(db)
}

func TestLinkRepositoryCreate(t *testing.T) {
	repo := setupTestRepo(t)
	now := time.Now().Truncate(time.Second)

	got, err := repo.Create(model.Link{
		URL:       "https://example.com",
		Title:     "Example",
		Memo:      "memo",
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID == 0 {
		t.Errorf("expected ID to be assigned, got %d", got.ID)
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

	links, err := repo.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(links) != 2 {
		t.Errorf("expected 2 links, got %d", len(links))
	}
}

func TestLinkRepositoryUpdate(t *testing.T) {
	repo := setupTestRepo(t)
	now := time.Now().Truncate(time.Second)

	created, err := repo.Create(model.Link{URL: "https://example.com", Title: "Example", CreatedAt: now, UpdatedAt: now})
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

	t.Run("存在しないIDならsql.ErrNoRowsを返す", func(t *testing.T) {
		_, err := repo.Update(model.Link{ID: 999999, URL: "https://example.com", UpdatedAt: time.Now()})
		if err != sql.ErrNoRows {
			t.Errorf("expected sql.ErrNoRows, got %v", err)
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

		links, err := repo.List()
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
