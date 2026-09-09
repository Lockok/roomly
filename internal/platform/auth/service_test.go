package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/Lockok/roomly/internal/user"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type fakeUserReader struct {
	getByEmail func(ctx context.Context, email string) (user.User, error)
}

func (f fakeUserReader) GetByEmail(ctx context.Context, email string) (user.User, error) {
	if f.getByEmail == nil {
		return user.User{}, user.ErrNotFound
	}

	return f.getByEmail(ctx, email)
}

type fakeTokenGenerator struct {
	generate func(userID uuid.UUID, role string) (string, error)
}

func (f fakeTokenGenerator) Generate(userID uuid.UUID, role string) (string, error) {
	if f.generate == nil {
		return "token", nil
	}

	return f.generate(userID, role)
}

func TestServiceLogin(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte("secure-password"),
		bcrypt.MinCost,
	)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	userID := uuid.New()

	tests := []struct {
		name      string
		foundUser user.User
		userErr   error
		password  string
		wantErrIs error
	}{
		{
			name: "successful login",
			foundUser: user.User{
				ID:           userID,
				Email:        "anna@example.com",
				PasswordHash: string(passwordHash),
				Role:         user.RoleEmployee,
				IsActive:     true,
			},
			password: "secure-password",
		},
		{
			name:      "user not found",
			userErr:   user.ErrNotFound,
			password:  "secure-password",
			wantErrIs: ErrInvalidCredentials,
		},
		{
			name: "inactive user",
			foundUser: user.User{
				ID:           userID,
				PasswordHash: string(passwordHash),
				IsActive:     false,
			},
			password:  "secure-password",
			wantErrIs: ErrUserInactive,
		},
		{
			name: "wrong password",
			foundUser: user.User{
				ID:           userID,
				PasswordHash: string(passwordHash),
				IsActive:     true,
			},
			password:  "wrong-password",
			wantErrIs: ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(
				fakeUserReader{
					getByEmail: func(_ context.Context, email string) (user.User, error) {
						if tt.userErr != nil {
							return user.User{}, tt.userErr
						}

						if email != "anna@example.com" {
							t.Fatalf("unexpected email: %q", email)
						}

						return tt.foundUser, nil
					},
				},
				fakeTokenGenerator{
					generate: func(id uuid.UUID, role string) (string, error) {
						if id != userID {
							t.Fatalf("unexpected user ID: %s", id)
						}

						return "jwt-token", nil
					},
				},
			)

			token, err := service.Login(
				context.Background(),
				"  ANNA@EXAMPLE.COM ",
				tt.password,
			)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if token != "jwt-token" {
				t.Fatalf("expected jwt-token, got %q", token)
			}
		})
	}
}
