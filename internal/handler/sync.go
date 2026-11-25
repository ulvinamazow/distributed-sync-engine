package handler

import (
	"net/http"

	"github.com/ulvinamazow/distributed-data-synchronization-system/internal/sync"
)

type SyncHandler struct {
	scv *sync.Service
}

func NewSyncHandler(svc *sync.Service) *SyncHandler {
	return &SyncHandler{scv: svc}
}

func (h SyncHandler) Pull(w http.ResponseWriter, r *http.Request) {
	relusts := h.scv.PullFromPeers(r.Context())
	writeJSON(w, http.StatusOK, relusts)
}
