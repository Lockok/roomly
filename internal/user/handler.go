package user

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Lockok/roomly/internal/platform/httputil"
	"github.com/google/uuid"
)

type Handler struct {
	useCase UseCase
}

func NewHandler(useCase UseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request CreateRequest

	if err := httputil.DecodeJSON(w, r, &request); err != nil {
		httputil.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"request body must contain valid JSON",
		)
		return
	}

	created, err := h.useCase.Create(r.Context(), CreateInput{
		Email:    request.Email,
		Password: request.Password,
		FullName: request.FullName,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrAlreadyExists):
			httputil.WriteError(
				w,
				http.StatusConflict,
				"USER_ALREADY_EXISTS",
				"user with this email already exists",
			)
		default:
			var validationErr ValidationError
			if errors.As(err, &validationErr) {
				httputil.WriteError(
					w,
					http.StatusBadRequest,
					"VALIDATION_ERROR",
					validationErr.Message,
				)
				return
			}

			slog.Error("create user failed", "error", err)

			httputil.WriteError(
				w,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"internal server error",
			)
		}
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, NewUserResponse(created))
}

func (h *Handler) CreateByAdmin(w http.ResponseWriter, r *http.Request) {
	var request CreateByAdminRequest

	if err := httputil.DecodeJSON(w, r, &request); err != nil {
		httputil.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"request body must contain valid JSON",
		)
		return
	}

	created, err := h.useCase.Create(r.Context(), CreateInput{
		Email:    request.Email,
		Password: request.Password,
		FullName: request.FullName,
		Role:     request.Role,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrAlreadyExists):
			httputil.WriteError(
				w,
				http.StatusConflict,
				"USER_ALREADY_EXISTS",
				"user with this email already exists",
			)
		default:
			var validationErr ValidationError
			if errors.As(err, &validationErr) {
				httputil.WriteError(
					w,
					http.StatusBadRequest,
					"VALIDATION_ERROR",
					validationErr.Message,
				)
				return
			}

			slog.Error("createByAdmin failed", "error", err)

			httputil.WriteError(
				w,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"internal server error",
			)
		}
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, NewUserResponse(created))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	var filter ListFilter

	if value := r.URL.Query().Get("active"); value != "" {
		isActive, err := strconv.ParseBool(value)
		if err != nil {
			httputil.WriteError(
				w,
				http.StatusBadRequest,
				"INVALID_ACTIVE_FILTER",
				"active must be true or false",
			)
			return
		}

		filter.IsActive = &isActive
	}

	users, err := h.useCase.List(r.Context(), filter)
	if err != nil {
		slog.Error("list user failed", "error", err)

		httputil.WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"internal server error",
		)
		return
	}

	response := make([]UserResponse, 0, len(users))
	for _, item := range users {
		response = append(response, NewUserResponse(item))
	}

	httputil.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httputil.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_USER_ID",
			"user ID must be a valid UUID",
		)
		return
	}

	var request UpdateStatusRequest

	if err := httputil.DecodeJSON(w, r, &request); err != nil {
		httputil.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"request body must contain valid JSON",
		)
		return
	}

	updated, err := h.useCase.UpdateStatus(r.Context(), UpdateStatusInput{
		ID:       id,
		IsActive: request.IsActive,
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.WriteError(
				w,
				http.StatusNotFound,
				"USER_NOT_FOUND",
				"user not found",
			)
			return
		}

		slog.Error("update status failed", "error", err)

		httputil.WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"internal server error",
		)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, NewUserResponse(updated))
}
