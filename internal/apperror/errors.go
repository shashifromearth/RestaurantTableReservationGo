package apperror

import "errors"

var (
	ErrNotFound            = errors.New("resource not found")
	ErrInvalidInput        = errors.New("invalid input")
	ErrTableAlreadyBooked  = errors.New("table already booked for this slot")
	ErrCapacityExceeded    = errors.New("guest count exceeds table capacity")
	ErrCancellationDenied  = errors.New("cancellation not allowed within lead time")
	ErrAlreadyCancelled    = errors.New("reservation is already cancelled")
	ErrNoAvailableTable    = errors.New("no available table for the requested party size")
	ErrDuplicateTable      = errors.New("table number already exists")
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func New(code, message string) AppError {
	return AppError{Code: code, Message: message}
}

func From(err error) AppError {
	switch {
	case errors.Is(err, ErrNotFound):
		return New("NOT_FOUND", err.Error())
	case errors.Is(err, ErrInvalidInput):
		return New("INVALID_INPUT", err.Error())
	case errors.Is(err, ErrTableAlreadyBooked):
		return New("TABLE_ALREADY_BOOKED", err.Error())
	case errors.Is(err, ErrCapacityExceeded):
		return New("CAPACITY_EXCEEDED", err.Error())
	case errors.Is(err, ErrCancellationDenied):
		return New("CANCELLATION_DENIED", err.Error())
	case errors.Is(err, ErrAlreadyCancelled):
		return New("ALREADY_CANCELLED", err.Error())
	case errors.Is(err, ErrNoAvailableTable):
		return New("NO_AVAILABLE_TABLE", err.Error())
	case errors.Is(err, ErrDuplicateTable):
		return New("DUPLICATE_TABLE", err.Error())
	default:
		return New("INTERNAL_ERROR", "an unexpected error occurred")
	}
}
