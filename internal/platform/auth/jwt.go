package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTService(secret string, ttl time.Duration) *JWTService {
	return &JWTService{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (s *JWTService) Generate(userID uuid.UUID, role string) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign JWT: %w", err)
	}

	return signed, nil
}

func (s *JWTService) Parse(tokenString string) (Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}

			return s.secret, nil
		},
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return Claims{}, ErrExpiredToken
		}

		return Claims{}, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid || claims.UserID == uuid.Nil {
		return Claims{}, ErrInvalidToken
	}

	return *claims, nil
}
