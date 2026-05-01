package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/xligenda/reports/pkg/kit"
)

type UptimeCommand struct {
	*kit.EmptyCommand
	recoveries int
	contact    string
	initTime   time.Time
}

func NewUptimeCommand(contact string) *UptimeCommand {
	return &UptimeCommand{
		initTime:   time.Now(),
		recoveries: 0,
		contact:    contact,
	}
}

func (c *UptimeCommand) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "uptime",
		Description: "Информация о состоянии приложения",
		Type:        discordgo.ChatApplicationCommand,
		IntegrationTypes: &[]discordgo.ApplicationIntegrationType{
			discordgo.ApplicationIntegrationUserInstall,
		},
		Contexts: &[]discordgo.InteractionContextType{
			discordgo.InteractionContextGuild,
		},
	}
}

func (c *UptimeCommand) IncrementRecoveries() {
	c.recoveries++
}

func (c *UptimeCommand) Scope() kit.Scope { return kit.ScopeGuild }

func (c *UptimeCommand) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Последний Рестарт: <t:%d:R>\nHB: <t:%d:R>\nВосстановлений: %d\nПо ошибкам писать [сюда](<%s>)", c.initTime.Unix(), s.LastHeartbeatSent.Unix(), c.recoveries, c.contact),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
