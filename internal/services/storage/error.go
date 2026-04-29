package storage

import "fmt"

type Error struct {
	Op  string
	Err error
}

func (e *Error) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("storage: %s operation failed", e.Op)
	}
	return fmt.Sprintf("storage: %s: %v", e.Op, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}
