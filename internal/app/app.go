package app

import (
	"context"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
	"github.com/xligenda/reports/internal/discord/commands/report"
	"github.com/xligenda/reports/internal/services/hooks"
	"github.com/xligenda/reports/internal/services/perms"
	"github.com/xligenda/reports/internal/services/reports"
	"github.com/xligenda/reports/pkg/kit"
	"github.com/xligenda/reports/pkg/minio"
)

type App struct {
	DB         *sqlx.DB
	session    *discordgo.Session
	handler    *kit.Router
	cancelSync context.CancelFunc
}

func New() (*App, error) {
	a := &App{}

	if err := a.InitDB(DBConfigFromEnv()); err != nil {
		return nil, err
	}

	session, err := discordgo.New("Bot " + MustEnv("DISCORD_TOKEN"))
	if err != nil {
		return nil, err
	}
	a.session = session
	a.handler = kit.NewRouter(a.session)

	storage, err := minio.NewMinioClient(
		MustEnv("MINIO_ENDPOINT"),
		MustEnv("MINIO_ACCESS_KEY"),
		MustEnv("MINIO_SECRET_KEY"),
		EnvBool("MINIO_USE_SSL"),
	)
	if err != nil {
		return nil, err
	}

	guildID := MustEnv("DISCORD_GUILD_ID")
	handler := kit.NewRouter(a.session)

	permsClient, err := perms.NewClient(MustEnv("PERMS_SERVICE"))
	if err != nil {
		return nil, err
	}

	handler.AddCommand(report.NewReportCommand(
		permsClient,
		reports.NewService(a.DB, &hooks.NoOpHooks{}),
		storage,
	))

	handler.RegisterCommands(guildID)

	return a, nil
}

func (a *App) Run() error {
	if err := a.session.Open(); err != nil {
		return err
	}

	a.handler.Init()

	log.Println("Bot is running!")

	return nil
}

func (a *App) Shutdown() {
	log.Println("Shutting down...")
	if a.session != nil {
		a.session.Close()
	}
	a.CloseDB()
}
