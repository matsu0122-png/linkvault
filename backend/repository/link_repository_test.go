package repository

import (
	"database/sql"
	"fmt"
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

	if _, err := db.Exec("TRUNCATE TABLE links, tags, collections RESTART IDENTITY CASCADE"); err != nil {
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
	if got.Status != model.StatusUnknown || got.CheckedAt != nil {
		t.Errorf("expected a fresh link to be unchecked, got status=%q checkedAt=%v", got.Status, got.CheckedAt)
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

func TestLinkRepositoryUpdateStatus(t *testing.T) {
	repo := setupTestRepo(t)
	now := time.Now().Truncate(time.Second)

	created, err := repo.Create(model.Link{URL: "https://example.com", Title: "Example", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("setup: failed to create link: %v", err)
	}

	t.Run("statusとchecked_atを更新できる", func(t *testing.T) {
		checkedAt := time.Now().Truncate(time.Second)
		if err := repo.UpdateStatus(created.ID, model.StatusOK, checkedAt); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		links, err := repo.List("", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(links) != 1 || links[0].Status != model.StatusOK {
			t.Fatalf("expected status to be ok, got %+v", links)
		}
		if links[0].CheckedAt == nil || !links[0].CheckedAt.Equal(checkedAt) {
			t.Errorf("expected checkedAt %v, got %v", checkedAt, links[0].CheckedAt)
		}
	})

	t.Run("存在しないIDならsql.ErrNoRowsを返す", func(t *testing.T) {
		err := repo.UpdateStatus(999999, model.StatusBroken, time.Now())
		if err != sql.ErrNoRows {
			t.Errorf("expected sql.ErrNoRows, got %v", err)
		}
	})

	t.Run("Updateではstatus/checked_atが上書きされず保持される", func(t *testing.T) {
		checkedAt := time.Now().Truncate(time.Second)
		if err := repo.UpdateStatus(created.ID, model.StatusBroken, checkedAt); err != nil {
			t.Fatalf("setup: failed to set status: %v", err)
		}

		updated, err := repo.Update(model.Link{ID: created.ID, URL: created.URL, Title: "Renamed", UpdatedAt: time.Now()})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Status != model.StatusBroken || updated.CheckedAt == nil || !updated.CheckedAt.Equal(checkedAt) {
			t.Errorf("expected status/checkedAt to be preserved, got status=%q checkedAt=%v", updated.Status, updated.CheckedAt)
		}
	})
}

func TestLinkRepositoryListPage(t *testing.T) {
	repo := setupTestRepo(t)

	// 25 links created 1 second apart (oldest to newest) so created_at
	// order is deterministic and distinct across rows.
	base := time.Now().Truncate(time.Second)
	created := make([]model.Link, 25)
	for i := range created {
		ts := base.Add(time.Duration(i) * time.Second)
		link, err := repo.Create(model.Link{
			URL:       fmt.Sprintf("https://example.com/%d", i),
			Title:     fmt.Sprintf("Title %02d", i),
			CreatedAt: ts,
			UpdatedAt: ts,
		})
		if err != nil {
			t.Fatalf("setup: failed to create link %d: %v", i, err)
		}
		created[i] = link
	}

	t.Run("1ページ目はlimit件返し、totalは全件数", func(t *testing.T) {
		links, total, err := repo.ListPage("", "", 0, model.SortCreatedDesc, 20, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 25 {
			t.Errorf("total = %d, want 25", total)
		}
		if len(links) != 20 {
			t.Fatalf("expected 20 links, got %d", len(links))
		}
		// created_at_desc: 最後に作った (index 24) リンクが先頭に来る
		if links[0].ID != created[24].ID {
			t.Errorf("first link = %+v, want id=%d", links[0], created[24].ID)
		}
	})

	t.Run("2ページ目は残りの件数を返す", func(t *testing.T) {
		links, total, err := repo.ListPage("", "", 0, model.SortCreatedDesc, 20, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 25 {
			t.Errorf("total = %d, want 25", total)
		}
		if len(links) != 5 {
			t.Errorf("expected 5 links on page 2, got %d", len(links))
		}
	})

	t.Run("ページをまたいでも重複・欠落なく全件をカバーする", func(t *testing.T) {
		seen := map[int]bool{}
		for offset := 0; offset < 25; offset += 10 {
			links, _, err := repo.ListPage("", "", 0, model.SortCreatedDesc, 10, offset)
			if err != nil {
				t.Fatalf("unexpected error at offset %d: %v", offset, err)
			}
			for _, l := range links {
				if seen[l.ID] {
					t.Errorf("id %d seen more than once across pages", l.ID)
				}
				seen[l.ID] = true
			}
		}
		if len(seen) != 25 {
			t.Errorf("expected 25 unique links across all pages, got %d", len(seen))
		}
	})

	t.Run("limitを変えても総件数は変わらない", func(t *testing.T) {
		_, total, err := repo.ListPage("", "", 0, model.SortCreatedDesc, 5, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 25 {
			t.Errorf("total = %d, want 25", total)
		}
	})

	t.Run("sort=created_at_ascなら最初に作ったリンクが先頭に来る", func(t *testing.T) {
		links, _, err := repo.ListPage("", "", 0, model.SortCreatedAsc, 5, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(links) == 0 || links[0].ID != created[0].ID {
			t.Errorf("first link id = %v, want %d", links, created[0].ID)
		}
	})

	t.Run("sort=title_ascならタイトルの昇順で並ぶ", func(t *testing.T) {
		links, _, err := repo.ListPage("", "", 0, model.SortTitleAsc, 3, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(links) != 3 || links[0].Title != "Title 00" || links[1].Title != "Title 01" || links[2].Title != "Title 02" {
			t.Errorf("unexpected order: %+v", links)
		}
	})

	t.Run("sort=title_descならタイトルの降順で並ぶ", func(t *testing.T) {
		links, _, err := repo.ListPage("", "", 0, model.SortTitleDesc, 1, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(links) != 1 || links[0].Title != "Title 24" {
			t.Errorf("unexpected order: %+v", links)
		}
	})

	t.Run("検索条件とページネーションを組み合わせても件数がずれない", func(t *testing.T) {
		// "Title 1" は Title 10-19 の10件にマッチする
		links, total, err := repo.ListPage("Title 1", "", 0, model.SortCreatedDesc, 5, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 10 {
			t.Errorf("total = %d, want 10", total)
		}
		if len(links) != 5 {
			t.Errorf("expected 5 links, got %d", len(links))
		}

		links2, total2, err := repo.ListPage("Title 1", "", 0, model.SortCreatedDesc, 5, 5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total2 != 10 {
			t.Errorf("total = %d, want 10", total2)
		}
		if len(links2) != 5 {
			t.Errorf("expected 5 links, got %d", len(links2))
		}
	})

	t.Run("タグ絞り込みとページネーションを組み合わせても件数がずれない", func(t *testing.T) {
		if _, err := repo.Update(model.Link{ID: created[0].ID, URL: created[0].URL, Title: created[0].Title, Tags: []string{"paged"}, UpdatedAt: time.Now()}); err != nil {
			t.Fatalf("setup: failed to tag link: %v", err)
		}
		if _, err := repo.Update(model.Link{ID: created[1].ID, URL: created[1].URL, Title: created[1].Title, Tags: []string{"paged"}, UpdatedAt: time.Now()}); err != nil {
			t.Fatalf("setup: failed to tag link: %v", err)
		}

		links, total, err := repo.ListPage("", "paged", 0, model.SortCreatedDesc, 20, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 2 {
			t.Errorf("total = %d, want 2", total)
		}
		if len(links) != 2 {
			t.Errorf("expected 2 links, got %d", len(links))
		}
	})

	t.Run("collection条件と組み合わせても件数がずれない", func(t *testing.T) {
		collRepo := NewCollectionRepository(repo.db)
		coll, err := collRepo.Create(model.Collection{Name: "paged-collection", CreatedAt: time.Now(), UpdatedAt: time.Now()})
		if err != nil {
			t.Fatalf("setup: failed to create collection: %v", err)
		}
		for _, l := range created[:3] {
			if err := collRepo.AddLink(coll.ID, l.ID); err != nil {
				t.Fatalf("setup: failed to add link to collection: %v", err)
			}
		}

		links, total, err := repo.ListPage("", "", coll.ID, model.SortCreatedDesc, 20, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 3 {
			t.Errorf("total = %d, want 3", total)
		}
		if len(links) != 3 {
			t.Errorf("expected 3 links, got %d", len(links))
		}
	})

	t.Run("範囲外のoffsetは空のスライスを返す", func(t *testing.T) {
		links, total, err := repo.ListPage("", "", 0, model.SortCreatedDesc, 20, 1000)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 25 {
			t.Errorf("total = %d, want 25", total)
		}
		if len(links) != 0 {
			t.Errorf("expected 0 links, got %d", len(links))
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
