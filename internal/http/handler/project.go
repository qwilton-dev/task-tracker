package handler

import (
	"encoding/json"
	"net/http"
	"task-tracker/internal/service"

	"github.com/go-chi/chi/v5"
)

type ProjectHandler struct {
	svc *service.ProjectService
}

func NewProjectHandler(svc *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "workspaceSlug")
	var req struct {
		Name string `json:"name"`
		Key  string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	p, err := h.svc.CreateProject(r.Context(), slug, req.Name, req.Key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "workspaceSlug")
	ps, err := h.svc.ListProjects(r.Context(), slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(ps)
}
