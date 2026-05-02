package report

import (
	"context"
	"fmt"
	"math"

	"github.com/bwmarrin/discordgo"
	"github.com/xligenda/reports/internal/config"
	"github.com/xligenda/reports/internal/discord"
	"github.com/xligenda/reports/internal/services/perms"
	"github.com/xligenda/reports/pkg/kit"
	"github.com/xligenda/reports/pkg/kit/options"
	"github.com/xligenda/reports/pkg/repo"
)

const (
	pageSize = 25
)

type Servers interface {
	FindByID(id config.Index) *config.Server
}

func (c *ReportCommand) HandleList(
	ctx context.Context,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	opts options.OptionsMap,
) error {
	issuer, err := discord.ResolveUserID(i)
	if err != nil {
		return discord.ErrInternal
	}

	if granted, err := c.permsProvider.Check(ctx, issuer, perms.ViewReportsExtended, fieldStackInfo); !granted || err != nil {
		return discord.ErrForbidden
	}

	filters, areFiltersUsed, err := buildListFilters(i, opts, c.servers)
	if err != nil {
		return discord.ErrBadRequest
	}

	totalItems, err := c.reports.Count(ctx, filters)
	if err != nil {
		return discord.ErrInternal
	}

	totalPages := max(1, int(math.Ceil(float64(totalItems)/pageSize)))
	currentPage := min(int(max(0, opts.Int(page)-1)), totalPages-1)

	queryOpts := repo.NewQueryOptions().
		WithLimit(pageSize).
		WithOffset(currentPage * pageSize)

	reportList, err := c.reports.Search(ctx, filters, queryOpts)
	if err != nil {
		return discord.ErrInternal
	}

	startIndex := currentPage * pageSize
	endIndex := min(startIndex+pageSize, int(totalItems))

	embed := buildListEmbed(reportList, startIndex, endIndex, int(totalItems))
	buttons := buildPageButtons(currentPage, totalPages, areFiltersUsed, opts)

	return kit.RespondOrEdit(s, i, &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{embed},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{Components: buttons},
		},
	})
}

func buildPageButtons(
	currentPage, totalPages int,
	areFiltersUsed bool,
	opts options.OptionsMap,
) []discordgo.MessageComponent {
	guild := opts.Bool(guild)
	closed := opts.Bool(closed)

	makeButton := func(page int) discordgo.Button {
		return discordgo.Button{
			Label: fmt.Sprintf("Страница %d", page+1), // todo: move out to func
			Style: discordgo.SecondaryButton,
			CustomID: fmt.Sprintf(
				"report:page:%d:%t:%t", page, guild, closed,
			),
			Disabled: page == currentPage || areFiltersUsed,
		}
	}

	buttons := []discordgo.MessageComponent{makeButton(0)}

	if totalPages > 1 {
		startPage := max(1, min(currentPage, totalPages-2))
		for p := startPage; p <= min(startPage+2, totalPages-1); p++ {
			buttons = append(buttons, makeButton(p))
		}
		if totalPages > 4 && startPage+2 < totalPages-1 {
			buttons = append(buttons, makeButton(totalPages-1))
		}
	}

	return buttons
}
