package service

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/matsu0122-png/linkvault/backend/model"
	"github.com/matsu0122-png/linkvault/backend/repository"
)

type mockCollectionRepository struct {
	createFn     func(c model.Collection) (model.Collection, error)
	listFn       func(linkID int) ([]model.Collection, error)
	getFn        func(id int) (model.Collection, error)
	updateFn     func(c model.Collection) (model.Collection, error)
	deleteFn     func(id int) error
	addLinkFn    func(collectionID, linkID int) error
	removeLinkFn func(collectionID, linkID int) error
}

func (m *mockCollectionRepository) Create(c model.Collection) (model.Collection, error) {
	return m.createFn(c)
}

func (m *mockCollectionRepository) List(linkID int) ([]model.Collection, error) {
	return m.listFn(linkID)
}

func (m *mockCollectionRepository) Get(id int) (model.Collection, error) {
	return m.getFn(id)
}

func (m *mockCollectionRepository) Update(c model.Collection) (model.Collection, error) {
	return m.updateFn(c)
}

func (m *mockCollectionRepository) Delete(id int) error {
	return m.deleteFn(id)
}

func (m *mockCollectionRepository) AddLink(collectionID, linkID int) error {
	return m.addLinkFn(collectionID, linkID)
}

func (m *mockCollectionRepository) RemoveLink(collectionID, linkID int) error {
	return m.removeLinkFn(collectionID, linkID)
}

func TestCreateCollection(t *testing.T) {
	t.Run("成功したら保存されたCollectionを返す", func(t *testing.T) {
		repo := &mockCollectionRepository{
			createFn: func(c model.Collection) (model.Collection, error) {
				c.ID = 1
				return c, nil
			},
		}
		s := NewCollectionService(repo)

		got, err := s.CreateCollection("Go学習", "Go関連の学習資料", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != 1 || got.Name != "Go学習" || got.Description != "Go関連の学習資料" {
			t.Errorf("unexpected collection: %+v", got)
		}
	})

	t.Run("前後の空白を除去する", func(t *testing.T) {
		var gotName string
		repo := &mockCollectionRepository{
			createFn: func(c model.Collection) (model.Collection, error) {
				gotName = c.Name
				return c, nil
			},
		}
		s := NewCollectionService(repo)

		if _, err := s.CreateCollection("  Go学習  ", "", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotName != "Go学習" {
			t.Errorf("expected trimmed name, got %q", gotName)
		}
	})

	t.Run("名前が空ならバリデーションエラーを返す", func(t *testing.T) {
		repo := &mockCollectionRepository{
			createFn: func(c model.Collection) (model.Collection, error) {
				t.Fatal("repository should not be called")
				return model.Collection{}, nil
			},
		}
		s := NewCollectionService(repo)

		if _, err := s.CreateCollection("", "", nil); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("空白のみの名前はバリデーションエラーを返す", func(t *testing.T) {
		repo := &mockCollectionRepository{
			createFn: func(c model.Collection) (model.Collection, error) {
				t.Fatal("repository should not be called")
				return model.Collection{}, nil
			},
		}
		s := NewCollectionService(repo)

		if _, err := s.CreateCollection("   ", "", nil); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("名前が長すぎる場合はバリデーションエラーを返す", func(t *testing.T) {
		repo := &mockCollectionRepository{
			createFn: func(c model.Collection) (model.Collection, error) {
				t.Fatal("repository should not be called")
				return model.Collection{}, nil
			},
		}
		s := NewCollectionService(repo)

		longName := ""
		for i := 0; i < maxCollectionNameLength+1; i++ {
			longName += "a"
		}

		if _, err := s.CreateCollection(longName, "", nil); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("名前が重複していればErrDuplicateNameを返す", func(t *testing.T) {
		repo := &mockCollectionRepository{
			createFn: func(c model.Collection) (model.Collection, error) {
				return model.Collection{}, repository.ErrDuplicateName
			},
		}
		s := NewCollectionService(repo)

		_, err := s.CreateCollection("Go学習", "", nil)
		if !errors.Is(err, ErrDuplicateName) {
			t.Errorf("expected ErrDuplicateName, got %v", err)
		}
	})

	t.Run("parentIDをそのままrepositoryへ渡す", func(t *testing.T) {
		var gotParentID *int
		repo := &mockCollectionRepository{
			createFn: func(c model.Collection) (model.Collection, error) {
				gotParentID = c.ParentID
				return c, nil
			},
		}
		s := NewCollectionService(repo)

		parentID := 3
		if _, err := s.CreateCollection("Backend", "", &parentID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotParentID == nil || *gotParentID != 3 {
			t.Errorf("expected parentID=3, got %v", gotParentID)
		}
	})

	t.Run("存在しない親を指定するとErrParentNotFoundを返す", func(t *testing.T) {
		repo := &mockCollectionRepository{
			createFn: func(c model.Collection) (model.Collection, error) {
				return model.Collection{}, repository.ErrParentNotFound
			},
		}
		s := NewCollectionService(repo)

		parentID := 999999
		_, err := s.CreateCollection("Backend", "", &parentID)
		if !errors.Is(err, ErrParentNotFound) {
			t.Errorf("expected ErrParentNotFound, got %v", err)
		}
	})
}

func TestListCollections(t *testing.T) {
	want := []model.Collection{{ID: 1, Name: "Go学習"}}
	var gotLinkID int
	repo := &mockCollectionRepository{
		listFn: func(linkID int) ([]model.Collection, error) {
			gotLinkID = linkID
			return want, nil
		},
	}
	s := NewCollectionService(repo)

	got, err := s.ListCollections(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ID != want[0].ID {
		t.Errorf("unexpected collections: %+v", got)
	}
	if gotLinkID != 5 {
		t.Errorf("expected linkID=5, got %d", gotLinkID)
	}
}

func TestGetCollection(t *testing.T) {
	t.Run("成功したらCollectionを返す", func(t *testing.T) {
		repo := &mockCollectionRepository{
			getFn: func(id int) (model.Collection, error) {
				return model.Collection{ID: id, Name: "Go学習"}, nil
			},
		}
		s := NewCollectionService(repo)

		got, err := s.GetCollection(1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != 1 {
			t.Errorf("unexpected collection: %+v", got)
		}
	})

	t.Run("存在しなければErrCollectionNotFoundを返す", func(t *testing.T) {
		repo := &mockCollectionRepository{
			getFn: func(id int) (model.Collection, error) {
				return model.Collection{}, sql.ErrNoRows
			},
		}
		s := NewCollectionService(repo)

		_, err := s.GetCollection(999)
		if !errors.Is(err, ErrCollectionNotFound) {
			t.Errorf("expected ErrCollectionNotFound, got %v", err)
		}
	})
}

func TestUpdateCollection(t *testing.T) {
	t.Run("成功したら更新後のCollectionを返す", func(t *testing.T) {
		repo := &mockCollectionRepository{
			updateFn: func(c model.Collection) (model.Collection, error) {
				return c, nil
			},
		}
		s := NewCollectionService(repo)

		got, err := s.UpdateCollection(1, "仕事", "仕事関連")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "仕事" {
			t.Errorf("unexpected collection: %+v", got)
		}
	})

	t.Run("名前が空ならバリデーションエラーを返す", func(t *testing.T) {
		repo := &mockCollectionRepository{
			updateFn: func(c model.Collection) (model.Collection, error) {
				t.Fatal("repository should not be called")
				return model.Collection{}, nil
			},
		}
		s := NewCollectionService(repo)

		if _, err := s.UpdateCollection(1, "", ""); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("対象が存在しなければErrCollectionNotFoundを返す", func(t *testing.T) {
		repo := &mockCollectionRepository{
			updateFn: func(c model.Collection) (model.Collection, error) {
				return model.Collection{}, sql.ErrNoRows
			},
		}
		s := NewCollectionService(repo)

		_, err := s.UpdateCollection(1, "仕事", "")
		if !errors.Is(err, ErrCollectionNotFound) {
			t.Errorf("expected ErrCollectionNotFound, got %v", err)
		}
	})

	t.Run("名前が重複していればErrDuplicateNameを返す", func(t *testing.T) {
		repo := &mockCollectionRepository{
			updateFn: func(c model.Collection) (model.Collection, error) {
				return model.Collection{}, repository.ErrDuplicateName
			},
		}
		s := NewCollectionService(repo)

		_, err := s.UpdateCollection(1, "お気に入り", "")
		if !errors.Is(err, ErrDuplicateName) {
			t.Errorf("expected ErrDuplicateName, got %v", err)
		}
	})
}

func TestDeleteCollection(t *testing.T) {
	t.Run("成功したらnilを返す", func(t *testing.T) {
		repo := &mockCollectionRepository{
			deleteFn: func(id int) error {
				return nil
			},
		}
		s := NewCollectionService(repo)

		if err := s.DeleteCollection(1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("対象が存在しなければErrCollectionNotFoundを返す", func(t *testing.T) {
		repo := &mockCollectionRepository{
			deleteFn: func(id int) error {
				return sql.ErrNoRows
			},
		}
		s := NewCollectionService(repo)

		err := s.DeleteCollection(999)
		if !errors.Is(err, ErrCollectionNotFound) {
			t.Errorf("expected ErrCollectionNotFound, got %v", err)
		}
	})
}

func TestAddLink(t *testing.T) {
	t.Run("成功したらnilを返す", func(t *testing.T) {
		repo := &mockCollectionRepository{
			addLinkFn: func(collectionID, linkID int) error {
				return nil
			},
		}
		s := NewCollectionService(repo)

		if err := s.AddLink(1, 5); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("存在しないCollectionならErrCollectionNotFoundを返す", func(t *testing.T) {
		repo := &mockCollectionRepository{
			addLinkFn: func(collectionID, linkID int) error {
				return repository.ErrCollectionNotFound
			},
		}
		s := NewCollectionService(repo)

		err := s.AddLink(999, 5)
		if !errors.Is(err, ErrCollectionNotFound) {
			t.Errorf("expected ErrCollectionNotFound, got %v", err)
		}
	})

	t.Run("存在しないLinkならErrLinkNotFoundを返す", func(t *testing.T) {
		repo := &mockCollectionRepository{
			addLinkFn: func(collectionID, linkID int) error {
				return repository.ErrLinkNotFound
			},
		}
		s := NewCollectionService(repo)

		err := s.AddLink(1, 999999)
		if !errors.Is(err, ErrLinkNotFound) {
			t.Errorf("expected ErrLinkNotFound, got %v", err)
		}
	})
}

func TestRemoveLink(t *testing.T) {
	t.Run("成功したらnilを返す", func(t *testing.T) {
		repo := &mockCollectionRepository{
			removeLinkFn: func(collectionID, linkID int) error {
				return nil
			},
		}
		s := NewCollectionService(repo)

		if err := s.RemoveLink(1, 5); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("紐付いていなければErrCollectionNotFoundを返す", func(t *testing.T) {
		repo := &mockCollectionRepository{
			removeLinkFn: func(collectionID, linkID int) error {
				return sql.ErrNoRows
			},
		}
		s := NewCollectionService(repo)

		err := s.RemoveLink(1, 5)
		if !errors.Is(err, ErrCollectionNotFound) {
			t.Errorf("expected ErrCollectionNotFound, got %v", err)
		}
	})
}
