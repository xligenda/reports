package app

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"

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
	handler := kit.NewRouter(a.session)
	handler.OnError = func(s *discordgo.Session, i *discordgo.InteractionCreate, err error) {
		log.Printf("Error handling command: %v\n%s", err, debug.Stack())
		kit.DefaultErrorHandler(s, i, err)
	}

	storage, err := minio.NewMinioClient(
		MustEnv("MINIO_ENDPOINT"),
		MustEnv("MINIO_ACCESS_KEY"),
		MustEnv("MINIO_SECRET_KEY"),
		EnvBool("MINIO_USE_SSL"),
	)
	if err != nil {
		return nil, err
	}

	storageCtx := context.Background()
	for _, name := range []minio.BucketType{
		minio.BucketAudio, minio.BucketImage, minio.BucketVideo, minio.BucketOthers,
	} {
		if e, err := storage.BucketExists(storageCtx, string(name)); !e {
			if storage.CreateBucket(storageCtx, string(name), ""); err != nil {
				return nil, err
			}
			fmt.Printf("Created bucket: %s\n", name)
		}

	}

	permsClient, err := perms.NewClient(MustEnv("PERMS_SERVICE"))
	if err != nil {
		return nil, err
	}

	handler.AddCommand(report.NewReportCommand(
		permsClient,
		reports.NewService(a.DB, &hooks.NoOpHooks{}),
		storage,
	))

	a.handler = handler

	return a, nil
}

func (a *App) Run() error {
	if err := a.session.Open(); err != nil {
		return err
	}

	guildID := MustEnv("DISCORD_GUILD_ID")
	a.handler.RegisterCommands(guildID)
	log.Println("bot commands registered")

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
