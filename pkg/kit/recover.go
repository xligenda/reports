package kit

import (
	"fmt"
	"runtime/debug"

	"github.com/bwmarrin/discordgo"
)

// RecoverHandler is called when a panic is caught during interaction handling.
// stack is the formatted goroutine stack trace.
type RecoverHandler func(s *discordgo.Session, i *discordgo.InteractionCreate, recovered any, stack []byte)

// DefaultRecoverHandler responds with an ephemeral error message and prints the stack trace.
func DefaultRecoverHandler(s *discordgo.Session, i *discordgo.InteractionCreate, recovered any, stack []byte) {
	fmt.Printf("panic recovered: %v\n%s\n", recovered, stack)
	RespondOrEdit(s, i, &discordgo.InteractionResponseData{
		Content: "An unexpected error occurred.",
		Flags:   discordgo.MessageFlagsEphemeral,
	})
}

// recoverWith wraps a handler call with panic recovery.
// If OnRecover is nil, panics are not caught.
func recoverWith(
	handler RecoverHandler,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	fn func(),
) {
	if handler == nil {
		fn()
		return
	}
	defer func() {
		if r := recover(); r != nil {
			handler(s, i, r, debug.Stack())
		}
	}()
	fn()
}
