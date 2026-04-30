package kit

import "errors"

// errors returned by the router.
var (
	ErrUnknownCommand  = errors.New("unknown command")
	ErrUnknownButton   = errors.New("unknown button")
	ErrUnknownModal    = errors.New("unknown modal")
	ErrSessionNotReady = errors.New("session not ready")
	ErrHookAborted     = errors.New("hook aborted")
	ErrNotACommand     = errors.New("not an application command interaction")
	ErrCommandFailed   = errors.New("command failed")
	ErrButtonFailed    = errors.New("button failed")
	ErrModalFailed     = errors.New("modal failed")
)

// commandErr wraps a sentinel with the interaction name for context.
type interactionErr struct {
	sentinel error
	name     string
	cause    error
}

func (e *interactionErr) Error() string {
	if e.cause != nil {
		return e.sentinel.Error() + " " + e.name + ": " + e.cause.Error()
	}
	return e.sentinel.Error() + " " + e.name
}

func (e *interactionErr) Is(target error) bool { return target == e.sentinel }
func (e *interactionErr) Unwrap() error        { return e.cause }

func errUnknownCommand(name string) error {
	return &interactionErr{sentinel: ErrUnknownCommand, name: name}
}

func errUnknownButton(id string) error {
	return &interactionErr{sentinel: ErrUnknownButton, name: id}
}

func errUnknownModal(id string) error {
	return &interactionErr{sentinel: ErrUnknownModal, name: id}
}

func errHook(cause error) error {
	return &interactionErr{sentinel: ErrHookAborted, cause: cause}
}

func errCommand(name string, cause error) error {
	return &interactionErr{sentinel: ErrCommandFailed, name: name, cause: cause}
}

func errButton(id string, cause error) error {
	return &interactionErr{sentinel: ErrButtonFailed, name: id, cause: cause}
}

func errModal(id string, cause error) error {
	return &interactionErr{sentinel: ErrModalFailed, name: id, cause: cause}
}
