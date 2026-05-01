package report

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/xligenda/reports/internal/discord"
	"github.com/xligenda/reports/pkg/kit"
	"github.com/xligenda/reports/pkg/kit/options"
)

func (c *ReportCommand) HandleDelete(
	ctx context.Context,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	opts options.OptionsMap,
) error {

	channelID := resolveChannelID(opts, i.ChannelID)
	report, err := c.reports.FindByID(ctx, channelID)
	if err != nil {
		return discord.ErrInternal
	}
	if report == nil {
		return discord.ErrNotFound
	}

	err = c.reports.Delete(ctx, report.GetID())
	if err != nil {
		return discord.ErrInternal
	}

	return kit.RespondOrEdit(s, i, &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{
			buildDeleteEmbed(time.Now().Unix()),
		},
	})
}
