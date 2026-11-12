package record

import "errors"

var (
	ErrInvalidType = errors.New("invalid record type")
	ErrInvalidData = errors.New("invalid data")
	ErrNotFound    = errors.New("record is not found")
)
