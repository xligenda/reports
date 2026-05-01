package discord

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/bwmarrin/discordgo"
	"github.com/xligenda/reports/pkg/kit"
)

type InteractionError struct {
	response *discordgo.InteractionResponseData
	err      string
}

func (e *InteractionError) Respond(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return kit.RespondOrEdit(s, i, e.response)
}

func (e *InteractionError) WithError(err string) *InteractionError {
	e.err = err
	return e
}

func (e *InteractionError) WithRespond(resp *discordgo.InteractionResponseData) *InteractionError {
	e.response = resp
	return e
}

func (e InteractionError) Error() string {
	return e.err
}

// Error constructors for common scenarios
func NewError(message string) *InteractionError {
	return &InteractionError{
		response: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
		err: message,
	}
}

var (
	ErrNotImplemented = NewError("This command is not implemented yet")
	ErrBadRequest     = NewError("Invalid request format")
	ErrNotFound       = NewError("Resource not found")
	ErrForbidden      = NewError("You don't have permission to do this")
	ErrInternal       = NewError("An internal server error occurred")
	ErrAlreadyExists  = NewError("Resource already exists")
	ErrImmutable      = NewError("This resource cannot be modified")
	ErrGRPCError      = func(err error) *InteractionError {
		return NewError(fmt.Sprintf("API error: %v", err))
	}
	ErrStorageError = func(err error) *InteractionError {
		return NewError(fmt.Sprintf("Storage error: %v", err))
	}
	ErrInvalidInput = func(field string) *InteractionError {
		return NewError(fmt.Sprintf("Invalid value for %s", field))
	}
)

func ErrorHandler(s *discordgo.Session, i *discordgo.InteractionCreate, e error) {
	err, ok := e.(*InteractionError)
	if !ok {
		log.Printf("Unexpected error type: %T\n%s", e, debug.Stack())
		kit.DefaultErrorHandler(s, i, e)
		return
	}

	content := fmt.Sprintf("Произошла ошибка: %s", err.Error())
	if err == ErrInternal {
		hash := errorHash(e)
		log.Printf("Internal error [%s]: %v\n%s", hash, e, debug.Stack())
		content = fmt.Sprintf("Произошла внутренняя ошибка\n||%s||", hash)
	}

	err.WithRespond(&discordgo.InteractionResponseData{
		Content: content,
		Flags:   discordgo.MessageFlagsEphemeral,
	})

	if respondErr := err.Respond(s, i); respondErr != nil {
		hash := errorHash(respondErr)
		log.Printf("Failed to respond to interaction error [%s]: %v\n%s", hash, respondErr, debug.Stack())
		kit.RespondOrEdit(s, i, &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s\n||%s||", content, hash),
			Flags:   discordgo.MessageFlagsEphemeral,
		})
	}
}
