package discord

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/bwmarrin/discordgo"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/jace-ys/hot-potato-discord/internal/hotpotato"
)

const (
	ctxLogger string = "discord.bot.logger"
)

var (
	commandHandlerPanics = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "command_handler_panics_total",
		Help: "Total number of panics encountered by Discord Application Command handlers.",
	})
)

func init() {
	prometheus.MustRegister(commandHandlerPanics)
}

type Bot struct {
	logger  log.Logger
	server  *http.Server
	discord *discordgo.Session
	command *discordgo.ApplicationCommand

	hotpotato hotpotato.Service
}

func NewBot(logger log.Logger, hotpotato hotpotato.Service, discordToken string, port int) (*Bot, error) {
	session, err := discordgo.New(fmt.Sprintf("Bot %s", discordToken))
	if err != nil {
		return nil, fmt.Errorf("failed to create discord session: %w", err)
	}

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		level.Info(logger).Log("event", "discord.ready", "session", r.SessionID, "guilds", len(r.Guilds))
	})

	if err := session.Open(); err != nil {
		return nil, fmt.Errorf("failed to connect to discord: %w", err)
	}

	bot := &Bot{
		logger:    logger,
		discord:   session,
		hotpotato: hotpotato,
	}

	bot.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: bot.router(),
	}

	return bot, nil
}

func (b *Bot) router() http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/ping", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	return router
}

func (b *Bot) Start(ctx context.Context) error {
	if err := b.handleDiscord(); err != nil {
		return fmt.Errorf("failed to start discord handlers: %w", err)
	}

	level.Info(b.logger).Log("event", "server.started", "name", "bot", "addr", b.server.Addr)
	defer level.Info(b.logger).Log("event", "server.stopped")

	if err := b.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start bot server: %w", err)
	}

	return nil
}

func (b *Bot) handleDiscord() error {
	rootCmd, rootHandler := b.HotPotatoRootCommand()

	cmd, err := b.discord.ApplicationCommandCreate(b.discord.State.User.ID, "", rootCmd)
	if err != nil {
		return err
	}
	b.command = cmd

	b.discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if r := recover(); r != nil {
				commandHandlerPanics.Inc()
				level.Error(b.logger).Log("event", "command.handle.panic", "handler", "root", "err", r, "trace", debug.Stack())
			}
		}()

		rootHandler(s, i)
	})

	return nil
}

func (b *Bot) Stop(ctx context.Context) error {
	if err := b.discord.Close(); err != nil {
		return fmt.Errorf("failed to diconnect from discord: %w", err)
	}
	level.Info(b.logger).Log("event", "discord.disconnected", "session", b.discord.State.SessionID, "guilds", len(b.discord.State.Guilds))

	if err := b.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown bot server: %w", err)
	}

	return nil
}
