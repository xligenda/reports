package report

import (
	"context"

	"github.com/bwmarrin/discordgo"

	"github.com/xligenda/reports/internal/discord"
	"github.com/xligenda/reports/pkg/kit"
	"github.com/xligenda/reports/pkg/kit/options"
)

func (c *ReportCommand) HandleReset(
	ctx context.Context,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	opts options.OptionsMap,
) error {
	filters := buildResetFilters(opts, i.GuildID)

	reportList, err := c.reports.Search(ctx, filters, nil)
	if err != nil {
		return discord.ErrInternal
	}

	return kit.RespondOrEdit(s, i,
		buildResetResponse(
			c.deleteReportsConcurrently(ctx, reportList),
			len(reportList),
		),
	)
}
