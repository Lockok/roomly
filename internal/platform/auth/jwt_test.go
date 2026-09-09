package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTServiceGenerateAndParse(t *testing.T) {
	service := NewJWTService("test-secret", time.Hour)
	userID := uuid.New()

	token, err := service.Generate(userID, "employee")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	claims, err := service.Parse(token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}

	if claims.UserID != userID {
		t.Fatalf("expected user ID %s, got %s", userID, claims.UserID)
	}

	if claims.Role != "employee" {
		t.Fatalf("expected role employee, got %s", claims.Role)
	}
}

func TestJWTServiceRejectsInvalidToken(t *testing.T) {
	service := NewJWTService("test-secret", time.Hour)

	_, err := service.Parse("not-a-token")

	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestJWTServiceRejectsExpiredToken(t *testing.T) {
	service := NewJWTService("test-secret", -time.Hour)

	token, err := service.Generate(uuid.New(), "employee")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	_, err = service.Parse(token)

	if !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("expected ErrExpiredToken, got %v", err)
	}
}
