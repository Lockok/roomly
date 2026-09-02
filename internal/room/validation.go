package room

import (
	"strings"
)

func (input CreateInput) Validate() error {
	if strings.TrimSpace(input.Name) == "" {
		return ValidationError{Message: "name is required"}
	}

	if strings.TrimSpace(input.Location) == "" {
		return ValidationError{Message: "location is required"}
	}

	if input.Capacity <= 0 {
		return ValidationError{Message: "capacity must be greater than zero"}
	}

	for _, item := range input.Equipment {
		if strings.TrimSpace(item) == "" {
			return ValidationError{Message: "equipment items must not be empty"}
		}
	}

	return nil
}
