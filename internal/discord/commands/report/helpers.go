package report

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/xligenda/reports/internal/discord"
	"github.com/xligenda/reports/internal/services/perms"
	"github.com/xligenda/reports/internal/structs"
	"github.com/xligenda/reports/pkg/kit/options"
	buckets "github.com/xligenda/reports/pkg/minio"
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

func isFilled(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
