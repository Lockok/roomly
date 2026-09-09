package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Lockok/roomly/internal/booking"
	"github.com/Lockok/roomly/internal/platform/clock"
	"github.com/Lockok/roomly/internal/platform/config"
	"github.com/Lockok/roomly/internal/platform/db"
	"github.com/Lockok/roomly/internal/platform/health"
	"github.com/Lockok/roomly/internal/platform/middleware"
	"github.com/Lockok/roomly/internal/room"
	"github.com/Lockok/roomly/internal/user"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := db.NewPool(dbCtx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	mux := http.NewServeMux()

	roomRepository := room.NewPostgresRepository(pool)
	roomService := room.NewService(roomRepository)
	roomHandler := room.NewHandler(roomService)

	userRepository := user.NewPostgresRepository(pool)
	userService := user.NewService(userRepository)
	userHandler := user.NewHandler(userService)

	bookingRepository := booking.NewPostgresRepository(pool)
	bookingService := booking.NewService(bookingRepository, roomService, userService, clock.RealClock{})
	bookingHandler := booking.NewHandler(bookingService)

	mux.HandleFunc("GET /health/live", health.Live)
	mux.HandleFunc("GET /health/ready", health.Ready(pool))

	mux.HandleFunc("POST /api/v1/rooms", roomHandler.Create)
	mux.HandleFunc("GET /api/v1/rooms", roomHandler.List)
	mux.HandleFunc("GET /api/v1/rooms/available", roomHandler.ListAvailable)
	mux.HandleFunc("GET /api/v1/rooms/{id}", roomHandler.GetByID)
	mux.HandleFunc("PATCH /api/v1/rooms/{id}", roomHandler.Update)

	mux.HandleFunc("POST /api/v1/users", userHandler.Create)
	mux.HandleFunc("GET /api/v1/users", userHandler.List)

	mux.HandleFunc("POST /api/v1/bookings", bookingHandler.Create)
	mux.HandleFunc("GET /api/v1/bookings", bookingHandler.List)
	mux.HandleFunc("GET /api/v1/bookings/{id}", bookingHandler.GetByID)
	mux.HandleFunc("PATCH /api/v1/bookings/{id}", bookingHandler.Update)
	mux.HandleFunc("POST /api/v1/bookings/{id}/cancel", bookingHandler.Cancel)

	handler := middleware.RequestID(
		middleware.Recovery(mux),
	)

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("API started on http://localhost:%s", cfg.HTTPPort)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
