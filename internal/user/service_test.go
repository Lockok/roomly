package user

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type fakeRepository struct {
	create       func(ctx context.Context, input CreateInput) (User, error)
	getByEmail   func(ctx context.Context, email string) (User, error)
	updateStatus func(ctx context.Context, input UpdateStatusInput) (User, error)
}

func (f fakeRepository) Create(ctx context.Context, input CreateInput) (User, error) {
	if f.create == nil {
		return User{}, nil
	}

	return f.create(ctx, input)
}

func (f fakeRepository) List(context.Context, ListFilter) ([]User, error) {
	return nil, nil
}

func (f fakeRepository) GetByID(context.Context, uuid.UUID) (User, error) {
	return User{}, ErrNotFound
}

func (f fakeRepository) GetByEmail(ctx context.Context, email string) (User, error) {
	if f.getByEmail == nil {
		return User{}, ErrNotFound
	}

	return f.getByEmail(ctx, email)
}

func (f fakeRepository) UpdateStatus(
	ctx context.Context,
	input UpdateStatusInput,
) (User, error) {
	if f.updateStatus == nil {
		return User{}, ErrNotFound
	}

	return f.updateStatus(ctx, input)
}

func TestServiceCreateNormalizesInput(t *testing.T) {
	repository := fakeRepository{
		create: func(_ context.Context, input CreateInput) (User, error) {
			if input.PasswordHash == "" {
				t.Fatal("expected password hash")
			}

			if err := bcrypt.CompareHashAndPassword(
				[]byte(input.PasswordHash),
				[]byte("secure-password"),
			); err != nil {
				t.Fatalf("password hash does not match: %v", err)
			}

			return User{ID: uuid.New(), Email: input.Email}, nil
		},
	}

	service := NewService(repository)

	_, err := service.Create(context.Background(), CreateInput{
		Email:    "  USER@EXAMPLE.COM ",
		FullName: "  Jane Doe  ",
		Password: "secure-password",
		Role:     RoleEmployee,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestServiceCreateRejectsInvalidInput(t *testing.T) {
	repository := fakeRepository{
		create: func(_ context.Context, _ CreateInput) (User, error) {
			t.Fatal("repository must not be called")
			return User{}, nil
		},
	}

	service := NewService(repository)

	_, err := service.Create(context.Background(), CreateInput{
		Email:    "invalid",
		FullName: "Jane Doe",
		Password: "secure-password",
		Role:     RoleEmployee,
	})

	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestServiceUpdateStatus(t *testing.T) {
	userID := uuid.New()

	repository := fakeRepository{
		updateStatus: func(_ context.Context, input UpdateStatusInput) (User, error) {
			if input.ID != userID {
				t.Fatalf("expected user ID %s, got %s", userID, input.ID)
			}

			if input.IsActive {
				t.Fatal("expected user to be deactivated")
			}

			return User{
				ID:       userID,
				IsActive: false,
			}, nil
		},
	}

	service := NewService(repository)

	result, err := service.UpdateStatus(context.Background(), UpdateStatusInput{
		ID:       userID,
		IsActive: false,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.IsActive {
		t.Fatal("expected inactive user")
	}
}
