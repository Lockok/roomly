package booking

import "errors"

var (
	ErrRoomNotFound     = errors.New("room not found")
	ErrRoomInactive     = errors.New("room is not active")
	ErrCapacityExceeded = errors.New("attendees exceed room capacity")
	ErrBookingConflict  = errors.New("room is already booked for this time range")
	ErrNotFound         = errors.New("booking not found")
	ErrAlreadyCancelled = errors.New("booking is already cancelled")
	ErrBookingStarted   = errors.New("booking has already started or finished")
)

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}
