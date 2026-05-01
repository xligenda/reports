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
	ErrNotImplemented = NewError("Эта команда ещё не реализована")
	ErrBadRequest     = NewError("Указаны неверные аргументы")
	ErrNotFound       = NewError("Ресурс не найден")
	ErrForbidden      = NewError("У вас нет прав для выполнения этого действия")
	ErrInternal       = NewError("Произошла внутренняя ошибка сервера")
	ErrAlreadyExists  = NewError("Ресурс уже существует")
	ErrImmutable      = NewError("Этот ресурс нельзя изменить")
	ErrGRPCError      = func(err error) *InteractionError {
		return NewError(fmt.Sprintf("Ошибка API: %v", err))
	}
	ErrStorageError = func(err error) *InteractionError {
		return NewError(fmt.Sprintf("Ошибка хранилища: %v", err))
	}
	ErrInvalidInput = func(field string) *InteractionError {
		return NewError(fmt.Sprintf("Недопустимое значение для %s", field))
	}
)

func ErrorHandler(s *discordgo.Session, i *discordgo.InteractionCreate, e error) {
	err, ok := e.(*InteractionError)
	if !ok {
		log.Printf("Unexpected error type: %T\n%s\n%s", e, debug.Stack(), e.Error())
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
