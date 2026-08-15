package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/matsu0122-png/linkvault/backend/model"
	"github.com/matsu0122-png/linkvault/backend/service"
)

// stubLinkRepo is a minimal linkRepository stand-in for exercising the
// handler's HTTP status-code mapping. It doesn't aim to cover business
// logic (that's service's job); only the handful of methods a given test
// actually calls need a non-nil func.
type stubLinkRepo struct {
	createFn func(link model.Link) (model.Link, error)
	updateFn func(link model.Link) (model.Link, error)
}

func (s *stubLinkRepo) Create(link model.Link) (model.Link, error) { return s.createFn(link) }
func (s *stubLinkRepo) List(query, tag string) ([]model.Link, error) {
	return nil, nil
}
func (s *stubLinkRepo) ListPage(query, tag string, collectionID int, sort model.SortOption, limit, offset int) ([]model.Link, int, error) {
	return nil, 0, nil
}
func (s *stubLinkRepo) Update(link model.Link) (model.Link, error) { return s.updateFn(link) }
func (s *stubLinkRepo) Delete(id int) error                        { return nil }
func (s *stubLinkRepo) UpdateStatus(id int, status string, checkedAt time.Time) error {
	return nil
}

type stubFetcher struct{}

func (stubFetcher) FetchMetadata(url string) (model.Metadata, error) {
	return model.Metadata{}, nil
}
func (stubFetcher) CheckAlive(url string) (bool, error) { return false, nil }

func newTestLinkHandler(repo *stubLinkRepo) *LinkHandler {
	return NewLinkHandler(service.NewLinkService(repo, stubFetcher{}))
}

func TestCreateLinkHandler(t *testing.T) {
	t.Run("urlが空なら400とバリデーションメッセージを返す", func(t *testing.T) {
		h := newTestLinkHandler(&stubLinkRepo{
			createFn: func(link model.Link) (model.Link, error) {
				t.Fatal("repository should not be called")
				return model.Link{}, nil
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader(`{"url":""}`))
		rec := httptest.NewRecorder()
		h.CreateLink(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "url is required") {
			t.Errorf("expected validation message in body, got %q", rec.Body.String())
		}
	})

	t.Run("repositoryが失敗したら500を返しエラー詳細を露出しない", func(t *testing.T) {
		h := newTestLinkHandler(&stubLinkRepo{
			createFn: func(link model.Link) (model.Link, error) {
				return model.Link{}, errors.New("pq: connection refused at 10.0.0.5:5432")
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader(`{"url":"https://example.com"}`))
		rec := httptest.NewRecorder()
		h.CreateLink(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", rec.Code)
		}
		if strings.Contains(rec.Body.String(), "10.0.0.5") {
			t.Errorf("internal error detail leaked to client: %q", rec.Body.String())
		}
	})

	t.Run("成功したら201とLinkを返す", func(t *testing.T) {
		h := newTestLinkHandler(&stubLinkRepo{
			createFn: func(link model.Link) (model.Link, error) {
				link.ID = 1
				return link, nil
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader(`{"url":"https://example.com"}`))
		rec := httptest.NewRecorder()
		h.CreateLink(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201", rec.Code)
		}

		var got model.Link
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if got.ID != 1 {
			t.Errorf("unexpected link: %+v", got)
		}
	})
}

func TestUpdateLinkHandler(t *testing.T) {
	t.Run("存在しないリンクなら404を返す", func(t *testing.T) {
		h := newTestLinkHandler(&stubLinkRepo{
			updateFn: func(link model.Link) (model.Link, error) {
				return model.Link{}, sql.ErrNoRows
			},
		})

		req := httptest.NewRequest(http.MethodPut, "/api/links/999", strings.NewReader(`{"url":"https://example.com"}`))
		req.SetPathValue("id", "999")
		rec := httptest.NewRecorder()
		h.UpdateLink(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", rec.Code)
		}
	})

	t.Run("urlが空なら400を返す", func(t *testing.T) {
		h := newTestLinkHandler(&stubLinkRepo{
			updateFn: func(link model.Link) (model.Link, error) {
				t.Fatal("repository should not be called")
				return model.Link{}, nil
			},
		})

		req := httptest.NewRequest(http.MethodPut, "/api/links/1", strings.NewReader(`{"url":""}`))
		req.SetPathValue("id", "1")
		rec := httptest.NewRecorder()
		h.UpdateLink(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("不正なidなら400を返す", func(t *testing.T) {
		h := newTestLinkHandler(&stubLinkRepo{})

		req := httptest.NewRequest(http.MethodPut, "/api/links/abc", strings.NewReader(`{"url":"https://example.com"}`))
		req.SetPathValue("id", "abc")
		rec := httptest.NewRecorder()
		h.UpdateLink(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})
}
