package report

import (
	"context"

	"github.com/bwmarrin/discordgo"

	"github.com/xligenda/reports/internal/discord"
	"github.com/xligenda/reports/internal/services/perms"
	"github.com/xligenda/reports/pkg/kit"
	"github.com/xligenda/reports/pkg/kit/options"
)

const (
	fieldStackInfo = "commands.report.info/field"
)

func (c *ReportCommand) HandleInfo(
	ctx context.Context,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	opts options.OptionsMap,
) error {
	issuer, err := discord.ResolveUserID(i)
	if err != nil {
		return discord.ErrInternal
	}

	if opts.Has(channel) {
		if err := c.checkPermission(ctx, issuer, perms.ViewReportsExtended, fieldStackInfo); err != nil {
			return err
		}
	}

	channelID := resolveChannelID(opts, i.ChannelID)

	report, err := c.reports.FindByID(ctx, channelID)
	if err != nil {
		return discord.ErrInternal
	}
	if report == nil {
		return discord.ErrNotFound
	}

	buttons, err := c.buildInfoButtons(ctx, issuer, report)
	if err != nil {
		return discord.ErrInternal
	}

	return kit.RespondOrEdit(s, i,
		buildInfoResponse(report, buttons),
	)
}
