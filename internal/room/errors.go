package room

import "errors"

var ErrAlreadyExists = errors.New("room already exists")
var ErrNotFound = errors.New("room not found")
