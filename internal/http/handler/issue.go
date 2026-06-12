package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"task-tracker/internal/domain"
	"task-tracker/internal/repository"
	"task-tracker/internal/service"

	"github.com/go-chi/chi/v5"
)

type IssueHandler struct {
	svc *service.IssueService
}

func NewIssueHandler(svc *service.IssueService) *IssueHandler {
	return &IssueHandler{svc: svc}
}

type createIssueRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type updateIssueRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	AssigneeID  string `json:"assignee_id"`
}

type moveIssueRequest struct {
	Status   string  `json:"status"`
	Position float64 `json:"position"`
}

func (h *IssueHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	projectID := chi.URLParam(r, "projectID")
	var req createIssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}

	issue, err := h.svc.CreateIssue(r.Context(), projectID, req.Title, req.Description, userID)
	if err != nil {
		switch err {
		case domain.ErrIssueTitleRequired:
			writeError(w, http.StatusBadRequest, "title_required", err.Error())
		case domain.ErrIssueTitleTooLong:
			writeError(w, http.StatusBadRequest, "title_too_long", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", http.StatusText(http.StatusInternalServerError))
		}
		return
	}

	writeJSON(w, http.StatusCreated, issue)
}

func (h *IssueHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	filters := repository.IssueFilters{
		Status:   r.URL.Query().Get("status"),
		Assignee: r.URL.Query().Get("assignee"),
		Q:        r.URL.Query().Get("q"),
	}

	issues, err := h.svc.ListIssues(r.Context(), projectID, filters)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", http.StatusText(http.StatusInternalServerError))
		return
	}

	writeJSON(w, http.StatusOK, issues)
}

func (h *IssueHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "issueID")
	issue, err := h.svc.GetIssue(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "issue not found")
		return
	}
	writeJSON(w, http.StatusOK, issue)
}

func (h *IssueHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "issueID")
	var req updateIssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}

	issue, err := h.svc.UpdateIssue(r.Context(), id, req.Title, req.Description, req.Priority, req.AssigneeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, issue)
}

func (h *IssueHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "issueID")
	if err := h.svc.DeleteIssue(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, "not_found", "issue not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *IssueHandler) Move(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "issueID")
	var req moveIssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}

	posStr := r.URL.Query().Get("position")
	if posStr != "" {
		if p, err := strconv.ParseFloat(posStr, 64); err == nil {
			req.Position = p
		}
	}

	if err := h.svc.MoveIssue(r.Context(), id, req.Status, req.Position); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_move", err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}
