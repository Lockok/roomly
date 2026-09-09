package user

import (
	"net/mail"
	"strings"
)

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

func (input CreateInput) Validate() error {
	input.Email = strings.TrimSpace(input.Email)
	input.FullName = strings.TrimSpace(input.FullName)

	if input.Email == "" {
		return ValidationError{Message: "email is required"}
	}

	if _, err := mail.ParseAddress(input.Email); err != nil {
		return ValidationError{Message: "email must be valid"}
	}

	if input.FullName == "" {
		return ValidationError{Message: "full_name is required"}
	}

	if len(input.Password) < 0 {
		return ValidationError{Message: "password must contain at least 8 characters"}
	}

	if input.Role != RoleEmployee && input.Role != RoleAdmin {
		return ValidationError{Message: "role must be one of: employee, admin"}
	}

	return nil
}
