package discord

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/xligenda/reports/pkg/kit"
)

func RecoverHandler(increment func()) func(s *discordgo.Session, i *discordgo.InteractionCreate, recovered any, stack []byte) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate, recovered any, stack []byte) {
		increment()
		{
			err := fmt.Errorf("panic: %v", recovered)
			hash := errorHash(err)

			log.Printf("Panic recovered [%s]:\n  User: %s\n  Type: %v\n  Panic: %v\n  Stack:\n%s",
				hash,
				i.Member.User.ID,
				i.Type,
				recovered,
				string(stack),
			)

			if respondErr := kit.RespondOrEdit(s, i, &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Произошла внутренняя ошибка\n||%s||", hash),
				Flags:   discordgo.MessageFlagsEphemeral,
			}); respondErr != nil {
				log.Printf("Failed to respond after panic [%s]: %v", hash, respondErr)
			}
		}
	}
}
