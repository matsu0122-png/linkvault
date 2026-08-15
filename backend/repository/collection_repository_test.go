package repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/matsu0122-png/linkvault/backend/model"
)

func setupTestCollectionRepo(t *testing.T) (*CollectionRepository, *LinkRepository) {
	t.Helper()
	linkRepo := setupTestRepo(t)
	return NewCollectionRepository(linkRepo.db), linkRepo
}

func TestCollectionRepositoryCreate(t *testing.T) {
	collRepo, _ := setupTestCollectionRepo(t)
	now := time.Now().Truncate(time.Second)

	t.Run("成功したら保存されたCollectionを返す", func(t *testing.T) {
		got, err := collRepo.Create(model.Collection{Name: "Go学習", Description: "Go関連の学習資料", CreatedAt: now, UpdatedAt: now})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID == 0 {
			t.Errorf("expected ID to be assigned, got %d", got.ID)
		}
		if got.LinkCount != 0 {
			t.Errorf("expected link_count 0, got %d", got.LinkCount)
		}
	})

	t.Run("同じ名前は重複エラーになる", func(t *testing.T) {
		if _, err := collRepo.Create(model.Collection{Name: "重複テスト", CreatedAt: now, UpdatedAt: now}); err != nil {
			t.Fatalf("setup: failed to create collection: %v", err)
		}

		_, err := collRepo.Create(model.Collection{Name: "重複テスト", CreatedAt: now, UpdatedAt: now})
		if err != ErrDuplicateName {
			t.Errorf("expected ErrDuplicateName, got %v", err)
		}
	})

	t.Run("parent_idを指定して子として作成できる", func(t *testing.T) {
		parent, err := collRepo.Create(model.Collection{Name: "親", CreatedAt: now, UpdatedAt: now})
		if err != nil {
			t.Fatalf("setup: failed to create parent: %v", err)
		}

		child, err := collRepo.Create(model.Collection{Name: "子", ParentID: &parent.ID, CreatedAt: now, UpdatedAt: now})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if child.ParentID == nil || *child.ParentID != parent.ID {
			t.Errorf("expected parent_id=%d, got %v", parent.ID, child.ParentID)
		}
	})

	t.Run("存在しないparent_idを指定するとErrParentNotFoundを返す", func(t *testing.T) {
		badParent := 999999
		_, err := collRepo.Create(model.Collection{Name: "孤児", ParentID: &badParent, CreatedAt: now, UpdatedAt: now})
		if err != ErrParentNotFound {
			t.Errorf("expected ErrParentNotFound, got %v", err)
		}
	})
}

func TestCollectionRepositoryParentChild(t *testing.T) {
	collRepo, _ := setupTestCollectionRepo(t)
	now := time.Now().Truncate(time.Second)

	t.Run("親を削除すると子は最上位へ昇格する", func(t *testing.T) {
		parent, err := collRepo.Create(model.Collection{Name: "親2", CreatedAt: now, UpdatedAt: now})
		if err != nil {
			t.Fatalf("setup: failed to create parent: %v", err)
		}
		child, err := collRepo.Create(model.Collection{Name: "子2", ParentID: &parent.ID, CreatedAt: now, UpdatedAt: now})
		if err != nil {
			t.Fatalf("setup: failed to create child: %v", err)
		}

		if err := collRepo.Delete(parent.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, err := collRepo.Get(child.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ParentID != nil {
			t.Errorf("expected parent_id to be promoted to nil, got %v", got.ParentID)
		}
	})

	t.Run("Listでparent_idが正しく返る", func(t *testing.T) {
		parent, err := collRepo.Create(model.Collection{Name: "親3", CreatedAt: now, UpdatedAt: now})
		if err != nil {
			t.Fatalf("setup: failed to create parent: %v", err)
		}
		child, err := collRepo.Create(model.Collection{Name: "子3", ParentID: &parent.ID, CreatedAt: now, UpdatedAt: now})
		if err != nil {
			t.Fatalf("setup: failed to create child: %v", err)
		}

		collections, err := collRepo.List(0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var gotChild *model.Collection
		for i := range collections {
			if collections[i].ID == child.ID {
				gotChild = &collections[i]
			}
		}
		if gotChild == nil {
			t.Fatal("child not found in list")
		}
		if gotChild.ParentID == nil || *gotChild.ParentID != parent.ID {
			t.Errorf("expected parent_id=%d, got %v", parent.ID, gotChild.ParentID)
		}
	})
}

func TestCollectionRepositoryList(t *testing.T) {
	collRepo, linkRepo := setupTestCollectionRepo(t)
	now := time.Now().Truncate(time.Second)

	go学習, err := collRepo.Create(model.Collection{Name: "Go学習", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("setup: failed to create collection: %v", err)
	}
	if _, err := collRepo.Create(model.Collection{Name: "仕事", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("setup: failed to create collection: %v", err)
	}

	link, err := linkRepo.Create(model.Link{URL: "https://go.dev", Title: "Go", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("setup: failed to create link: %v", err)
	}

	t.Run("link_idを指定しなければ全件返す", func(t *testing.T) {
		collections, err := collRepo.List(0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(collections) != 2 {
			t.Errorf("expected 2 collections, got %d", len(collections))
		}
	})

	t.Run("link_countを正しく集計する", func(t *testing.T) {
		if err := collRepo.AddLink(go学習.ID, link.ID); err != nil {
			t.Fatalf("setup: failed to add link: %v", err)
		}

		collections, err := collRepo.List(0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, c := range collections {
			if c.ID == go学習.ID && c.LinkCount != 1 {
				t.Errorf("expected link_count 1 for Go学習, got %d", c.LinkCount)
			}
		}
	})

	t.Run("link_idを指定するとそのLinkが所属するCollectionのみ返す", func(t *testing.T) {
		collections, err := collRepo.List(link.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(collections) != 1 || collections[0].ID != go学習.ID {
			t.Errorf("unexpected collections: %+v", collections)
		}
	})
}

func TestCollectionRepositoryGet(t *testing.T) {
	collRepo, _ := setupTestCollectionRepo(t)
	now := time.Now().Truncate(time.Second)

	created, err := collRepo.Create(model.Collection{Name: "Go学習", Description: "desc", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("setup: failed to create collection: %v", err)
	}

	t.Run("成功したらCollectionを返す", func(t *testing.T) {
		got, err := collRepo.Get(created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "Go学習" || got.Description != "desc" {
			t.Errorf("unexpected collection: %+v", got)
		}
	})

	t.Run("存在しないIDならsql.ErrNoRowsを返す", func(t *testing.T) {
		_, err := collRepo.Get(999999)
		if err != sql.ErrNoRows {
			t.Errorf("expected sql.ErrNoRows, got %v", err)
		}
	})
}

func TestCollectionRepositoryUpdate(t *testing.T) {
	collRepo, _ := setupTestCollectionRepo(t)
	now := time.Now().Truncate(time.Second)

	created, err := collRepo.Create(model.Collection{Name: "Go学習", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("setup: failed to create collection: %v", err)
	}

	t.Run("名前と説明を更新できる", func(t *testing.T) {
		updated, err := collRepo.Update(model.Collection{ID: created.ID, Name: "Go学習(更新)", Description: "更新後", UpdatedAt: time.Now()})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name != "Go学習(更新)" || updated.Description != "更新後" {
			t.Errorf("unexpected collection: %+v", updated)
		}
	})

	t.Run("存在しないIDならsql.ErrNoRowsを返す", func(t *testing.T) {
		_, err := collRepo.Update(model.Collection{ID: 999999, Name: "x", UpdatedAt: time.Now()})
		if err != sql.ErrNoRows {
			t.Errorf("expected sql.ErrNoRows, got %v", err)
		}
	})

	t.Run("他のCollectionと同じ名前に変更しようとすると重複エラーになる", func(t *testing.T) {
		other, err := collRepo.Create(model.Collection{Name: "仕事", CreatedAt: now, UpdatedAt: now})
		if err != nil {
			t.Fatalf("setup: failed to create collection: %v", err)
		}

		_, err = collRepo.Update(model.Collection{ID: other.ID, Name: "Go学習(更新)", UpdatedAt: time.Now()})
		if err != ErrDuplicateName {
			t.Errorf("expected ErrDuplicateName, got %v", err)
		}
	})
}

func TestCollectionRepositoryDelete(t *testing.T) {
	collRepo, linkRepo := setupTestCollectionRepo(t)
	now := time.Now().Truncate(time.Second)

	t.Run("Collectionを削除してもLink本体は残る", func(t *testing.T) {
		coll, err := collRepo.Create(model.Collection{Name: "Go学習", CreatedAt: now, UpdatedAt: now})
		if err != nil {
			t.Fatalf("setup: failed to create collection: %v", err)
		}
		link, err := linkRepo.Create(model.Link{URL: "https://go.dev", Title: "Go", CreatedAt: now, UpdatedAt: now})
		if err != nil {
			t.Fatalf("setup: failed to create link: %v", err)
		}
		if err := collRepo.AddLink(coll.ID, link.ID); err != nil {
			t.Fatalf("setup: failed to add link: %v", err)
		}

		if err := collRepo.Delete(coll.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, err := collRepo.Get(coll.ID); err != sql.ErrNoRows {
			t.Errorf("expected collection to be gone, got err=%v", err)
		}

		links, err := linkRepo.List("", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		found := false
		for _, l := range links {
			if l.ID == link.ID {
				found = true
			}
		}
		if !found {
			t.Errorf("expected link %d to still exist after collection deletion", link.ID)
		}
	})

	t.Run("存在しないIDならsql.ErrNoRowsを返す", func(t *testing.T) {
		err := collRepo.Delete(999999)
		if err != sql.ErrNoRows {
			t.Errorf("expected sql.ErrNoRows, got %v", err)
		}
	})
}

func TestCollectionRepositoryAddLink(t *testing.T) {
	collRepo, linkRepo := setupTestCollectionRepo(t)
	now := time.Now().Truncate(time.Second)

	coll, err := collRepo.Create(model.Collection{Name: "Go学習", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("setup: failed to create collection: %v", err)
	}
	link, err := linkRepo.Create(model.Link{URL: "https://go.dev", Title: "Go", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("setup: failed to create link: %v", err)
	}

	t.Run("成功する", func(t *testing.T) {
		if err := collRepo.AddLink(coll.ID, link.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("同じ組み合わせを2回追加してもエラーにならず1件のまま", func(t *testing.T) {
		if err := collRepo.AddLink(coll.ID, link.ID); err != nil {
			t.Fatalf("unexpected error on duplicate add: %v", err)
		}

		got, err := collRepo.Get(coll.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.LinkCount != 1 {
			t.Errorf("expected link_count 1 after duplicate add, got %d", got.LinkCount)
		}
	})

	t.Run("1つのLinkを複数Collectionへ追加できる", func(t *testing.T) {
		other, err := collRepo.Create(model.Collection{Name: "お気に入り", CreatedAt: now, UpdatedAt: now})
		if err != nil {
			t.Fatalf("setup: failed to create collection: %v", err)
		}
		if err := collRepo.AddLink(other.ID, link.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		collections, err := collRepo.List(link.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(collections) != 2 {
			t.Errorf("expected link to belong to 2 collections, got %d", len(collections))
		}
	})

	t.Run("存在しないCollectionへの追加はErrCollectionNotFoundを返す", func(t *testing.T) {
		err := collRepo.AddLink(999999, link.ID)
		if err != ErrCollectionNotFound {
			t.Errorf("expected ErrCollectionNotFound, got %v", err)
		}
	})

	t.Run("存在しないLinkの追加はErrLinkNotFoundを返す", func(t *testing.T) {
		err := collRepo.AddLink(coll.ID, 999999)
		if err != ErrLinkNotFound {
			t.Errorf("expected ErrLinkNotFound, got %v", err)
		}
	})
}

func TestCollectionRepositoryRemoveLink(t *testing.T) {
	collRepo, linkRepo := setupTestCollectionRepo(t)
	now := time.Now().Truncate(time.Second)

	coll, err := collRepo.Create(model.Collection{Name: "Go学習", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("setup: failed to create collection: %v", err)
	}
	link, err := linkRepo.Create(model.Link{URL: "https://go.dev", Title: "Go", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("setup: failed to create link: %v", err)
	}
	if err := collRepo.AddLink(coll.ID, link.ID); err != nil {
		t.Fatalf("setup: failed to add link: %v", err)
	}

	t.Run("成功したら関連付けが消える", func(t *testing.T) {
		if err := collRepo.RemoveLink(coll.ID, link.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, err := collRepo.Get(coll.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.LinkCount != 0 {
			t.Errorf("expected link_count 0, got %d", got.LinkCount)
		}
	})

	t.Run("紐付いていない組み合わせはsql.ErrNoRowsを返す", func(t *testing.T) {
		err := collRepo.RemoveLink(coll.ID, link.ID)
		if err != sql.ErrNoRows {
			t.Errorf("expected sql.ErrNoRows, got %v", err)
		}
	})
}
