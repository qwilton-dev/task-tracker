package handler

import (
	"encoding/json"
	"net/http"
	"task-tracker/internal/domain"
	"task-tracker/internal/http/middleware"
	"task-tracker/internal/service"

	"github.com/go-chi/chi/v5"
)

type InviteHandler struct {
	svc *service.InviteService
}

func NewInviteHandler(svc *service.InviteService) *InviteHandler {
	return &InviteHandler{svc: svc}
}

func (h *InviteHandler) Create(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspaceID")
	userID, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	invite, err := h.svc.CreateInvite(r.Context(), workspaceID, req.Email, req.Role, userID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, invite)
}

func (h *InviteHandler) List(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspaceID")
	invites, err := h.svc.ListInvites(r.Context(), workspaceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, invites)
}

func (h *InviteHandler) Accept(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	userID, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	if err := h.svc.AcceptInvite(r.Context(), token, userID); err != nil {
		switch err {
		case domain.ErrInviteNotFound:
			writeError(w, http.StatusNotFound, "not_found", err.Error())
		case domain.ErrInviteAlreadyAccepted:
			writeError(w, http.StatusConflict, "invite_already_accepted", err.Error())
		case domain.ErrInviteExpired:
			writeError(w, http.StatusBadRequest, "invite_expired", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
