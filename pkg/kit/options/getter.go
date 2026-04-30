package options

import (
	"github.com/bwmarrin/discordgo"
	kit "github.com/xligenda/reports/pkg/kit"
)

// OptionsMap is a flat name→value map built from interaction options.
// It handles subcommand unwrapping transparently.
type OptionsMap map[string]*discordgo.ApplicationCommandInteractionDataOption

// ParseOptions builds an OptionsMap from interaction data.
// If the interaction targets a subcommand (or subcommand group → subcommand),
// it unwraps automatically and returns the leaf options.
// The active subcommand path is also returned: ["group", "sub"] or ["sub"] or nil.
func ParseOptions(i *discordgo.InteractionCreate) (OptionsMap, []string, error) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return nil, nil, kit.ErrNotACommand
	}
	opts := i.ApplicationCommandData().Options
	return parseOptions(opts)
}

func parseOptions(opts []*discordgo.ApplicationCommandInteractionDataOption) (OptionsMap, []string, error) {
	if len(opts) == 0 {
		return OptionsMap{}, nil, nil
	}

	var path []string

	if opts[0].Type == discordgo.ApplicationCommandOptionSubCommandGroup {
		group := opts[0]
		path = append(path, group.Name)
		opts = group.Options
	}

	if len(opts) > 0 && opts[0].Type == discordgo.ApplicationCommandOptionSubCommand {
		sub := opts[0]
		path = append(path, sub.Name)
		opts = sub.Options
	}

	m := make(OptionsMap, len(opts))
	for _, o := range opts {
		m[o.Name] = o
	}

	return m, path, nil
}

// String returns the string value of the named option, or "" if absent.
func (m OptionsMap) String(name string) string {
	if o, ok := m[name]; ok {
		return o.StringValue()
	}
	return ""
}

// Int returns the int value of the named option, or 0 if absent.
func (m OptionsMap) Int(name string) int64 {
	if o, ok := m[name]; ok {
		return o.IntValue()
	}
	return 0
}

// Bool returns the bool value of the named option, or false if absent.
func (m OptionsMap) Bool(name string) bool {
	if o, ok := m[name]; ok {
		return o.BoolValue()
	}
	return false
}

// User returns the *discordgo.User for the named option, or nil if absent.
func (m OptionsMap) User(name string) *discordgo.User {
	if o, ok := m[name]; ok {
		return o.UserValue(nil)
	}
	return nil
}

// Role returns the role ID string for the named option, or "" if absent.
func (m OptionsMap) Role(name string) string {
	if o, ok := m[name]; ok {
		return o.Value.(string)
	}
	return ""
}

// Channel returns the channel ID string for the named option, or "" if absent.
func (m OptionsMap) Channel(name string) string {
	if o, ok := m[name]; ok {
		return o.Value.(string)
	}
	return ""
}

// Has reports whether the named option is present.
func (m OptionsMap) Has(name string) bool {
	_, ok := m[name]
	return ok
}

// Raw returns the raw option for advanced use, or nil if absent.
func (m OptionsMap) Raw(name string) *discordgo.ApplicationCommandInteractionDataOption {
	return m[name]
}
