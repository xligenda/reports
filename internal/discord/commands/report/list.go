package report

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/xligenda/reports/internal/config"
	"github.com/xligenda/reports/internal/discord"
	"github.com/xligenda/reports/internal/services/perms"
	"github.com/xligenda/reports/internal/structs"
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

func buildListFilters(
	i *discordgo.InteractionCreate,
	opts options.OptionsMap,
	servers Servers,
) ([]repo.Filter, bool, error) {
	var areFiltersUsed bool

	closedOp := repo.IsNull
	if opts.Bool(closed) {
		closedOp = repo.IsNotNull
		areFiltersUsed = true
	}

	filters := []repo.Filter{
		repo.NewFilter("closed_at", closedOp, nil),
	}

	if opts.Bool(guild) {
		filters = append(filters, repo.NewFilter("guild", repo.Equals, i.GuildID))
		areFiltersUsed = true
	}

	if topic := opts.String(topic); topic != "" {
		filters = append(filters, repo.NewFilter("topic", repo.Equals, topic))
		areFiltersUsed = true
	}

	if opts.String(guilds) != "" {

		targets, err := parseServers(opts.String(guilds))
		if err != nil {
			return nil, false, fmt.Errorf("invalid guilds format: %w", err)
		}

		var guildIDs []string
		for _, server := range targets {
			if serverData := servers.FindByID(config.Index(server)); serverData != nil {
				guildIDs = append(guildIDs, serverData.Guild)
			}
		}

		if len(guildIDs) > 0 {
			filters = append(filters, repo.NewFilter("guild", repo.In, guildIDs))
			areFiltersUsed = true
		}
	}

	return filters, areFiltersUsed, nil
}

func buildListEmbed(displayList []*structs.Report, startIndex, endIndex, total int) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf(
			"Список обращений (%d–%d / %d)",
			min(startIndex+1, total), endIndex, total,
		),
		Color: 7419530,
	}

	if total == 0 {
		embed.Description = "Не найдено обращений по этим фильтрам"
		return embed
	}

	for _, rep := range displayList {
		embed.Description += fmt.Sprintf(
			"<#%s> | Тема — \"%s\" | Передавший — <@%s>\n",
			rep.Channel, rep.Topic, rep.Issuer,
		)
	}

	return embed
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

func parseServers(input string) ([]int, error) {
	input = strings.ReplaceAll(input, " ", "")

	if input == "" {
		return []int{}, nil
	}

	parts := strings.Split(input, ",")
	result := make([]int, 0)

	for _, part := range parts {
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s", part)
			}

			start, err := strconv.Atoi(rangeParts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid number in range start: %s", rangeParts[0])
			}

			end, err := strconv.Atoi(rangeParts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid number in range end: %s", rangeParts[1])
			}

			if start > end {
				return nil, fmt.Errorf("invalid range: start %d is greater than end %d", start, end)
			}

			for i := start; i <= end; i++ {
				result = append(result, i)
			}
		} else {
			num, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid number: %s", part)
			}
			result = append(result, num)
		}
	}

	sort.Ints(result)

	unique := make([]int, 0)
	seen := make(map[int]bool)
	for _, num := range result {
		if !seen[num] {
			unique = append(unique, num)
			seen[num] = true
		}
	}

	return unique, nil
}
