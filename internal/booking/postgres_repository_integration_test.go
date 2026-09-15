package booking

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Lockok/roomly/internal/platform/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func integrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	pool, err := db.NewPool(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}

	t.Cleanup(pool.Close)
	return pool
}

func integrationFixtures(t *testing.T, pool *pgxpool.Pool) (uuid.UUID, uuid.UUID) {
	t.Helper()

	ctx := context.Background()
	userID := uuid.New()
	roomID := uuid.New()

	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name)
		VALUES ($1, $2, 'integration-test', 'Integration Test User')
	`, userID, userID.String()+"@example.com")
	if err != nil {
		t.Fatalf("create integration user: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO rooms (id, name, location, capacity)
		VALUES ($1, $2, $3, 10)
	`, roomID, "Integration "+roomID.String(), "Integration")
	if err != nil {
		t.Fatalf("create integration room: %v", err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	})

	return roomID, userID
}

func integrationBookingInput(roomID, userID uuid.UUID, startsAt, endsAt time.Time) CreateInput {
	return CreateInput{
		RoomID:         roomID,
		OrganizerID:    userID,
		Title:          "Integration booking",
		StartsAt:       startsAt,
		EndsAt:         endsAt,
		AttendeesCount: 2,
	}
}

func TestPostgresRepositoryOverlapConstraint(t *testing.T) {
	pool := integrationPool(t)
	roomID, userID := integrationFixtures(t, pool)
	repository := NewPostgresRepository(pool)

	startsAt := time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC)
	first, err := repository.Create(context.Background(), integrationBookingInput(
		roomID,
		userID,
		startsAt,
		startsAt.Add(time.Hour),
	))
	if err != nil {
		t.Fatalf("create first booking: %v", err)
	}

	_, err = repository.Create(context.Background(), integrationBookingInput(
		roomID,
		userID,
		startsAt,
		startsAt.Add(time.Hour),
	))
	if !errors.Is(err, ErrBookingConflict) {
		t.Fatalf("expected ErrBookingConflict, got %v", err)
	}

	_, err = repository.Create(context.Background(), integrationBookingInput(
		roomID,
		userID,
		startsAt.Add(time.Hour),
		startsAt.Add(2*time.Hour),
	))
	if err != nil {
		t.Fatalf("adjacent booking must be allowed: %v", err)
	}

	otherRoomID := uuid.New()
	_, err = pool.Exec(context.Background(), `
		INSERT INTO rooms (id, name, location, capacity)
		VALUES ($1, $2, $3, 10)
	`, otherRoomID, "Integration "+otherRoomID.String(), "Integration")
	if err != nil {
		t.Fatalf("create second integration room: %v", err)
	}

	_, err = repository.Create(context.Background(), integrationBookingInput(
		otherRoomID,
		userID,
		startsAt,
		startsAt.Add(time.Hour),
	))
	if err != nil {
		t.Fatalf("same time in another room must be allowed: %v", err)
	}

	if _, err := repository.Cancel(context.Background(), CancelInput{
		ID:          first.ID,
		CancelledBy: userID,
	}); err != nil {
		t.Fatalf("cancel first booking: %v", err)
	}

	_, err = repository.Create(context.Background(), integrationBookingInput(
		roomID,
		userID,
		startsAt,
		startsAt.Add(time.Hour),
	))
	if err != nil {
		t.Fatalf("cancelled booking must not block a new booking: %v", err)
	}
}

func TestPostgresRepositoryConcurrentOverlap(t *testing.T) {
	pool := integrationPool(t)
	roomID, userID := integrationFixtures(t, pool)
	repository := NewPostgresRepository(pool)

	startsAt := time.Date(2026, 9, 21, 10, 0, 0, 0, time.UTC)
	input := integrationBookingInput(roomID, userID, startsAt, startsAt.Add(time.Hour))

	results := make(chan error, 2)
	var waitGroup sync.WaitGroup
	waitGroup.Add(2)

	for range 2 {
		go func() {
			defer waitGroup.Done()
			_, err := repository.Create(context.Background(), input)
			results <- err
		}()
	}

	waitGroup.Wait()
	close(results)

	var successCount, conflictCount int
	for err := range results {
		switch {
		case err == nil:
			successCount++
		case errors.Is(err, ErrBookingConflict):
			conflictCount++
		default:
			t.Fatalf("unexpected concurrent insert error: %v", err)
		}
	}

	if successCount != 1 || conflictCount != 1 {
		t.Fatalf("expected one success and one conflict, got %d successes and %d conflicts", successCount, conflictCount)
	}
}

func TestPostgresRepositoryNotFoundMapping(t *testing.T) {
	pool := integrationPool(t)
	repository := NewPostgresRepository(pool)

	_, err := repository.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	_, err = repository.Cancel(context.Background(), CancelInput{
		ID:          uuid.New(),
		CancelledBy: uuid.New(),
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound when cancelling missing booking, got %v", err)
	}
}
