package handler

import (
	"encoding/json"
	"net/http"
	"task-tracker/internal/http/middleware"
	"task-tracker/internal/service"

	"github.com/go-chi/chi/v5"
)

type WorkspaceMemberHandler struct {
	workspaceMemberService *service.WorkspaceMemberService
}

func NewWorkspaceMemberHandler(workspaceMemberService *service.WorkspaceMemberService) *WorkspaceMemberHandler {
	return &WorkspaceMemberHandler{workspaceMemberService: workspaceMemberService}
}

func (h *WorkspaceMemberHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspaceID")
	_, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	var req struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	err := h.workspaceMemberService.AddMember(r.Context(), workspaceID, req.UserID, req.Role)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *WorkspaceMemberHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspaceID")
	_, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	members, err := h.workspaceMemberService.ListMembers(r.Context(), workspaceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, members)
}

func (h *WorkspaceMemberHandler) DeleteMember(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspaceID")
	userID := chi.URLParam(r, "memberID")
	_, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	err := h.workspaceMemberService.RemoveMember(r.Context(), workspaceID, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkspaceMemberHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspaceID")
	userID := chi.URLParam(r, "memberID")
	_, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	err := h.workspaceMemberService.UpdateMemberRole(r.Context(), workspaceID, userID, req.Role)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
