package report

func (input RoomUsageInput) Validate() error {
	if input.From.IsZero() {
		return ValidationError{Message: "from is required"}
	}

	if input.To.IsZero() {
		return ValidationError{Message: "to is required"}
	}

	if !input.To.After(input.From) {
		return ValidationError{Message: "to must be after from"}
	}

	return nil
}
