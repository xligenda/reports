package kit

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// Scope defines where a command is registered.
type Scope int

const (
	ScopeGlobal Scope = iota
	ScopeGuild
)

// Hook runs before a handler. Return an error to abort execution.
type Hook interface {
	Execute(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error
}

// Command is a slash command handler.
type Command interface {
	Definition() *discordgo.ApplicationCommand
	Scope() Scope
	Hooks() []Hook
	Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error
	Autocomplete(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error
}

// Button is a message component (button) handler.
type Button interface {
	CustomID() string
	Hooks() []Hook

	Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error
}

// Modal is a modal submit handler.
type Modal interface {
	CustomID() string
	Hooks() []Hook
	Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error
}
