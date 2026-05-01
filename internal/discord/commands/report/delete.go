package report

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/xligenda/reports/internal/discord"
	"github.com/xligenda/reports/pkg/kit"
	"github.com/xligenda/reports/pkg/kit/options"
	"github.com/xligenda/reports/pkg/repo"
)

func (c *ReportCommand) HandleDelete(
	ctx context.Context,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	opts options.OptionsMap,
) error {

	channel := resolveChannelID(opts, i.ChannelID)
	exists, err := c.reports.Exists(ctx, []repo.Filter{
		repo.NewFilter("id", repo.Equals, channel),
	})
	if err != nil {
		return discord.ErrInternal
	}
	if !exists {
		return discord.ErrNotFound
	}

	err = c.reports.Delete(ctx, channel)
	if err != nil {
		return discord.ErrInternal
	}

	return kit.RespondOrEdit(s, i, &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{
			buildDeleteEmbed(time.Now().Unix()),
		},
	})
}
