package report

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Lockok/roomly/internal/platform/httputil"
	"github.com/Lockok/roomly/internal/platform/middleware"
)

type RoomUsageResponse struct {
	RoomID          string  `json:"room_id"`
	RoomName        string  `json:"room_name"`
	Location        string  `json:"location"`
	BookingsCount   int     `json:"bookings_count"`
	BookedMinutes   int     `json:"booked_minutes"`
	UtilizationRate float64 `json:"utilization_rate"`
}

type RoomUsageListResponse struct {
	Items []RoomUsageResponse `json:"items"`
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RoomUsage(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	from, err := time.Parse(time.RFC3339, query.Get("from"))
	if err != nil {
		httputil.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_FROM",
			"from must be a valid RFC3339 timestamp",
		)
		return
	}

	to, err := time.Parse(time.RFC3339, query.Get("to"))
	if err != nil {
		httputil.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_TO",
			"to must be a valid RFC3339 timestamp",
		)
		return
	}

	items, err := h.service.RoomUsage(r.Context(), RoomUsageInput{
		From: from,
		To:   to,
	})
	if err != nil {
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

		slog.Error(
			"RoomUsage failed",
			"request_id", middleware.GetRequestID(r.Context()),
			"error", err,
		)

		httputil.WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"internal server error",
		)
		return
	}

	response := make([]RoomUsageResponse, 0, len(items))
	for _, item := range items {
		response = append(response, RoomUsageResponse{
			RoomID:          item.RoomID.String(),
			RoomName:        item.RoomName,
			Location:        item.Location,
			BookingsCount:   item.BookingsCount,
			BookedMinutes:   item.BookedMinutes,
			UtilizationRate: item.UtilizationRate,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, RoomUsageListResponse{
		Items: response,
	})
}
