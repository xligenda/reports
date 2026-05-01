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

func (c *ReportCommand) HandleAdd(
	ctx context.Context,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	opts options.OptionsMap,
) error {
	issuer, err := discord.ResolveUserID(i)
	if err != nil {
		return err
	}

	if e, err := c.reports.Exists(ctx, []repo.Filter{
		{Field: "id", Operator: repo.Equals, Value: i.ChannelID},
	}); err != nil {
		return discord.ErrInternal
	} else if e {
		return discord.ErrAlreadyExists
	}

	proofLink := opts.String(proofLink)
	if opts.Has(proof) {
		attachment, ok := resolveAttachment(i, opts.Attachment(proof))
		if !ok {
			return discord.ErrBadRequest
		}

		proofLink, err = resolveProofLink(ctx, c.storage, attachment, proofLink)
		if err != nil {
			return discord.ErrInternal
		}
	}

	rep, err := c.reports.Create(
		ctx,
		i.ChannelID,
		i.GuildID,
		issuer,
		opts.String(topic),
		time.Now().Unix(),
		isFilled(opts.String(note)),
		isFilled(proofLink),
	)
	if err != nil || rep == nil {
		return discord.ErrInternal
	}

	return kit.RespondOrEdit(s, i, &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{
			buildReportEmbed(rep, proofLink),
		},
	})
}
