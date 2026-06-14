package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"task-tracker/internal/domain"
	"task-tracker/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type LabelHandler struct {
	svc *service.LabelService
}

type createLabelRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func NewLabelHandler(svc *service.LabelService) *LabelHandler {
	return &LabelHandler{svc: svc}
}

func (h *LabelHandler) CreateLabel(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspaceID")
	var req createLabelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	label, err := h.svc.CreateLabel(r.Context(), workspaceID, req.Name, req.Color)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrLabelNameRequired):
			writeError(w, http.StatusBadRequest, "label_name_required", err.Error())
		case errors.Is(err, domain.ErrLabelWorkspaceIDRequired):
			writeError(w, http.StatusBadRequest, "label_workspace_required", err.Error())
		case errors.Is(err, domain.ErrLabelColorInvalid):
			writeError(w, http.StatusBadRequest, "label_color_invalid", err.Error())
		default:
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				writeError(w, http.StatusConflict, "label_already_exists", "label with this name already exists")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to create label")
		}
		return
	}

	writeJSON(w, http.StatusCreated, label)
}

func (h *LabelHandler) ListLabels(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspaceID")
	labels, err := h.svc.ListLabels(r.Context(), workspaceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list labels")
		return
	}
	writeJSON(w, http.StatusOK, labels)
}

func (h *LabelHandler) AttachLabel(w http.ResponseWriter, r *http.Request) {
	issueID := chi.URLParam(r, "issueID")
	labelID := chi.URLParam(r, "labelID")
	if err := h.svc.AttachLabel(r.Context(), issueID, labelID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to attach label")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *LabelHandler) DetachLabel(w http.ResponseWriter, r *http.Request) {
	issueID := chi.URLParam(r, "issueID")
	labelID := chi.URLParam(r, "labelID")
	if err := h.svc.DetachLabel(r.Context(), issueID, labelID); err != nil {
		if errors.Is(err, domain.ErrLabelNotFound) {
			writeError(w, http.StatusNotFound, "label_not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to detach label")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
