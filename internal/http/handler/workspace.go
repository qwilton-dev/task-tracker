package handler

import (
	"encoding/json"
	"net/http"
	"task-tracker/internal/http/middleware"
	"task-tracker/internal/service"

	"github.com/go-chi/chi/v5"
)

type WorkspaceHandler struct {
	svc *service.WorkspaceService
}

func NewWorkspaceHandler(svc *service.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{svc: svc}
}

func (h *WorkspaceHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	ws, err := h.svc.CreateWorkspace(r.Context(), userID, req.Name, req.Slug)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, ws)
}

func (h *WorkspaceHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFrom(r.Context())
	wss, err := h.svc.ListWorkspaces(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, wss)
}

func (h *WorkspaceHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceID")
	ws, err := h.svc.GetWorkspace(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "workspace not found")
		return
	}
	writeJSON(w, http.StatusOK, ws)
}

func (h *WorkspaceHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceID")
	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	ws, err := h.svc.UpdateWorkspace(r.Context(), id, req.Name, req.Slug)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ws)
}

func (h *WorkspaceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceID")
	if err := h.svc.DeleteWorkspace(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, "not_found", "workspace not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
