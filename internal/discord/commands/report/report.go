package report

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/xligenda/reports/internal/discord"
	"github.com/xligenda/reports/pkg/kit"
	"github.com/xligenda/reports/pkg/kit/options"
)

type optName = string

var (
	add    optName = "add"
	close  optName = "close"
	delete optName = "delete"
	info   optName = "info"
	list   optName = "list"
	stats  optName = "stats"
	reset  optName = "reset"

	guild     optName = "guild"
	guilds    optName = "guilds"
	topic     optName = "topic"
	note      optName = "note"
	proof     optName = "proof"
	proofLink optName = "proof_link"
	user      optName = "user"
	channelID optName = "channel_id"
	closed    optName = "closed"
	page      optName = "page"
)

type ReportCommand struct {
	*kit.EmptyCommand
}

func NewReportCommand() *ReportCommand {
	return &ReportCommand{}
}

func (*ReportCommand) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "report",
		Description: "Управление передачей обращений",
		Options: []*discordgo.ApplicationCommandOption{
			options.Sub(
				add, "Передать обращение",
				options.String(topic, "Тема").Choices("Обжалование", "Жалоба", "Другое").Required(),
				options.String(note, "Заметка").MaxLength(800),
				options.Attachment(proof, "Доказательство"),
				options.String(proofLink, "Ссылка на доказательство").MaxLength(800),
			).Build(),
			options.Sub(
				close, "Закрыть обращение",
				options.Channel(channelID, "Канал обращения").Types(discordgo.ChannelTypeGuildText),
				options.String(note, "Заметка").MaxLength(800),
			).Build(),
			options.Sub(
				delete, "Удалить обращение",
				options.Channel(channelID, "Канал обращения").Types(discordgo.ChannelTypeGuildText),
			).Build(),
			options.Sub(
				info, "Информация об обращении",
				options.Channel(channelID, "Канал обращения").Types(discordgo.ChannelTypeGuildText),
			).Build(),
			options.Sub(
				list, "Список обращений",
				options.Bool(guild, "Этот сервер"),
				options.String(guilds, "Список серверов"),
				options.String(topic, "Фильтрация по теме").Choices("Обжалование", "Жалоба", "Другое"),
				options.Int(page, "Номер страницы").Min(1),
			).Build(),
			options.Sub(
				stats, "Статистика обращений",
				options.Bool(guild, "Этот сервер"),
				options.String(guilds, "Список серверов"),
				options.User(user, "Пользователь"),
				options.String(topic, "Тема").Choices("Обжалование", "Жалоба", "Другое"),
			).Build(),
			options.Sub(
				reset, "Сбросить обращения",
				options.Bool(guild, "Этот сервер"),
				options.Bool(closed, "Добавить закрытые"),
			).Build(),
		},
		Type: discordgo.ChatApplicationCommand,
	}
}

func (c *ReportCommand) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {

	_, sub, err := options.ParseOptions(i)
	if err != nil || len(sub) != 1 {
		return discord.ErrBadRequest
	}

	switch sub[0] {
	case add:
	case close:
	case delete:
	case info:
	case list:
	case stats:
	case reset:
	}

	return discord.ErrNotImplemented
}

func (c *ReportCommand) Scope() kit.Scope { return kit.ScopeGuild }
