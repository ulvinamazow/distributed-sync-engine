package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/ulvinamazow/distributed-data-synchronization-system/internal/domain"
	"github.com/ulvinamazow/distributed-data-synchronization-system/internal/dto/request"
	"github.com/ulvinamazow/distributed-data-synchronization-system/internal/dto/response"
	"github.com/ulvinamazow/distributed-data-synchronization-system/internal/repository"
	"github.com/ulvinamazow/distributed-data-synchronization-system/internal/service"
)

type ItemHandler struct {
	srv *service.ItemService
}

func NewItemHandler(svc *service.ItemService) *ItemHandler {
	return &ItemHandler{srv: svc}
}

func toItemResponse(it *domain.Item) response.ItemResponse {
	return response.ItemResponse{
		ID:          it.ID.String(),
		Name:        it.Name,
		Description: it.Description,
		CreatedAt:   it.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:   it.UpdatedAt.Format(time.RFC3339Nano),
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "item not found")
	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func (h *ItemHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req request.CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON not readed")
		return
	}

	item, err := h.srv.Create(r.Context(), req.Name, req.Description)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.Header().Set("Location", "/items/"+item.ID.String())
	writeJSON(w, http.StatusCreated, toItemResponse(item))
}

func (h *ItemHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "incorrect id format")
		return
	}

	item, err := h.srv.Get(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toItemResponse(item))
}

func (h *ItemHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "incorrect id format")
		return
	}

	var req request.CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON not readed")
		return
	}

	item, err := h.srv.Update(r.Context(), id, req.Name, req.Description)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toItemResponse(item))
}

func (h *ItemHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "incorrect id format")
		return
	}

	if err := h.srv.Delete(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ItemHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.srv.List(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := make([]response.ItemResponse, 0, len(items))
	for _, it := range items {
		resp = append(resp, toItemResponse(it))
	}

	writeJSON(w, http.StatusOK, resp)
}
