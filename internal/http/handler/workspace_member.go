package handler

import (
	"encoding/json"
	"net/http"
	"task-tracker/internal/service"
)

type WorkspaceMemberHandler struct {
	workspaceMemberService *service.WorkspaceMemberService
}

type AddMemberRequest struct {
	WorkspaceID string `json:"workspace_id"`
	UserID      string `json:"user_id"`
	Role        string `json:"role"`
}

type DeleteMemberRequest struct {
	WorkspaceID string `json:"workspace_id"`
	UserID      string `json:"user_id"`
}
type UpdateMemberRoleRequest struct {
	WorkspaceID string `json:"workspace_id"`
	UserID      string `json:"user_id"`
	Role        string `json:"role"`
}

type ListMembersResponse struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

func NewWorkspaceMemberHandler(workspaceMemberService *service.WorkspaceMemberService) *WorkspaceMemberHandler {
	return &WorkspaceMemberHandler{workspaceMemberService: workspaceMemberService}
}

func (h *WorkspaceMemberHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	err := h.workspaceMemberService.AddMember(r.Context(), req.WorkspaceID, req.UserID, req.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *WorkspaceMemberHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	workspaceID := r.URL.Query().Get("workspace_id")
	if workspaceID == "" {
		http.Error(w, "Bad Request: missing workspace_id", http.StatusBadRequest)
		return
	}

	members, err := h.workspaceMemberService.ListMembers(r.Context(), workspaceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(members)
}

func (h *WorkspaceMemberHandler) DeleteMember(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req DeleteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	err := h.workspaceMemberService.RemoveMember(r.Context(), req.WorkspaceID, req.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkspaceMemberHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdateMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	err := h.workspaceMemberService.UpdateMemberRole(r.Context(), req.WorkspaceID, req.UserID, req.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
