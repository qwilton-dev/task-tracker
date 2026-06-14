package handler

import (
	"encoding/json"
	"net/http"
	"task-tracker/internal/http/middleware"
	"task-tracker/internal/service"

	"github.com/go-chi/chi/v5"
)

type CommentHandler struct {
	svc *service.CommentService
}

func NewCommentHandler(svc *service.CommentService) *CommentHandler {
	return &CommentHandler{svc: svc}
}

type createCommentRequest struct {
	Body string `json:"body"`
}

func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	issueID := chi.URLParam(r, "issueID")
	var req createCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}

	comment, err := h.svc.CreateComment(r.Context(), issueID, userID, req.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "body_required", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, comment)
}

func (h *CommentHandler) List(w http.ResponseWriter, r *http.Request) {
	issueID := chi.URLParam(r, "issueID")
	comments, err := h.svc.ListComments(r.Context(), issueID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", http.StatusText(http.StatusInternalServerError))
		return
	}
	writeJSON(w, http.StatusOK, comments)
}
