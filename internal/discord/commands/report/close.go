package report

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/xligenda/reports/internal/discord"
	"github.com/xligenda/reports/internal/services/perms"
	"github.com/xligenda/reports/internal/structs"
	"github.com/xligenda/reports/pkg/kit"
	"github.com/xligenda/reports/pkg/kit/options"
)

const (
	fieldStackClose = "commands.report.close/field"
)

func (c *ReportCommand) HandleClose(
	ctx context.Context,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	opts options.OptionsMap,
) error {
	issuer, err := discord.ResolveUserID(i)
	if err != nil {
		return err
	}

	if err := c.checkPermission(ctx, issuer, perms.CloseReports, fieldStackClose); err != nil {
		return err
	}

	channel := resolveChannelID(opts, i.ChannelID)
	report, err := c.reports.FindByID(ctx, channel)
	if err != nil {
		return discord.ErrInternal
	}
	if report == nil {
		return discord.ErrNotFound
	}

	if isAlreadyClosed(report) {
		return discord.ErrImmutable
	}

	closedAt := time.Now().Unix()
	_, err = c.reports.Close(ctx, report.GetID(), issuer, closedAt)
	if err != nil {
		return discord.ErrInternal
	}

	return kit.RespondOrEdit(s, i, &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{
			buildCloseEmbed(closedAt),
		},
	})
}

func isAlreadyClosed(report *structs.Report) bool {
	return report.ClosedAt != nil && *report.ClosedAt != 0
}
