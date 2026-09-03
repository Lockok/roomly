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

func (input UpdateInput) Validate() error {
	if !input.Name.Set &&
		!input.Location.Set &&
		!input.Floor.Set &&
		!input.Capacity.Set &&
		!input.Equipment.Set &&
		!input.Description.Set &&
		!input.IsActive.Set {
		return ValidationError{Message: "at least one field must be provided"}
	}

	if input.Name.Set {
		if input.Name.Value == nil || strings.TrimSpace(*input.Name.Value) == "" {
			return ValidationError{Message: "name must not be empty"}
		}
	}

	if input.Location.Set {
		if input.Location.Value == nil || strings.TrimSpace(*input.Location.Value) == "" {
			return ValidationError{Message: "location must not be empty"}
		}
	}

	if input.Capacity.Set {
		if input.Capacity.Value == nil || *input.Capacity.Value <= 0 {
			return ValidationError{Message: "capacity must be greater than zero"}
		}
	}

	if input.Equipment.Set {
		if input.Equipment.Value == nil {
			return ValidationError{Message: "equipment must not be null"}
		}

		for _, item := range *input.Equipment.Value {
			if strings.TrimSpace(item) == "" {
				return ValidationError{Message: "equipment items must not be empty"}
			}
		}
	}

	return nil
}
