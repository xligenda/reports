package kit

import (
	"github.com/bwmarrin/discordgo"
)

// ErrorHandler is called when a handler or hook returns an error.
// Replace Router.OnError to provide custom error handling behavior.
type ErrorHandler func(s *discordgo.Session, i *discordgo.InteractionCreate, err error)

// DefaultErrorHandler responds with a default ephemeral error message.
func DefaultErrorHandler(s *discordgo.Session, i *discordgo.InteractionCreate, err error) {
	_ = RespondOrEdit(s, i, &discordgo.InteractionResponseData{
		Content: "Something went wrong.",
		Flags:   discordgo.MessageFlagsEphemeral,
	})
}

// RespondOrEdit sends a response based on the current interaction state:
//   - Not yet responded → sends a new channel message response
//   - Deferred (loading state, no content) → edits the deferred message
//   - Already responded → sends a followup message
func RespondOrEdit(s *discordgo.Session, i *discordgo.InteractionCreate, data *discordgo.InteractionResponseData) error {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	})
	if err == nil {
		return nil
	}

	switch detectState(s, i) {
	case stateDeferred:
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content:         &data.Content,
			Embeds:          &data.Embeds,
			Components:      &data.Components,
			AllowedMentions: data.AllowedMentions,
		})
		return err
	default: // stateResponded
		_, err = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Content:         data.Content,
			Embeds:          data.Embeds,
			Components:      data.Components,
			AllowedMentions: data.AllowedMentions,
			Flags:           data.Flags,
		})
		return err
	}
}

// Set ephemeral flag to make the eventual message visible only to the invoking user.
func Defer(s *discordgo.Session, i *discordgo.InteractionCreate, ephemeral bool) error {
	var flags discordgo.MessageFlags
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: flags},
	})
}

// RespondEphemeral is a shortcut for sending a simple ephemeral text message.
func RespondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, message string) error {
	return RespondOrEdit(s, i, &discordgo.InteractionResponseData{
		Content: message,
		Flags:   discordgo.MessageFlagsEphemeral,
	})
}

// RespondEmptyAutocomplete sends an empty autocomplete result list.
func RespondEmptyAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: []*discordgo.ApplicationCommandOptionChoice{},
		},
	})
}

type responseState int

const (
	stateFresh     responseState = iota // no response sent yet
	stateDeferred                       // deferred response (loading indicator, no content)
	stateResponded                      // message already sent
)

func detectState(s *discordgo.Session, i *discordgo.InteractionCreate) responseState {
	resp, err := s.InteractionResponse(i.Interaction)
	if err != nil || resp == nil {
		return stateFresh
	}

	// deferred messages have no content, embeds, components
	if resp.Content == "" && len(resp.Embeds) == 0 && len(resp.Components) == 0 {
		return stateDeferred
	}
	return stateResponded
}
