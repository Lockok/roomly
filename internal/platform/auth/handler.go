package auth

import (
	"errors"
	"net/http"

	"github.com/Lockok/roomly/internal/platform/httputil"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest

	if err := httputil.DecodeJSON(w, r, &request); err != nil {
		httputil.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"request body must contain valid JSON",
		)
		return
	}

	token, err := h.service.Login(r.Context(), request.Email, request.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			httputil.WriteError(
				w,
				http.StatusUnauthorized,
				"INVALID_CREDENTIALS",
				"invalid email or password",
			)
		case errors.Is(err, ErrUserInactive):
			httputil.WriteError(
				w,
				http.StatusForbidden,
				"USER_INACTIVE",
				"user is inactive",
			)
		default:
			httputil.WriteError(
				w,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"internal server error",
			)
		}
		return
	}

	httputil.WriteJSON(w, http.StatusOK, LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
	})
}
