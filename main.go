package main

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/log/level"
	"golang.org/x/sync/errgroup"
	"gopkg.in/alecthomas/kingpin.v2"

	_ "github.com/lib/pq"

	"github.com/jace-ys/hot-potato-discord/internal/bedrock"
	"github.com/jace-ys/hot-potato-discord/internal/discord"
	"github.com/jace-ys/hot-potato-discord/internal/game"
	"github.com/jace-ys/hot-potato-discord/internal/hotpotato"
	"github.com/jace-ys/hot-potato-discord/internal/room"
)

var logger log.Logger

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	c := parseCommand()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	db, err := sql.Open("postgres", c.DatabaseURL)
	if err != nil {
		exit(fmt.Errorf("error opening database connection: %w", err))
	}

	rooms := room.NewRepository(db)
	games := game.NewRepository(db)
	gamemaster := hotpotato.NewGameMaster(logger, rooms, games)

	bot, err := discord.NewBot(logger, gamemaster, c.DiscordToken, c.Port)
	if err != nil {
		exit(fmt.Errorf("error initialising bot server: %w", err))
	}

	admin, err := bedrock.NewAdmin(logger, c.AdminPort)
	if err != nil {
		exit(fmt.Errorf("error initialising admin server: %w", err))
	}
	admin.RegisterHealthChecks(bot)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return bot.Start(ctx)
	})
	g.Go(func() error {
		return admin.Start(ctx)
	})
	g.Go(func() error {
		select {
		case <-ctx.Done():
			stop()
			bot.Stop(ctx)
			admin.Stop(ctx)
			return ctx.Err()
		}
	})

	if err := g.Wait(); err != nil {
		exit(err)
	}
}

type config struct {
	Port         int
	AdminPort    int
	DiscordToken string
	DatabaseURL  string
}

func parseCommand() *config {
	var c config

	kingpin.Flag("port", "Target port number for the Hot Potato Bot server.").Envar("PORT").Default("8080").IntVar(&c.Port)
	kingpin.Flag("admin-port", "Target port number for the admin server.").Envar("ADMIN_PORT").Default("9090").IntVar(&c.AdminPort)
	kingpin.Flag("discord-token", "Token for authenticating with Discord.").Envar("DISCORD_TOKEN").Required().StringVar(&c.DiscordToken)
	kingpin.Flag("database-url", "URL for connecting to the Hot Potato Bot database.").Envar("DATABASE_URL").Required().StringVar(&c.DatabaseURL)
	kingpin.Parse()

	return &c
}

func exit(err error) {
	level.Error(logger).Log("event", "app.fatal", "error", err)
	os.Exit(1)
}
