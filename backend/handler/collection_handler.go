package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/matsu0122-png/linkvault/backend/service"
)

type CollectionHandler struct {
	service *service.CollectionService
}

func NewCollectionHandler(service *service.CollectionService) *CollectionHandler {
	return &CollectionHandler{service: service}
}

type collectionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    *int   `json:"parent_id"`
}

func (h *CollectionHandler) ListCollections(w http.ResponseWriter, r *http.Request) {
	linkID, _ := strconv.Atoi(r.URL.Query().Get("link_id"))

	collections, err := h.service.ListCollections(linkID)
	if err != nil {
		http.Error(w, "Failed to fetch collections", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, collections)
}

func (h *CollectionHandler) CreateCollection(w http.ResponseWriter, r *http.Request) {
	req, err := decodeCollectionRequest(r)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	c, err := h.service.CreateCollection(req.Name, req.Description, req.ParentID)
	if errors.Is(err, service.ErrDuplicateName) {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	if errors.Is(err, service.ErrParentNotFound) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if writeServiceError(w, err) {
		return
	}

	writeJSON(w, http.StatusCreated, c)
}

func (h *CollectionHandler) GetCollection(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	c, err := h.service.GetCollection(id)
	if errors.Is(err, service.ErrCollectionNotFound) {
		http.Error(w, "Collection not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to fetch collection", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, c)
}

func (h *CollectionHandler) UpdateCollection(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	req, err := decodeCollectionRequest(r)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	c, err := h.service.UpdateCollection(id, req.Name, req.Description)
	if errors.Is(err, service.ErrCollectionNotFound) {
		http.Error(w, "Collection not found", http.StatusNotFound)
		return
	}
	if errors.Is(err, service.ErrDuplicateName) {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	if writeServiceError(w, err) {
		return
	}

	writeJSON(w, http.StatusOK, c)
}

func (h *CollectionHandler) DeleteCollection(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteCollection(id)
	if errors.Is(err, service.ErrCollectionNotFound) {
		http.Error(w, "Collection not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to delete collection", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type addLinkRequest struct {
	LinkID int `json:"link_id"`
}

func (h *CollectionHandler) AddLink(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	var req addLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = h.service.AddLink(id, req.LinkID)
	if errors.Is(err, service.ErrCollectionNotFound) {
		http.Error(w, "Collection not found", http.StatusNotFound)
		return
	}
	if errors.Is(err, service.ErrLinkNotFound) {
		http.Error(w, "Link not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to add link", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CollectionHandler) RemoveLink(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	linkID, err := strconv.Atoi(r.PathValue("linkId"))
	if err != nil {
		http.Error(w, "Invalid linkId", http.StatusBadRequest)
		return
	}

	err = h.service.RemoveLink(id, linkID)
	if errors.Is(err, service.ErrCollectionNotFound) {
		http.Error(w, "Link not found in collection", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to remove link", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func decodeCollectionRequest(r *http.Request) (collectionRequest, error) {
	var req collectionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}
