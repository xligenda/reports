package kit

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// EmptyCommand provides default implementations for optional Command methods.
type EmptyCommand struct{}

func (*EmptyCommand) Scope() Scope  { return ScopeGlobal }
func (*EmptyCommand) Hooks() []Hook { return nil }
func (*EmptyCommand) Autocomplete(_ context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return RespondEmptyAutocomplete(s, i)
}

// EmptyButton provides default implementations for optional Button methods.
type EmptyButton struct{}

func (*EmptyButton) Hooks() []Hook { return nil }

// EmptyModal provides default implementations for optional Modal methods.
type EmptyModal struct{}

func (*EmptyModal) Hooks() []Hook { return nil }

// EmptyHook provides a no-op Execute implementation.
type EmptyHook struct{}

func (*EmptyHook) Execute(_ context.Context, _ *discordgo.Session, _ *discordgo.InteractionCreate) error {
	return nil
}
