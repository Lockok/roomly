package report

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRepository struct {
	roomUsage func(ctx context.Context, input RoomUsageInput) ([]RoomUsage, error)
}

func (f fakeRepository) RoomUsage(
	ctx context.Context,
	input RoomUsageInput,
) ([]RoomUsage, error) {
	if f.roomUsage == nil {
		return nil, nil
	}

	return f.roomUsage(ctx, input)
}

func TestServiceRoomUsageRejectsInvalidPeriod(t *testing.T) {
	service := NewService(fakeRepository{
		roomUsage: func(context.Context, RoomUsageInput) ([]RoomUsage, error) {
			t.Fatal("repository must not be called")
			return nil, nil
		},
	})

	_, err := service.RoomUsage(context.Background(), RoomUsageInput{
		From: time.Date(2026, 9, 10, 11, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 9, 10, 10, 0, 0, 0, time.UTC),
	})

	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}
