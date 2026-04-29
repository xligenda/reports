package repo

import (
	"errors"
	"fmt"
)

// Sentinel errors for type-checking with errors.Is.
var (
	ErrNotFound   = errors.New("entity not found")
	ErrNoRows     = errors.New("no rows returned")
	ErrNoFields   = errors.New("no fields to update")
	ErrDatabase   = errors.New("database error")
	ErrValidation = errors.New("validation error")
)

// RepositoryError is the base error type for all repo errors.
type RepositoryError struct {
	// Op is the operation that failed (e.g. "query", "scan row", "commit transaction").
	Op string
	// Err is the underlying cause.
	Err error
}

func (e *RepositoryError) Error() string {
	if e.Op != "" {
		return fmt.Sprintf("repo %s: %v", e.Op, e.Err)
	}
	return fmt.Sprintf("repo: %v", e.Err)
}

func (e *RepositoryError) Unwrap() error { return e.Err }

// DatabaseError wraps a low-level database failure.
type DatabaseError struct {
	RepositoryError
}

func newDatabaseError(op string, err error) *DatabaseError {
	return &DatabaseError{RepositoryError{Op: op, Err: fmt.Errorf("%w: %w", ErrDatabase, err)}}
}

// NotFoundError is returned when a queried entity does not exist.
type NotFoundError struct {
	RepositoryError
}

func newNotFoundError(msg string) *NotFoundError {
	return &NotFoundError{RepositoryError{Err: fmt.Errorf("%w: %s", ErrNotFound, msg)}}
}

// ValidationError is returned when input data fails repository-level validation.
type ValidationError struct {
	RepositoryError
}

func newValidationError(msg string) *ValidationError {
	return &ValidationError{RepositoryError{Err: fmt.Errorf("%w: %s", ErrValidation, msg)}}
}

// IsNotFound reports whether any error in the chain is a NotFoundError.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsDatabaseError reports whether any error in the chain is a DatabaseError.
func IsDatabaseError(err error) bool {
	return errors.Is(err, ErrDatabase)
}

// IsValidationError reports whether any error in the chain is a ValidationError.
func IsValidationError(err error) bool {
	return errors.Is(err, ErrValidation)
}
