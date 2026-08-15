package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matsu0122-png/linkvault/backend/model"
	"github.com/matsu0122-png/linkvault/backend/repository"
	"github.com/matsu0122-png/linkvault/backend/service"
)

// stubCollectionRepo is a minimal collectionRepository stand-in for
// exercising the handler's HTTP status-code mapping.
type stubCollectionRepo struct {
	createFn func(c model.Collection) (model.Collection, error)
}

func (s *stubCollectionRepo) Create(c model.Collection) (model.Collection, error) {
	return s.createFn(c)
}
func (s *stubCollectionRepo) List(linkID int) ([]model.Collection, error) { return nil, nil }
func (s *stubCollectionRepo) Get(id int) (model.Collection, error)        { return model.Collection{}, nil }
func (s *stubCollectionRepo) Update(c model.Collection) (model.Collection, error) {
	return model.Collection{}, nil
}
func (s *stubCollectionRepo) Delete(id int) error                       { return nil }
func (s *stubCollectionRepo) AddLink(collectionID, linkID int) error    { return nil }
func (s *stubCollectionRepo) RemoveLink(collectionID, linkID int) error { return nil }

func newTestCollectionHandler(repo *stubCollectionRepo) *CollectionHandler {
	return NewCollectionHandler(service.NewCollectionService(repo))
}

func TestCreateCollectionHandler(t *testing.T) {
	t.Run("名前が空なら400とバリデーションメッセージを返す", func(t *testing.T) {
		h := newTestCollectionHandler(&stubCollectionRepo{
			createFn: func(c model.Collection) (model.Collection, error) {
				t.Fatal("repository should not be called")
				return model.Collection{}, nil
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/api/collections", strings.NewReader(`{"name":""}`))
		rec := httptest.NewRecorder()
		h.CreateCollection(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "name is required") {
			t.Errorf("expected validation message in body, got %q", rec.Body.String())
		}
	})

	t.Run("名前が重複していれば409を返す", func(t *testing.T) {
		h := newTestCollectionHandler(&stubCollectionRepo{
			createFn: func(c model.Collection) (model.Collection, error) {
				return model.Collection{}, repository.ErrDuplicateName
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/api/collections", strings.NewReader(`{"name":"Go学習"}`))
		rec := httptest.NewRecorder()
		h.CreateCollection(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, want 409", rec.Code)
		}
	})

	t.Run("存在しない親を指定すると404を返す", func(t *testing.T) {
		h := newTestCollectionHandler(&stubCollectionRepo{
			createFn: func(c model.Collection) (model.Collection, error) {
				return model.Collection{}, repository.ErrParentNotFound
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/api/collections", strings.NewReader(`{"name":"サブスク","parent_id":999}`))
		rec := httptest.NewRecorder()
		h.CreateCollection(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", rec.Code)
		}
	})

	t.Run("repositoryが失敗したら500を返しエラー詳細を露出しない", func(t *testing.T) {
		h := newTestCollectionHandler(&stubCollectionRepo{
			createFn: func(c model.Collection) (model.Collection, error) {
				return model.Collection{}, errors.New("pq: connection refused at 10.0.0.5:5432")
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/api/collections", strings.NewReader(`{"name":"Go学習"}`))
		rec := httptest.NewRecorder()
		h.CreateCollection(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", rec.Code)
		}
		if strings.Contains(rec.Body.String(), "10.0.0.5") {
			t.Errorf("internal error detail leaked to client: %q", rec.Body.String())
		}
	})

	t.Run("成功したら201とCollectionを返す", func(t *testing.T) {
		h := newTestCollectionHandler(&stubCollectionRepo{
			createFn: func(c model.Collection) (model.Collection, error) {
				c.ID = 1
				return c, nil
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/api/collections", strings.NewReader(`{"name":"Go学習"}`))
		rec := httptest.NewRecorder()
		h.CreateCollection(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201", rec.Code)
		}
	})
}

func TestGetCollectionHandler(t *testing.T) {
	t.Run("不正なidなら400を返す", func(t *testing.T) {
		h := newTestCollectionHandler(&stubCollectionRepo{})

		req := httptest.NewRequest(http.MethodGet, "/api/collections/abc", nil)
		req.SetPathValue("id", "abc")
		rec := httptest.NewRecorder()
		h.GetCollection(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})
}

func TestUpdateCollectionHandler(t *testing.T) {
	t.Run("対象が存在しなければ404を返す", func(t *testing.T) {
		h := NewCollectionHandler(service.NewCollectionService(notFoundOnUpdateRepo{}))

		req := httptest.NewRequest(http.MethodPut, "/api/collections/999", strings.NewReader(`{"name":"Go学習"}`))
		req.SetPathValue("id", "999")
		rec := httptest.NewRecorder()
		h.UpdateCollection(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", rec.Code)
		}
	})
}

// notFoundOnUpdateRepo always reports sql.ErrNoRows from Update, for the
// "target collection doesn't exist" handler test above.
type notFoundOnUpdateRepo struct{}

func (notFoundOnUpdateRepo) Create(c model.Collection) (model.Collection, error) {
	return model.Collection{}, nil
}
func (notFoundOnUpdateRepo) List(linkID int) ([]model.Collection, error) { return nil, nil }
func (notFoundOnUpdateRepo) Get(id int) (model.Collection, error)        { return model.Collection{}, nil }
func (notFoundOnUpdateRepo) Update(c model.Collection) (model.Collection, error) {
	return model.Collection{}, sql.ErrNoRows
}
func (notFoundOnUpdateRepo) Delete(id int) error                       { return nil }
func (notFoundOnUpdateRepo) AddLink(collectionID, linkID int) error    { return nil }
func (notFoundOnUpdateRepo) RemoveLink(collectionID, linkID int) error { return nil }
