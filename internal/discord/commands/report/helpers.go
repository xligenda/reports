package report

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/xligenda/reports/internal/config"
	"github.com/xligenda/reports/internal/discord"
	"github.com/xligenda/reports/internal/services/perms"
	"github.com/xligenda/reports/internal/structs"
	"github.com/xligenda/reports/pkg/kit/options"
	buckets "github.com/xligenda/reports/pkg/minio"
	"github.com/xligenda/reports/pkg/repo"
)

// resolveProofLink downloads the attachment from Discord, uploads it to object storage,
// and returns the resulting public URL. If no attachment is provided it returns
// the raw proofLink string from the options unchanged.
func resolveProofLink(
	ctx context.Context,
	storage Storage,
	attachment *discordgo.MessageAttachment,
	fallback string,
) (string, error) {
	if attachment == nil {
		return fallback, nil
	}

	body, err := fetchAttachmentBody(attachment.URL)
	if err != nil {
		return "", err
	}

	reader := bytes.NewReader(body)
	upd, err := storage.PutObject(
		ctx,
		string(buckets.BucketProof),
		fmt.Sprintf("%s/%s", uuid.New().String(), attachment.Filename),
		reader,
		int64(reader.Len()),
		minio.PutObjectOptions{
			ContentType: attachment.ContentType,
		},
	)
	if err != nil {
		return "", err
	}

	return upd.Location, nil
}

func fetchAttachmentBody(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	if _, err = io.Copy(&buf, resp.Body); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// resolveAttachment looks up the attachment by option key from the interaction's
// resolved data. Returns nil (not an error) when the option is absent.
func resolveAttachment(
	i *discordgo.InteractionCreate,
	optValue string,
) (*discordgo.MessageAttachment, bool) {
	attachment, ok := i.ApplicationCommandData().Resolved.Attachments[optValue]
	if !ok {
		return nil, false
	}
	return attachment, true
}

func (c *ReportCommand) checkPermission(ctx context.Context, userID string, perm perms.Permission, stack string) error {
	granted, err := c.permsProvider.Check(ctx, userID, perm, stack)
	if err != nil {
		return err
	}
	if !granted {
		return discord.ErrForbidden
	}
	return nil
}

// todo: move to folder /content
// buildReportEmbed constructs the Discord embed shown after a report is created.
func buildReportEmbed(rep *structs.Report, proofLink string) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Color:       15844367,
		Title:       "Репорт успешно сохранён",
		Description: fmt.Sprintf("Тема: %s", rep.Topic),
	}

	if rep.Note != nil {
		embed.Description += fmt.Sprintf("\nПримечание: %s", *rep.Note)
	}

	embed.Description += fmt.Sprintf("\nВремя передачи: <t:%d:f>", rep.CreatedAt)

	if proofLink != "" {
		embed.Description += fmt.Sprintf("\nДоказательство: %s", proofLink)
	}

	return embed
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

	if opts.Has(guilds) {

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

func resolveChannelID(opts options.OptionsMap, fallback string) string {
	if id := opts.Channel(channel); id != "" {
		return id
	}
	return fallback
}

// todo: move to folder /content
// buildCloseEmbed constructs the confirmation embed shown after closing a report.
func buildCloseEmbed(closedAt int64) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Color:       15844367,
		Title:       "Обращение успешно закрыто",
		Description: fmt.Sprintf("Время закрытия: <t:%d:f>", closedAt),
	}
}

func buildDeleteEmbed(deletedAt int64) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Color:       10038562,
		Title:       "Обращение успешно удалено из БД",
		Description: fmt.Sprintf("Информация об обращении **полностью удалена** из БД, он не будет учитвываться в подсчётах и показываться списках.\nВремя удаления: <t:%d:f>", deletedAt),
	}
}

func buildStatsEmbed(weekly, monthly *ReportStats) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Color:       7419530,
		Title:       "Статистика по обращениям",
		Description: "**Неделя:**\n",
	}

	embed.Description += fmt.Sprintf("Создано: %d\nЗакрыто: %d\n", weekly.Created, weekly.Closed)
	embed.Description += formatTopics(weekly.Topics)

	embed.Description += "\n**Месяц:**\n"
	embed.Description += fmt.Sprintf("Создано: %d\nЗакрыто: %d\n", monthly.Created, monthly.Closed)
	embed.Description += formatTopics(monthly.Topics)

	if len(embed.Description) > 4096 {
		embed.Description = embed.Description[:4090] + "..."
	}

	return embed
}

func buildResetResponse(deletedCount, total int) *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		Content: fmt.Sprintf("Успешно удалено %d/%d обращений", deletedCount, total),
	}
}

// buildInfoButtons checks which actions the issuer is permitted to perform and
// returns the corresponding button components.
func (c *ReportCommand) buildInfoButtons(
	ctx context.Context,
	issuer string,
	report *structs.Report,
) ([]discordgo.MessageComponent, error) {
	var buttons []discordgo.MessageComponent

	canClose, err := c.permsProvider.Check(ctx, issuer, perms.CloseReports, fieldStackInfo)
	if err != nil {
		return nil, err
	}
	if canClose {
		buttons = append(buttons, &discordgo.Button{
			Label:    "Закрыть",
			Style:    discordgo.PrimaryButton,
			CustomID: fmt.Sprintf("report:close:%s", report.Channel),
			Disabled: report.ClosedAt != nil,
		})
	}

	canDelete, err := c.permsProvider.Check(ctx, issuer, perms.DeleteReports, fieldStackInfo)
	if err != nil {
		return nil, err
	}
	if canDelete {
		buttons = append(buttons, &discordgo.Button{
			Label:    "Удалить",
			Style:    discordgo.DangerButton,
			CustomID: fmt.Sprintf("report:delete:%s", report.Channel),
		})
	}

	if report.Proof != nil {
		if _, rawURL := buildProofLink(report.Proof); rawURL != "" {
			buttons = append(buttons, &discordgo.Button{
				Label: "Доказательство",
				Style: discordgo.LinkButton,
				URL:   rawURL,
			})
		}
	}

	return buttons, nil
}

// buildInfoResponse assembles the full InteractionResponseData for the info embed.
func buildInfoResponse(report *structs.Report, buttons []discordgo.MessageComponent) *discordgo.InteractionResponseData {
	proofDisplay := "-"
	if report.Proof != nil {
		proofDisplay, _ = buildProofLink(report.Proof)
	}

	description := fmt.Sprintf(
		"<#%s>\nСоздатель: <@%s>\nТема: %s\nПримечание: %s\nДоказательство: %s\nВремя создания: <t:%d:f>",
		report.Channel,
		report.Issuer,
		report.Topic,
		derefOrDash(report.Note),
		proofDisplay,
		report.CreatedAt,
	)

	if report.ClosedAt != nil {
		description += fmt.Sprintf("\nВремя закрытия: <t:%d:f>", *report.ClosedAt) // ← bug fix: was passing pointer, not value
	}

	data := &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{{
			Color:       7419530,
			Title:       "Информация об обращении",
			Description: description,
		}},
	}

	if len(buttons) > 0 {
		data.Components = []discordgo.MessageComponent{
			discordgo.ActionsRow{Components: buttons},
		}
	}

	return data
}

// buildResetFilters constructs the filter slice from the provided options.
// guild=true  → restrict to the current guild.
// closed=true → only closed reports; false → only open reports.
func buildResetFilters(opts options.OptionsMap, guildID string) []repo.Filter {
	filters := make([]repo.Filter, 0, 2)

	if opts.Bool(guild) {
		filters = append(filters, repo.NewFilter("guild", repo.Equals, guildID))
	}

	if opts.Bool(closed) {
		filters = append(filters, repo.NewFilter("closed_at", repo.IsNotNull, nil))
	} else {
		filters = append(filters, repo.NewFilter("closed_at", repo.IsNull, nil))
	}

	return filters
}

// deleteReportsConcurrently deletes every report in the list in parallel and
// returns the number of successfully deleted reports.
func deleteReportsConcurrently(ctx context.Context, reports ReportService, list []*structs.Report) int {
	var (
		wg         sync.WaitGroup
		mu         sync.Mutex
		errCounter int
	)

	for _, r := range list {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			if err := reports.Delete(ctx, id); err != nil {
				mu.Lock()
				errCounter++
				mu.Unlock()
			}
		}(r.GetID())
	}

	wg.Wait()

	return len(list) - errCounter
}

func collectStats(
	ctx context.Context,
	reports ReportService,
	base []repo.Filter,
	selectedTopic string,
	period StatsPeriod,
) (*ReportStats, error) {
	stats := &ReportStats{Topics: make(map[string]int64)}

	created, err := reports.Count(ctx, append(base,
		repo.NewFilter("created_at", repo.GreaterEquals, period.From),
		repo.NewFilter("created_at", repo.LessEquals, period.To),
	))
	if err != nil {
		return nil, fmt.Errorf("count created: %w", err)
	}
	stats.Created = created

	closed, err := reports.Count(ctx, append(base,
		repo.NewFilter("closed_at", repo.IsNotNull, nil),
		repo.NewFilter("closed_at", repo.GreaterEquals, period.From),
		repo.NewFilter("closed_at", repo.LessEquals, period.To),
	))
	if err != nil {
		return nil, fmt.Errorf("count closed: %w", err)
	}
	stats.Closed = closed

	topics := topicOptions
	if selectedTopic != "" {
		topics = []string{selectedTopic}
	}

	timeCondition := repo.NewRawFilter(
		fmt.Sprintf("(created_at BETWEEN %d AND %d OR (closed_at IS NOT NULL AND closed_at BETWEEN %d AND %d))",
			period.From, period.To,
			period.From, period.To,
		),
	)

	for _, t := range topics {
		count, err := reports.Count(ctx, append(base,
			repo.NewFilter("topic", repo.Equals, t),
			timeCondition,
		))
		if err != nil {
			return nil, fmt.Errorf("count topic %s: %w", t, err)
		}
		if count > 0 {
			stats.Topics[t] = count
		}
	}

	return stats, nil
}

func buildStatsFilters(opts options.OptionsMap, servers Servers, guildID string) ([]repo.Filter, error) {
	filters := []repo.Filter{}

	if opts.Bool(guild) {
		filters = append(filters, repo.NewFilter("guild", repo.Equals, guildID))
	}

	if opts.Has(topic) {
		filters = append(filters, repo.NewFilter("topic", repo.Equals, opts.String(topic)))
	}

	if opts.Has(user) {
		filters = append(filters, repo.NewORFilter(
			repo.NewFilter("issuer", repo.Equals, opts.String(user)),
			repo.NewFilter("closed_by", repo.Equals, opts.String(user)),
		))
	}

	if opts.Has(guilds) {
		targets, err := parseServers(opts.String(guilds))
		if err != nil {
			return nil, fmt.Errorf("invalid guilds format: %w", err)
		}

		var guildIDs []string
		for _, server := range targets {
			if serverData := servers.FindByID(config.Index(server)); serverData != nil {
				guildIDs = append(guildIDs, serverData.Guild)
			}
		}

		if len(guildIDs) > 0 {
			filters = append(filters, repo.NewFilter("guild", repo.In, guildIDs))
		}
	}

	return filters, nil
}

func formatTopics(topics map[string]int64) string {
	if len(topics) == 0 {
		return "Темы: отсутствуют\n"
	}

	keys := make([]string, 0, len(topics))
	for k := range topics {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := "Темы:\n"
	for _, k := range keys {
		out += fmt.Sprintf("*%s: %d*\n", k, topics[k])
	}
	return out
}

// buildProofLink converts a raw proof URL into a markdown display string and
// returns the original URL for use as a button link.
// Returns ("-", "") on nil input, or an error placeholder with no URL on parse failure.
func buildProofLink(proof *string) (display, rawURL string) {
	if proof == nil {
		return "-", ""
	}

	parsedURL, err := url.Parse(*proof)
	if err != nil || parsedURL.Host == "" {
		return "*ошибка сохранения доказательства*", ""
	}

	return fmt.Sprintf("[%s](<%s>)", parsedURL.Host, *proof), *proof
}

func derefOrDash(s *string) string {
	if s == nil {
		return "-"
	}
	return *s
}

func isFilled(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

type StatsPeriod struct {
	From int64
	To   int64
}

func currentWeekPeriod() StatsPeriod {
	now := time.Now().UTC()

	weekday := int(now.Weekday())
	if weekday == 0 { // monday = 1
		weekday = 7 // sunday = 7
	}

	start := time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 7).Add(-time.Second)

	return StatsPeriod{From: start.Unix(), To: end.Unix()}
}

func currentMonthPeriod() StatsPeriod {
	now := time.Now().UTC()

	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0).Add(-time.Second)

	return StatsPeriod{From: start.Unix(), To: end.Unix()}
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
