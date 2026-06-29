package handler

import (
	"net/http"
	"task-tracker/internal/service"

	"github.com/go-chi/chi/v5"
)

type ActivityHandler struct {
	svc *service.ActivityEventService
}

func NewActivityHandler(svc *service.ActivityEventService) *ActivityHandler {
	return &ActivityHandler{svc: svc}
}

func (h *ActivityHandler) ListByIssue(w http.ResponseWriter, r *http.Request) {
	issueID := chi.URLParam(r, "issueID")
	events, err := h.svc.ListActivityEventsByIssue(r.Context(), issueID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list activity")
		return
	}
	writeJSON(w, http.StatusOK, events)
}
