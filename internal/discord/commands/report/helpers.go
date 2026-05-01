package report

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
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
		string(buckets.BucketTypeFromFileName(attachment.Filename)),
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
func (c *ReportCommand) deleteReportsConcurrently(ctx context.Context, reports []*structs.Report) int {
	var (
		wg         sync.WaitGroup
		mu         sync.Mutex
		errCounter int
	)

	for _, r := range reports {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			if err := c.reports.Delete(ctx, id); err != nil {
				mu.Lock()
				errCounter++
				mu.Unlock()
			}
		}(r.GetID())
	}

	wg.Wait()

	return len(reports) - errCounter
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

func ptr[T any](s T) *T {
	return &s
}
