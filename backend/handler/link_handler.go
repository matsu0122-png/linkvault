package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/matsu0122-png/linkvault/backend/model"
	"github.com/matsu0122-png/linkvault/backend/service"
)

type LinkHandler struct {
	service *service.LinkService
}

func NewLinkHandler(service *service.LinkService) *LinkHandler {
	return &LinkHandler{service: service}
}

type linkRequest struct {
	URL   string   `json:"url"`
	Title string   `json:"title"`
	Memo  string   `json:"memo"`
	Tags  []string `json:"tags"`
}

func (h *LinkHandler) ListLinks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	tag := r.URL.Query().Get("tag")

	links, err := h.service.ListLinks(query, tag)
	if err != nil {
		http.Error(w, "Failed to fetch links", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, links)
}

func (h *LinkHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
	req, err := decodeLinkRequest(r)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	link, err := h.service.CreateLink(req.URL, req.Title, req.Memo, req.Tags)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusCreated, link)
}

type bulkLinkRequest struct {
	URLs []string `json:"urls"`
	Tags []string `json:"tags"`
}

type bulkLinkFailure struct {
	URL   string `json:"url"`
	Error string `json:"error"`
}

type bulkLinkResponse struct {
	Created []model.Link      `json:"created"`
	Failed  []bulkLinkFailure `json:"failed"`
}

func (h *LinkHandler) BulkCreateLinks(w http.ResponseWriter, r *http.Request) {
	var req bulkLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.URLs) == 0 {
		http.Error(w, "urls is required", http.StatusBadRequest)
		return
	}
	if len(req.URLs) > service.MaxBulkURLs {
		http.Error(w, fmt.Sprintf("too many urls: max %d", service.MaxBulkURLs), http.StatusBadRequest)
		return
	}

	result := h.service.BulkCreateLinks(req.URLs, req.Tags)

	resp := bulkLinkResponse{
		Created: result.Created,
		Failed:  make([]bulkLinkFailure, len(result.Failed)),
	}
	for i, f := range result.Failed {
		resp.Failed[i] = bulkLinkFailure{URL: f.URL, Error: f.Error}
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *LinkHandler) CheckLinks(w http.ResponseWriter, r *http.Request) {
	links, err := h.service.CheckLinks()
	if err != nil {
		http.Error(w, "Failed to check links", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, links)
}

func (h *LinkHandler) UpdateLink(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	req, err := decodeLinkRequest(r)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	link, err := h.service.UpdateLink(id, req.URL, req.Title, req.Memo, req.Tags)
	if errors.Is(err, service.ErrNotFound) {
		http.Error(w, "Link not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, link)
}

func (h *LinkHandler) DeleteLink(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteLink(id)
	if errors.Is(err, service.ErrNotFound) {
		http.Error(w, "Link not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to delete link", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func decodeLinkRequest(r *http.Request) (linkRequest, error) {
	var req linkRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Println("failed to encode response:", err)
	}
}
