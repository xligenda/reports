package kit

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// IDMatcher determines whether a registered custom ID matches an incoming one.
// By default, exact string comparison is used.
type IDMatcher func(registered, incoming string) bool

func exactMatch(registered, incoming string) bool {
	return registered == incoming
}

// Router dispatches Discord interactions to registered handlers.
type Router struct {
	session   *discordgo.Session
	commands  []Command
	buttons   []Button
	modals    []Modal
	IDMatcher IDMatcher
	OnError   ErrorHandler
	// OnRecover is called when a handler panics.
	// Set to nil to disable recovery and let panics propagate.
	// Defaults to DefaultRecoverHandler.
	OnRecover RecoverHandler
}

// NewRouter creates a Router with sane defaults:
// exact-match ID comparison, default error handler, and panic recovery enabled.
func NewRouter(session *discordgo.Session) *Router {
	return &Router{
		session:   session,
		IDMatcher: exactMatch,
		OnError:   DefaultErrorHandler,
		OnRecover: DefaultRecoverHandler,
	}
}

func (r *Router) AddCommand(cmd Command) { r.commands = append(r.commands, cmd) }
func (r *Router) AddButton(btn Button)   { r.buttons = append(r.buttons, btn) }
func (r *Router) AddModal(m Modal)       { r.modals = append(r.modals, m) }

func (r *Router) ClearCommands() { r.commands = []Command{} }
func (r *Router) ClearButtons()  { r.buttons = []Button{} }
func (r *Router) ClearModals()   { r.modals = []Modal{} }

// AddEventHandler registers a raw discordgo event handler (e.g. MessageCreate).
func (r *Router) AddEventHandler(h any) {
	r.session.AddHandler(h)
}

// Init attaches the interaction dispatcher to the session.
func (r *Router) Init() {
	r.session.AddHandler(r.dispatch)
}

// dispatch routes all incoming interactions to the appropriate handler.
func (r *Router) dispatch(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()

	recoverWith(r.OnRecover, s, i, func() {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			r.handleCommand(ctx, s, i)
		case discordgo.InteractionApplicationCommandAutocomplete:
			r.handleAutocomplete(ctx, s, i)
		case discordgo.InteractionModalSubmit:
			r.handleModal(ctx, s, i)
		case discordgo.InteractionMessageComponent:
			r.handleButton(ctx, s, i)
		}
	})
}

func (r *Router) handleCommand(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	name := i.ApplicationCommandData().Name

	for _, cmd := range r.commands {
		if cmd.Definition().Name != name {
			continue
		}
		if err := runHooks(ctx, s, i, cmd.Hooks()); err != nil {
			r.OnError(s, i, errHook(err))
			return
		}
		if err := cmd.Handle(ctx, s, i); err != nil {
			r.OnError(s, i, errCommand(name, err))
		}
		return
	}

	r.OnError(s, i, errUnknownCommand(name))
}

func (r *Router) handleAutocomplete(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	name := i.ApplicationCommandData().Name

	for _, cmd := range r.commands {
		if cmd.Definition().Name != name {
			continue
		}
		if err := cmd.Autocomplete(ctx, s, i); err != nil {
			_ = RespondEmptyAutocomplete(s, i)
		}
		return
	}

	_ = RespondEmptyAutocomplete(s, i)
}

func (r *Router) handleModal(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	id := i.ModalSubmitData().CustomID

	for _, m := range r.modals {
		if !r.IDMatcher(m.CustomID(), id) {
			continue
		}
		if err := runHooks(ctx, s, i, m.Hooks()); err != nil {
			r.OnError(s, i, errHook(err))
			return
		}
		if err := m.Handle(ctx, s, i); err != nil {
			r.OnError(s, i, errModal(id, err))
		}
		return
	}

	r.OnError(s, i, errUnknownModal(id))
}

func (r *Router) handleButton(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	id := i.MessageComponentData().CustomID

	for _, btn := range r.buttons {
		if !r.IDMatcher(btn.CustomID(), id) {
			continue
		}
		if err := runHooks(ctx, s, i, btn.Hooks()); err != nil {
			r.OnError(s, i, errHook(err))
			return
		}
		if err := btn.Handle(ctx, s, i); err != nil {
			r.OnError(s, i, errButton(id, err))
		}
		return
	}

	r.OnError(s, i, errUnknownButton(id))
}

func runHooks(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, hooks []Hook) error {
	for _, h := range hooks {
		if err := h.Execute(ctx, s, i); err != nil {
			return err
		}
	}
	return nil
}

// RegisterCommands bulk-registers all commands with Discord.
// It registers global commands once, and guild-specific commands
// for every guild ID provided
func (r *Router) RegisterCommands(guildIDs ...string) error {
	global, guildCmds := r.splitCommands()

	if err := r.bulkOverwrite("", global); err != nil {
		return fmt.Errorf("failed to register global commands: %w", err)
	}

	for _, id := range guildIDs {
		if id == "" {
			continue
		}

		if err := r.bulkOverwrite(id, guildCmds); err != nil {
			return fmt.Errorf("failed to register commands for guild %s: %w", id, err)
		}
	}

	return nil
}

func (r *Router) splitCommands() (global, guild []*discordgo.ApplicationCommand) {
	for _, cmd := range r.commands {
		if cmd.Scope() == ScopeGuild {
			guild = append(guild, cmd.Definition())
		} else {
			// This covers ScopeGlobal or any other default
			global = append(global, cmd.Definition())
		}
	}
	return
}

func (r *Router) bulkOverwrite(guildID string, cmds []*discordgo.ApplicationCommand) error {
	if len(cmds) == 0 {
		return nil
	}
	if r.session.State == nil || r.session.State.User == nil {
		return ErrSessionNotReady
	}
	_, err := r.session.ApplicationCommandBulkOverwrite(r.session.State.User.ID, guildID, cmds)
	return err
}
