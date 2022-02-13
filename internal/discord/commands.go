package discord

import (
	"context"
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/log/level"

	"github.com/jace-ys/hot-potato-discord/internal/hotpotato"
)

const namespace = "discord"

type SubCommandEntry func() (*discordgo.ApplicationCommandOption, SubCommandHandler)
type SubCommandHandler = func(context.Context, *discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption) error

func (b *Bot) HotPotatoRootCommand() (*discordgo.ApplicationCommand, func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	cmd := &discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        "hotpotato",
		Description: "Interact with Hot Potato Bot",
	}

	subcommands := []SubCommandEntry{
		b.HotPotatoTossSubCommand,
		b.HotPotatoStealSubCommand,
		b.HotPotatoCookSubCommand,
		b.HotPotatoWhereSubCommand,
		b.HotPotatoLeaderboardSubCommand,
	}

	handlers := make(map[string]SubCommandHandler)
	for _, entry := range subcommands {
		subcommand, handler := entry()
		cmd.Options = append(cmd.Options, subcommand)
		handlers[subcommand.Name] = handler
	}

	return cmd, func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		command := i.ApplicationCommandData()
		if command.Name != cmd.Name {
			return
		}

		subcommand := command.Options[0]
		logger := log.WithSuffix(b.logger, "subcommand", subcommand.Name, "guild", i.GuildID, "channel", i.ChannelID, "interaction", i.Interaction.ID)

		handle, ok := handlers[subcommand.Name]
		if !ok {
			level.Error(logger).Log("event", "subcommand.unknown")
			return
		}

		level.Info(logger).Log("event", "subcommand.received")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err := handle(ctx, s, i, subcommand); err != nil {
			logger := log.With(logger, "source", log.Caller(2))
			level.Error(logger).Log("event", "subcommand.handle.failure", "err", err)
			b.reply(s, i, UnexpectedErrorReply())
		}

		level.Info(logger).Log("event", "subcommand.handle.success")
	}
}

func (b *Bot) HotPotatoTossSubCommand() (*discordgo.ApplicationCommandOption, SubCommandHandler) {
	opt := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "toss",
		Description: "Toss a hot potato to someone!",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to toss hot potato to",
				Required:    true,
			},
		},
	}

	return opt, func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, data *discordgo.ApplicationCommandInteractionDataOption) error {
		if data.Options[0].Type != discordgo.ApplicationCommandOptionUser || data.Options[0].Name != opt.Options[0].Name {
			return nil
		}

		actorUser := i.Interaction.Member.User
		targetUser := data.Options[0].UserValue(s)

		if targetUser.Bot {
			return b.reply(s, i, TossInvalidTargetReply(targetUser.ID))
		}

		rsp, err := b.hotpotato.Toss(ctx, &hotpotato.TossRequest{
			Namespace:    namespace,
			RoomID:       i.GuildID,
			ChannelID:    i.ChannelID,
			ActorUserID:  actorUser.ID,
			TargetUserID: targetUser.ID,
		})
		if err != nil {
			var e *hotpotato.NotHolderError
			switch {
			case errors.As(err, &e):
				return b.reply(s, i, TossNotHolderReply(e.HolderUserID))
			default:
				return fmt.Errorf("failed to handle toss request: %w", err)
			}
		}

		return b.reply(s, i, TossSuccessReply(actorUser.ID, targetUser.ID, rsp))
	}
}

func (b *Bot) HotPotatoStealSubCommand() (*discordgo.ApplicationCommandOption, SubCommandHandler) {
	opt := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "steal",
		Description: "Steal a hot potato from someone!",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to steal hot potato from",
				Required:    true,
			},
		},
	}

	return opt, func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, data *discordgo.ApplicationCommandInteractionDataOption) error {
		if data.Options[0].Type != opt.Options[0].Type || data.Options[0].Name != opt.Options[0].Name {
			return nil
		}

		actorUser := i.Interaction.Member.User
		targetUser := data.Options[0].UserValue(s)

		if targetUser.Bot {
			return b.reply(s, i, StealInvalidTargetReply(targetUser.ID))
		}

		rsp, err := b.hotpotato.Steal(ctx, &hotpotato.StealRequest{
			Namespace:    namespace,
			RoomID:       i.GuildID,
			ChannelID:    i.ChannelID,
			ActorUserID:  actorUser.ID,
			TargetUserID: targetUser.ID,
		})
		if err != nil {
			var e *hotpotato.NotHolderError
			switch {
			case errors.Is(err, hotpotato.ErrNoOngoingGame):
				return b.reply(s, i, NoOngoingGameReply())
			case errors.Is(err, hotpotato.ErrSelfStealUnallowed):
				return b.reply(s, i, StealInvalidTargetReply(targetUser.ID))
			case errors.As(err, &e):
				return b.reply(s, i, StealNotHolderReply(targetUser.ID, e.HolderUserID))
			default:
				return fmt.Errorf("failed to handle steal request: %w", err)
			}
		}

		return b.reply(s, i, StealSuccessReply(actorUser.ID, targetUser.ID, rsp))
	}
}

func (b *Bot) HotPotatoCookSubCommand() (*discordgo.ApplicationCommandOption, SubCommandHandler) {
	opt := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "cook",
		Description: "Cook a hot potato to make it hotter!",
	}

	return opt, func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, data *discordgo.ApplicationCommandInteractionDataOption) error {
		actorUser := i.Interaction.Member.User

		rsp, err := b.hotpotato.Cook(ctx, &hotpotato.CookRequest{
			Namespace:   namespace,
			RoomID:      i.GuildID,
			ChannelID:   i.ChannelID,
			ActorUserID: actorUser.ID,
		})
		if err != nil {
			var e *hotpotato.NotHolderError
			switch {
			case errors.Is(err, hotpotato.ErrNoOngoingGame):
				return b.reply(s, i, NoOngoingGameReply())
			case errors.As(err, &e):
				return b.reply(s, i, CookNotHolderReply(e.HolderUserID))
			default:
				return fmt.Errorf("failed to handle cook request: %w", err)
			}
		}

		return b.reply(s, i, CookSuccessReply(actorUser.ID, rsp))
	}
}

func (b *Bot) HotPotatoWhereSubCommand() (*discordgo.ApplicationCommandOption, SubCommandHandler) {
	opt := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "where",
		Description: "Check who currently holds the hot potato!",
	}

	return opt, func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, data *discordgo.ApplicationCommandInteractionDataOption) error {
		rsp, err := b.hotpotato.GetHolder(ctx, &hotpotato.GetHolderRequest{
			Namespace: namespace,
			RoomID:    i.GuildID,
			ChannelID: i.ChannelID,
		})
		if err != nil {
			switch {
			case errors.Is(err, hotpotato.ErrNoOngoingGame):
				return b.reply(s, i, NoOngoingGameReply())
			default:
				return fmt.Errorf("failed to handle where request: %w", err)
			}
		}

		return b.reply(s, i, WhereSuccessReply(rsp))
	}
}

func (b *Bot) HotPotatoLeaderboardSubCommand() (*discordgo.ApplicationCommandOption, SubCommandHandler) {
	opt := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "leaderboard",
		Description: "View the leaderboard for the most number of deaths by hot potato!",
	}

	return opt, func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, data *discordgo.ApplicationCommandInteractionDataOption) error {
		rsp, err := b.hotpotato.GetLeaderboard(ctx, &hotpotato.GetLeaderboardRequest{
			Namespace: namespace,
			RoomID:    i.GuildID,
			Top:       10,
		})
		if err != nil {
			return fmt.Errorf("failed to handle leaderboard request: %w", err)
		}

		return b.reply(s, i, LeaderboardSuccessReply(rsp))
	}
}
