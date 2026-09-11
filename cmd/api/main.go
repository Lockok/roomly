package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Lockok/roomly/internal/booking"
	"github.com/Lockok/roomly/internal/platform/auth"
	"github.com/Lockok/roomly/internal/platform/clock"
	"github.com/Lockok/roomly/internal/platform/config"
	"github.com/Lockok/roomly/internal/platform/db"
	"github.com/Lockok/roomly/internal/platform/health"
	"github.com/Lockok/roomly/internal/platform/logger"
	"github.com/Lockok/roomly/internal/platform/middleware"
	"github.com/Lockok/roomly/internal/report"
	"github.com/Lockok/roomly/internal/room"
	"github.com/Lockok/roomly/internal/user"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load configuration", "error", err)
		os.Exit(1)
	}

	logger := logger.New(cfg.AppEnv)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := db.NewPool(dbCtx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	mux := http.NewServeMux()

	roomRepository := room.NewPostgresRepository(pool)
	roomService := room.NewService(roomRepository)
	roomHandler := room.NewHandler(roomService)

	reportRepository := report.NewPostgresRepository(pool)
	reportService := report.NewService(reportRepository)
	reportHandler := report.NewHandler(reportService)

	userRepository := user.NewPostgresRepository(pool)
	userService := user.NewService(userRepository)
	userHandler := user.NewHandler(userService)

	jwtService := auth.NewJWTService(cfg.JWTSecret, cfg.JWTTTL)
	authService := auth.NewService(userRepository, jwtService)
	authHandler := auth.NewHandler(authService)

	requireAuth := middleware.RequireAuth(jwtService)
	requireAdmin := middleware.RequireRole("admin")

	bookingRepository := booking.NewPostgresRepository(pool)
	bookingService := booking.NewService(bookingRepository, roomService, userService, clock.RealClock{})
	bookingHandler := booking.NewHandler(bookingService)

	mux.HandleFunc("GET /health/live", health.Live)
	mux.HandleFunc("GET /health/ready", health.Ready(pool))

	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	mux.Handle("POST /api/v1/rooms", requireAuth(requireAdmin(http.HandlerFunc(roomHandler.Create))))
	mux.HandleFunc("GET /api/v1/rooms", roomHandler.List)
	mux.HandleFunc("GET /api/v1/rooms/available", roomHandler.ListAvailable)
	mux.Handle("GET /api/v1/rooms/{id}/calendar", requireAuth(http.HandlerFunc(bookingHandler.RoomCalendar)))
	mux.HandleFunc("GET /api/v1/rooms/{id}", roomHandler.GetByID)
	mux.Handle("PATCH /api/v1/rooms/{id}", requireAuth(requireAdmin(http.HandlerFunc(roomHandler.Update))))
	mux.Handle("GET /api/v1/reports/room-usage", requireAuth(requireAdmin(http.HandlerFunc(reportHandler.RoomUsage))))

	mux.HandleFunc("POST /api/v1/users", userHandler.Create)
	mux.Handle("POST /api/v1/admin/users", requireAuth(requireAdmin(http.HandlerFunc(userHandler.CreateByAdmin))))
	mux.HandleFunc("GET /api/v1/users", userHandler.List)
	mux.Handle("PATCH /api/v1/admin/users/{id}/status", requireAuth(requireAdmin(http.HandlerFunc(userHandler.UpdateStatus))))

	mux.Handle("POST /api/v1/bookings", requireAuth(http.HandlerFunc(bookingHandler.Create)))
	mux.Handle("GET /api/v1/bookings", requireAuth(http.HandlerFunc(bookingHandler.List)))
	mux.Handle("GET /api/v1/bookings/{id}", requireAuth(http.HandlerFunc(bookingHandler.GetByID)))
	mux.Handle("PATCH /api/v1/bookings/{id}", requireAuth(http.HandlerFunc(bookingHandler.Update)))
	mux.Handle("POST /api/v1/bookings/{id}/cancel", requireAuth(http.HandlerFunc(bookingHandler.Cancel)))

	handler := middleware.RequestID(
		middleware.AccessLog(
			middleware.Recovery(mux),
		),
	)

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("API started", "address", server.Addr)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}
}
