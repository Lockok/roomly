package room

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	create  func(ctx context.Context, input CreateInput) (Room, error)
	list    func(ctx context.Context, list ListFilter) ([]Room, error)
	getByID func(ctx context.Context, id uuid.UUID) (Room, error)
	update  func(ctx context.Context, input UpdateInput) (Room, error)
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
