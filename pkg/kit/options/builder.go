package options

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type StringOption struct {
	o *discordgo.ApplicationCommandOption
}
type IntOption struct {
	o *discordgo.ApplicationCommandOption
}
type BoolOption struct {
	o *discordgo.ApplicationCommandOption
}
type UserOption struct {
	o *discordgo.ApplicationCommandOption
}
type RoleOption struct {
	o *discordgo.ApplicationCommandOption
}
type ChannelOption struct {
	o *discordgo.ApplicationCommandOption
}
type AttachmentOption struct {
	o *discordgo.ApplicationCommandOption
}
type SubCommand struct {
	o *discordgo.ApplicationCommandOption
}
type SubCommandGroup struct {
	o *discordgo.ApplicationCommandOption
}

func newOpt(t discordgo.ApplicationCommandOptionType, name, desc string) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{Type: t, Name: name, Description: desc}
}

// String creates a string option.
func String(name, desc string) *StringOption {
	return &StringOption{newOpt(discordgo.ApplicationCommandOptionString, name, desc)}
}
func (o *StringOption) Required() *StringOption       { o.o.Required = true; return o }
func (o *StringOption) Autocomplete() *StringOption   { o.o.Autocomplete = true; return o }
func (o *StringOption) MinLength(n int) *StringOption { o.o.MinLength = &n; return o }
func (o *StringOption) MaxLength(n int) *StringOption { o.o.MaxLength = n; return o }
func (o *StringOption) Choices(pairs ...string) *StringOption {
	for i := 0; i+1 < len(pairs); i += 2 {
		o.o.Choices = append(o.o.Choices, &discordgo.ApplicationCommandOptionChoice{
			Name: pairs[i], Value: pairs[i+1],
		})
	}
	return o
}
func (o *StringOption) Build() *discordgo.ApplicationCommandOption { return o.o }

// Int creates an integer option.
func Int(name, desc string) *IntOption {
	return &IntOption{newOpt(discordgo.ApplicationCommandOptionInteger, name, desc)}
}
func (o *IntOption) Required() *IntOption     { o.o.Required = true; return o }
func (o *IntOption) Autocomplete() *IntOption { o.o.Autocomplete = true; return o }
func (o *IntOption) Min(v float64) *IntOption { o.o.MinValue = &v; return o }
func (o *IntOption) Max(v float64) *IntOption { o.o.MaxValue = v; return o }
func (o *IntOption) Choices(pairs ...any) *IntOption {
	for i := 0; i+1 < len(pairs); i += 2 {
		o.o.Choices = append(o.o.Choices, &discordgo.ApplicationCommandOptionChoice{
			Name: fmt.Sprint(pairs[i]), Value: pairs[i+1],
		})
	}
	return o
}
func (o *IntOption) Build() *discordgo.ApplicationCommandOption { return o.o }

// Bool creates a boolean option.
func Bool(name, desc string) *BoolOption {
	return &BoolOption{newOpt(discordgo.ApplicationCommandOptionBoolean, name, desc)}
}
func (o *BoolOption) Required() *BoolOption                      { o.o.Required = true; return o }
func (o *BoolOption) Build() *discordgo.ApplicationCommandOption { return o.o }

// User creates a user option.
func User(name, desc string) *UserOption {
	return &UserOption{newOpt(discordgo.ApplicationCommandOptionUser, name, desc)}
}
func (o *UserOption) Required() *UserOption                      { o.o.Required = true; return o }
func (o *UserOption) Build() *discordgo.ApplicationCommandOption { return o.o }

// Role creates a role option.
func Role(name, desc string) *RoleOption {
	return &RoleOption{newOpt(discordgo.ApplicationCommandOptionRole, name, desc)}
}
func (o *RoleOption) Required() *RoleOption                      { o.o.Required = true; return o }
func (o *RoleOption) Build() *discordgo.ApplicationCommandOption { return o.o }

// Channel creates a channel option.
func Channel(name, desc string) *ChannelOption {
	return &ChannelOption{newOpt(discordgo.ApplicationCommandOptionChannel, name, desc)}
}
func (o *ChannelOption) Required() *ChannelOption { o.o.Required = true; return o }
func (o *ChannelOption) Types(types ...discordgo.ChannelType) *ChannelOption {
	o.o.ChannelTypes = append(o.o.ChannelTypes, types...)
	return o
}
func (o *ChannelOption) Build() *discordgo.ApplicationCommandOption { return o.o }

// Attachment creates an attachment option.
func Attachment(name, desc string) *AttachmentOption {
	return &AttachmentOption{newOpt(discordgo.ApplicationCommandOptionAttachment, name, desc)}
}
func (o *AttachmentOption) Required() *AttachmentOption                { o.o.Required = true; return o }
func (o *AttachmentOption) Build() *discordgo.ApplicationCommandOption { return o.o }

// Buildable is any option builder that can produce an ApplicationCommandOption.
type Buildable interface {
	Build() *discordgo.ApplicationCommandOption
}

// Sub creates a subcommand with the given options.
func Sub(name, desc string, opts ...Buildable) *SubCommand {
	o := newOpt(discordgo.ApplicationCommandOptionSubCommand, name, desc)
	for _, b := range opts {
		o.Options = append(o.Options, b.Build())
	}
	return &SubCommand{o}
}
func (s *SubCommand) Build() *discordgo.ApplicationCommandOption { return s.o }

// Group creates a subcommand group containing subcommands.
func Group(name, desc string, subs ...*SubCommand) *SubCommandGroup {
	o := newOpt(discordgo.ApplicationCommandOptionSubCommandGroup, name, desc)
	for _, s := range subs {
		o.Options = append(o.Options, s.Build())
	}
	return &SubCommandGroup{o}
}
func (g *SubCommandGroup) Build() *discordgo.ApplicationCommandOption { return g.o }

// Options converts builders into a slice ready for ApplicationCommand.Options.
func Options(opts ...Buildable) []*discordgo.ApplicationCommandOption {
	out := make([]*discordgo.ApplicationCommandOption, 0, len(opts))
	for _, b := range opts {
		out = append(out, b.Build())
	}
	return out
}
