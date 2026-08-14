package service

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/matsu0122-png/linkvault/backend/model"
)

type mockLinkRepository struct {
	createFn func(link model.Link) (model.Link, error)
	listFn   func(query, tag string) ([]model.Link, error)
	updateFn func(link model.Link) (model.Link, error)
	deleteFn func(id int) error
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

func TestCreateLink(t *testing.T) {
	t.Run("成功したらリポジトリに保存されたリンクを返す", func(t *testing.T) {
		repo := &mockLinkRepository{
			createFn: func(link model.Link) (model.Link, error) {
				link.ID = 1
				return link, nil
			},
		}
		s := NewLinkService(repo)

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
		s := NewLinkService(repo)

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
		s := NewLinkService(repo)

		_, err := s.CreateLink("https://example.com", "Example", "memo", []string{" Go ", "web", "Go", "  "})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []string{"Go", "web"}
		if len(gotTags) != len(want) || gotTags[0] != want[0] || gotTags[1] != want[1] {
			t.Errorf("unexpected tags: %+v", gotTags)
		}
	})
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
	s := NewLinkService(repo)

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
		s := NewLinkService(repo)

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
		s := NewLinkService(repo)

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
		s := NewLinkService(repo)

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
		s := NewLinkService(repo)

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
		s := NewLinkService(repo)

		err := s.DeleteLink(1)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}
