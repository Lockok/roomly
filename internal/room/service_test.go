package room

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeRepository struct {
	create        func(ctx context.Context, input CreateInput) (Room, error)
	list          func(ctx context.Context, list ListFilter) ([]Room, error)
	listAvailable func(ctx context.Context, input AvailabilityInput) ([]Room, error)
	getByID       func(ctx context.Context, id uuid.UUID) (Room, error)
	update        func(ctx context.Context, input UpdateInput) (Room, error)
}

func (f fakeRepository) Create(ctx context.Context, input CreateInput) (Room, error) {
	if f.create == nil {
		return Room{}, nil
	}

	return f.create(ctx, input)
}

func (f fakeRepository) List(ctx context.Context, filter ListFilter) ([]Room, error) {
	if f.list == nil {
		return nil, nil
	}

	return f.list(ctx, filter)
}

func (f fakeRepository) ListAvailable(ctx context.Context, input AvailabilityInput) ([]Room, error) {
	if f.listAvailable == nil {
		return nil, nil
	}

	return f.listAvailable(ctx, input)
}

func (f fakeRepository) GetByID(ctx context.Context, id uuid.UUID) (Room, error) {
	if f.getByID == nil {
		return Room{}, ErrNotFound
	}

	return f.getByID(ctx, id)
}

func (f fakeRepository) Update(ctx context.Context, input UpdateInput) (Room, error) {
	if f.update == nil {
		return Room{}, nil
	}

	return f.update(ctx, input)
}

func TestServiceUpdateDoesNotCallRepositoryForInvalidInput(t *testing.T) {
	repository := fakeRepository{
		update: func(_ context.Context, _ UpdateInput) (Room, error) {
			t.Fatal("repository must not be called for invalid input")
			return Room{}, nil
		},
	}

	service := NewService(repository)

	_, err := service.Update(context.Background(), UpdateInput{})

	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}

func TestServiceListAvailableDoesNotCallRepositoryForInvalidInterval(t *testing.T) {
	repository := fakeRepository{
		listAvailable: func(
			_ context.Context,
			_ AvailabilityInput,
		) ([]Room, error) {
			t.Fatal("repository must not be called for invalid interval")
			return nil, nil
		},
	}

	service := NewService(repository)

	_, err := service.ListAvailable(context.Background(), AvailabilityInput{
		StartsAt: time.Date(2026, 9, 10, 11, 0, 0, 0, time.UTC),
		EndsAt:   time.Date(2026, 9, 10, 10, 0, 0, 0, time.UTC),
	})

	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}
