package report

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/xligenda/reports/internal/discord"
	"github.com/xligenda/reports/pkg/kit"
	"github.com/xligenda/reports/pkg/kit/options"
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

	proofLink := opts.String(proofLink)
	if opts.Has(proof) {
		attachment, ok := resolveAttachment(i, opts.String(proof))
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
		isFilled(opts.String(proof)),
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
