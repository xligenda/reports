package discord

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/xligenda/reports/pkg/kit"
)

// NewRecoverHandler creates a custom panic recovery handler
func NewRecoverHandler(logsChannelID string) kit.RecoverHandler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate, recovered any, stack []byte) {
		// Log the panic to stdout
		log.Printf("🚨 Panic recovered in interaction:\n  User: %s\n  Type: %v\n  Panic: %v\n  Stack:\n%s\n",
			i.User.ID,
			i.Type,
			recovered,
			string(stack),
		)

		// Send user-facing error message
		errMsg := &discordgo.InteractionResponseData{
			Content: "⚠️ An unexpected error occurred while processing your command. Please try again or contact support if the problem persists.",
			Flags:   discordgo.MessageFlagsEphemeral, // Only visible to the user
		}

		if err := kit.RespondOrEdit(s, i, errMsg); err != nil {
			log.Printf("Failed to send error response: %v", err)
		}

		// Send detailed error to logs channel (if configured)
		if logsChannelID != "" {
			logEmbed := &discordgo.MessageEmbed{
				Title:       "🚨 Command Handler Panic",
				Color:       0xFF0000, // Red
				Description: fmt.Sprintf("```\n%v\n```", recovered),
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "User ID",
						Value: i.User.ID,
					},
					{
						Name:  "Interaction Type",
						Value: fmt.Sprintf("%v", i.Type),
					},
					{
						Name:  "Stack Trace",
						Value: fmt.Sprintf("```\n%s\n```", string(stack)),
					},
				},
			}

			if commandData := i.ApplicationCommandData(); commandData.Name != "" {
				logEmbed.Fields = append(logEmbed.Fields, &discordgo.MessageEmbedField{
					Name:  "Command",
					Value: commandData.Name,
				})
			}

			if _, err := s.ChannelMessageSendEmbed(logsChannelID, logEmbed); err != nil {
				log.Printf("Failed to send panic log to logs channel: %v", err)
			}
		}
	}
}
