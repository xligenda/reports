package report

import (
	"context"
	"io"

	"github.com/bwmarrin/discordgo"
	"github.com/minio/minio-go/v7"
	"github.com/xligenda/reports/internal/discord"
	"github.com/xligenda/reports/internal/discord/hooks"
	"github.com/xligenda/reports/internal/services/perms"
	"github.com/xligenda/reports/internal/structs"
	"github.com/xligenda/reports/pkg/kit"
	"github.com/xligenda/reports/pkg/kit/options"
	"github.com/xligenda/reports/pkg/repo"
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
	channel   optName = "channel"
	closed    optName = "closed"
	page      optName = "page"
)

var (
	topicOptions []string = []string{
		"Обжалование", "Жалоба", "Другое",
	}
)

type PermsProvider interface {
	Check(ctx context.Context, id string, action perms.Permission, stack string) (bool, error)
}

type ReportService interface {
	Create(
		ctx context.Context,
		channelID, guildID, issuerID, topic string,
		createdAt int64,
		note, proof *string,
	) (*structs.Report, error)
	FindByID(ctx context.Context, id string) (*structs.Report, error)
	FindAll(ctx context.Context, opts *repo.QueryOptions) ([]*structs.Report, error)
	Search(ctx context.Context, filters []repo.Filter, opts *repo.QueryOptions) ([]*structs.Report, error)
	FindPaginated(ctx context.Context, filters []repo.Filter, page, pageSize int, orderBy repo.OrderBy) ([]*structs.Report, int64, error)
	Count(ctx context.Context, filters []repo.Filter) (int64, error)
	Exists(ctx context.Context, filters []repo.Filter) (bool, error)
	Delete(ctx context.Context, id string) error
	Close(ctx context.Context, id string, closedByID string, closedAt int64) (*structs.Report, error)
}

type Storage interface {
	PutObject(ctx context.Context, bucket, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	GetObject(ctx context.Context, bucket, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error)
}

type ReportCommand struct {
	*kit.EmptyCommand
	permsProvider PermsProvider
	reports       ReportService
	storage       Storage
	servers       Servers
}

func NewReportCommand(
	permsProvider PermsProvider,
	reportService ReportService,
	storage Storage,
	servers Servers,
) *ReportCommand {
	return &ReportCommand{
		permsProvider: permsProvider,
		reports:       reportService,
		storage:       storage,
		servers:       servers,
	}
}

func (c *ReportCommand) Hooks() []kit.Hook {
	return []kit.Hook{
		hooks.NewPermsHook(c.permsProvider, perms.None, map[string]perms.Permission{
			add:    perms.SaveReports,
			close:  perms.CloseReports,
			delete: perms.DeleteReports,
			info:   perms.ViewReports,
			list:   perms.ViewReports,
			stats:  perms.ViewReportsExtended,
			reset:  perms.DeleteReports,
		}),
	}
}

func (c *ReportCommand) Scope() kit.Scope { return kit.ScopeGuild }

func (*ReportCommand) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "report",
		Description: "Управление передачей обращений",
		Options: []*discordgo.ApplicationCommandOption{
			options.Sub(
				add, "Передать обращение",
				options.String(topic, "Тема").LazyChoices(topicOptions...).Required(),
				options.String(note, "Заметка").MaxLength(800),
				options.Attachment(proof, "Доказательство"),
				options.String(proofLink, "Ссылка на доказательство").MaxLength(800),
			).Build(),
			options.Sub(
				close, "Закрыть обращение",
				options.Channel(channel, "Канал обращения").Types(discordgo.ChannelTypeGuildText),
				options.String(note, "Заметка").MaxLength(800),
			).Build(),
			options.Sub(
				delete, "Удалить обращение",
				options.Channel(channel, "Канал обращения").Types(discordgo.ChannelTypeGuildText),
			).Build(),
			options.Sub(
				info, "Информация об обращении",
				options.Channel(channel, "Канал обращения").Types(discordgo.ChannelTypeGuildText),
			).Build(),
			options.Sub(
				list, "Список обращений",
				options.Bool(guild, "Этот сервер"),
				options.String(guilds, "Список серверов"),
				options.String(topic, "Фильтрация по теме").LazyChoices("Обжалование", "Жалоба", "Другое"),
				options.Int(page, "Номер страницы").Min(1),
			).Build(),
			options.Sub(
				stats, "Статистика обращений",
				options.Bool(guild, "Этот сервер"),
				options.String(guilds, "Список серверов"),
				options.User(user, "Пользователь"),
				options.String(topic, "Тема").LazyChoices("Обжалование", "Жалоба", "Другое"),
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

	opts, sub, err := options.ParseOptions(i)
	if err != nil || len(sub) != 1 {
		return discord.ErrBadRequest
	}

	kit.Defer(s, i, true)
	switch sub[0] {
	case add:
		return c.HandleAdd(ctx, s, i, opts)
	case close:
		return c.HandleClose(ctx, s, i, opts)
	case delete:
		return c.HandleDelete(ctx, s, i, opts)
	case info:
		return c.HandleInfo(ctx, s, i, opts)
	case list:
		return c.HandleList(ctx, s, i, opts)
	case stats:
		return c.HandleStats(ctx, s, i, opts)
	case reset:
		return c.HandleReset(ctx, s, i, opts)
	}

	return discord.ErrNotImplemented
}
