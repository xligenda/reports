package report

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/xligenda/reports/internal/discord"
	"github.com/xligenda/reports/pkg/kit"
	"github.com/xligenda/reports/pkg/kit/options"
)

type ReportStats struct {
	Created int64
	Closed  int64
	Topics  map[string]int64
}

func (c *ReportCommand) HandleStats(
	ctx context.Context,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	opts options.OptionsMap,
) error {
	baseFilters, err := buildStatsFilters(opts, c.servers, i.GuildID)
	if err != nil {
		return discord.ErrBadRequest
	}

	week := currentWeekPeriod()
	month := currentMonthPeriod()

	selectedTopic := opts.String(topic)

	weekly, err := collectStats(ctx, c.reports, baseFilters, selectedTopic, week)
	if err != nil {
		return discord.ErrInternal
	}

	monthly, err := collectStats(ctx, c.reports, baseFilters, selectedTopic, month)
	if err != nil {
		return discord.ErrInternal
	}

	return kit.RespondOrEdit(s, i, &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{
			buildStatsEmbed(weekly, monthly),
		},
	})
}
