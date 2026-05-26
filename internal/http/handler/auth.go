package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"task-tracker/internal/domain"
	"task-tracker/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type userResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json body")
		return
	}

	user, err := h.authService.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEmailRequired):
			writeError(w, http.StatusBadRequest, "email_required", err.Error())
		case errors.Is(err, domain.ErrPasswordRequired):
			writeError(w, http.StatusBadRequest, "password_required", err.Error())
		case errors.Is(err, domain.ErrNameRequired):
			writeError(w, http.StatusBadRequest, "name_required", err.Error())
		case errors.Is(err, domain.ErrEmailAlreadyExists):
			writeError(w, http.StatusConflict, "email_already_exists", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", http.StatusText(http.StatusInternalServerError))
		}
		return
	}

	resp := userResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.UTC().Format(http.TimeFormat),
		UpdatedAt: user.UpdatedAt.UTC().Format(http.TimeFormat),
	}
	writeJSON(w, http.StatusCreated, resp)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorEnvelope{
		Error: errorBody{
			Code:    code,
			Message: message,
		},
	})
}
