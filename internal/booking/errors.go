package booking

import "errors"

var (
	ErrRoomNotFound     = errors.New("room not found")
	ErrRoomInactive     = errors.New("room is not active")
	ErrCapacityExceeded = errors.New("attendees exceed room capacity")
	ErrBookingConflict  = errors.New("room is already booked for this time range")
)

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}
