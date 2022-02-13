package hotpotato

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/log/level"

	"github.com/jace-ys/hot-potato-discord/internal/game"
	"github.com/jace-ys/hot-potato-discord/internal/room"
)

type GameMaster struct {
	logger   log.Logger
	rooms    room.RoomRepository
	games    game.GameRepository
	potatoes []Potato
}

func NewGameMaster(logger log.Logger, rooms room.RoomRepository, games game.GameRepository) *GameMaster {
	return &GameMaster{
		logger: logger,
		rooms:  rooms,
		games:  games,
		potatoes: []Potato{
			RawPotato{},
			BakedPotato{},
			HotPotato{},
			BurntPotato{},
		},
	}
}

func (gm *GameMaster) Toss(ctx context.Context, req *TossRequest) (*TossResponse, error) {
	logger := log.WithSuffix(gm.logger, "namespace", req.Namespace, "room", req.RoomID, "channel", req.ChannelID)

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	r, err := gm.rooms.GetRoom(ctx, string(req.Namespace), req.RoomID)
	if err != nil {
		if !errors.Is(err, room.ErrRoomNotFound) {
			return nil, fmt.Errorf("error getting room: %w", err)
		}

		r, err = gm.rooms.CreateRoom(ctx, string(req.Namespace), req.RoomID)
		if err != nil {
			return nil, fmt.Errorf("error creating room: %w", err)
		}
		level.Info(logger).Log("event", "room.created")
	}

	g, err := gm.games.GetGame(ctx, r.Namespace, req.ChannelID)
	if err != nil && !errors.Is(err, game.ErrGameNotFound) {
		return nil, fmt.Errorf("error getting game: %w", err)
	}

	if errors.Is(err, game.ErrGameNotFound) || g.Finished {
		potato := gm.RandomPotato()
		g, err = gm.games.CreateNewGame(ctx, r.Namespace, r.ID, req.ChannelID, potato.Kind(), req.ActorUserID)
		if err != nil {
			return nil, fmt.Errorf("error creating game: %w", err)
		}
		level.Info(logger).Log("event", "game.created")
	}

	if req.ActorUserID != g.HolderUserID {
		return nil, &NotHolderError{g.HolderUserID}
	}

	potato, err := gm.GetPotato(g.PotatoKind)
	if err != nil {
		return nil, fmt.Errorf("error getting potato of kind '%s': %w", g.PotatoKind, err)
	}

	g, err = gm.games.NextTurn(ctx, r.Namespace, req.ChannelID, req.TargetUserID)
	if err != nil {
		return nil, fmt.Errorf("error handling turn: %w", err)
	}
	level.Info(logger).Log("event", "turn.handled")

	explode := gm.DecideExplode(potato, g.Turns, g.HeatLevel)
	if explode {
		g, err = gm.games.EndGame(ctx, r.Namespace, req.ChannelID)
		if err != nil {
			return nil, fmt.Errorf("error ending game: %w", err)
		}
		level.Info(logger).Log("event", "game.ended")

		err = gm.rooms.IncrementDeaths(ctx, r.Namespace, r.ID, req.TargetUserID)
		if err != nil {
			return nil, fmt.Errorf("error incrementing death count: %w", err)
		}
	}

	return &TossResponse{
		Turn:         g.Turns,
		Potato:       potato,
		HolderUserID: g.HolderUserID,
		Exploded:     explode,
	}, nil
}

func (gm *GameMaster) Steal(ctx context.Context, req *StealRequest) (*StealResponse, error) {
	logger := log.WithSuffix(gm.logger, "namespace", req.Namespace, "room", req.RoomID, "channel", req.ChannelID)

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	r, err := gm.rooms.GetRoom(ctx, string(req.Namespace), req.RoomID)
	if err != nil {
		if !errors.Is(err, room.ErrRoomNotFound) {
			return nil, fmt.Errorf("error getting room: %w", err)
		}

		r, err = gm.rooms.CreateRoom(ctx, string(req.Namespace), req.RoomID)
		if err != nil {
			return nil, fmt.Errorf("error creating room: %w", err)
		}
		level.Info(logger).Log("event", "room.created")
	}

	g, err := gm.games.GetGame(ctx, r.Namespace, req.ChannelID)
	if err != nil && !errors.Is(err, game.ErrGameNotFound) {
		return nil, fmt.Errorf("error getting game: %w", err)
	}

	if errors.Is(err, game.ErrGameNotFound) || g.Finished {
		return nil, ErrNoOngoingGame
	}

	if req.TargetUserID == req.ActorUserID {
		return nil, ErrSelfStealUnallowed
	}

	if req.TargetUserID != g.HolderUserID {
		return nil, &NotHolderError{g.HolderUserID}
	}

	potato, err := gm.GetPotato(g.PotatoKind)
	if err != nil {
		return nil, fmt.Errorf("error getting potato of kind '%s': %w", g.PotatoKind, err)
	}

	g, err = gm.games.NextTurn(ctx, r.Namespace, req.ChannelID, req.ActorUserID)
	if err != nil {
		return nil, fmt.Errorf("error making turn: %w", err)
	}
	level.Info(logger).Log("event", "turn.handled")

	explode := gm.DecideExplode(potato, g.Turns, g.HeatLevel)
	if explode {
		g, err = gm.games.EndGame(ctx, r.Namespace, req.ChannelID)
		if err != nil {
			return nil, fmt.Errorf("error ending game: %w", err)
		}
		level.Info(logger).Log("event", "game.ended")

		err = gm.rooms.IncrementDeaths(ctx, r.Namespace, r.ID, req.ActorUserID)
		if err != nil {
			return nil, fmt.Errorf("error incrementing death count: %w", err)
		}
	}

	return &StealResponse{
		Turn:         g.Turns,
		Potato:       potato,
		HolderUserID: g.HolderUserID,
		Exploded:     explode,
	}, nil
}

func (gm *GameMaster) Cook(ctx context.Context, req *CookRequest) (*CookResponse, error) {
	logger := log.WithSuffix(gm.logger, "namespace", req.Namespace, "room", req.RoomID, "channel", req.ChannelID)

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	r, err := gm.rooms.GetRoom(ctx, string(req.Namespace), req.RoomID)
	if err != nil {
		if !errors.Is(err, room.ErrRoomNotFound) {
			return nil, fmt.Errorf("error getting room: %w", err)
		}

		r, err = gm.rooms.CreateRoom(ctx, string(req.Namespace), req.RoomID)
		if err != nil {
			return nil, fmt.Errorf("error creating room: %w", err)
		}
		level.Info(logger).Log("event", "room.created")
	}

	g, err := gm.games.GetGame(ctx, r.Namespace, req.ChannelID)
	if err != nil && !errors.Is(err, game.ErrGameNotFound) {
		return nil, fmt.Errorf("error getting game: %w", err)
	}

	if errors.Is(err, game.ErrGameNotFound) || g.Finished {
		return nil, ErrNoOngoingGame
	}

	if req.ActorUserID != g.HolderUserID {
		return nil, &NotHolderError{g.HolderUserID}
	}

	potato, err := gm.GetPotato(g.PotatoKind)
	if err != nil {
		return nil, fmt.Errorf("error getting potato of kind '%s': %w", g.PotatoKind, err)
	}

	g, err = gm.games.IncrementHeatLevel(ctx, r.Namespace, req.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("error incrementing heat level: %w", err)
	}

	g, err = gm.games.NextTurn(ctx, r.Namespace, req.ChannelID, req.ActorUserID)
	if err != nil {
		return nil, fmt.Errorf("error making turn: %w", err)
	}
	level.Info(logger).Log("event", "turn.handled")

	explode := gm.DecideExplode(potato, g.Turns, g.HeatLevel)
	if explode {
		g, err = gm.games.EndGame(ctx, r.Namespace, req.ChannelID)
		if err != nil {
			return nil, fmt.Errorf("error ending game: %w", err)
		}
		level.Info(logger).Log("event", "game.ended")

		err = gm.rooms.IncrementDeaths(ctx, r.Namespace, r.ID, req.ActorUserID)
		if err != nil {
			return nil, fmt.Errorf("error incrementing death count: %w", err)
		}
	}

	return &CookResponse{
		Turn:         g.Turns,
		HeatLevel:    g.HeatLevel,
		Potato:       potato,
		HolderUserID: g.HolderUserID,
		Exploded:     explode,
	}, nil
}

func (gm *GameMaster) GetHolder(ctx context.Context, req *GetHolderRequest) (*GetHolderResponse, error) {
	logger := log.WithSuffix(gm.logger, "namespace", req.Namespace, "room", req.RoomID, "channel", req.ChannelID)

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	r, err := gm.rooms.GetRoom(ctx, string(req.Namespace), req.RoomID)
	if err != nil {
		if !errors.Is(err, room.ErrRoomNotFound) {
			return nil, fmt.Errorf("error getting room: %w", err)
		}

		r, err = gm.rooms.CreateRoom(ctx, string(req.Namespace), req.RoomID)
		if err != nil {
			return nil, fmt.Errorf("error creating room: %w", err)
		}
		level.Info(logger).Log("event", "room.created")
	}

	g, err := gm.games.GetGame(ctx, r.Namespace, req.ChannelID)
	if err != nil && !errors.Is(err, game.ErrGameNotFound) {
		return nil, fmt.Errorf("error getting game: %w", err)
	}

	if errors.Is(err, game.ErrGameNotFound) || g.Finished {
		return nil, ErrNoOngoingGame
	}

	potato, err := gm.GetPotato(g.PotatoKind)
	if err != nil {
		return nil, fmt.Errorf("error getting potato of kind '%s': %w", g.PotatoKind, err)
	}

	return &GetHolderResponse{
		Potato:       potato,
		HolderUserID: g.HolderUserID,
	}, nil
}

func (gm *GameMaster) GetLeaderboard(ctx context.Context, req *GetLeaderboardRequest) (*GetLeaderboardResponse, error) {
	logger := log.WithSuffix(gm.logger, "namespace", req.Namespace, "room", req.RoomID)

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	r, err := gm.rooms.GetRoom(ctx, string(req.Namespace), req.RoomID)
	if err != nil {
		if !errors.Is(err, room.ErrRoomNotFound) {
			return nil, fmt.Errorf("error getting room: %w", err)
		}

		r, err = gm.rooms.CreateRoom(ctx, string(req.Namespace), req.RoomID)
		if err != nil {
			return nil, fmt.Errorf("error creating room: %w", err)
		}
		level.Info(logger).Log("event", "room.created")
	}

	var leaderboard Scoreboard
	if req.Top > 0 {
		leaderboard = BuildLeaderboard(r.DeathCount, req.Top)
	} else {
		leaderboard = BuildLeaderboard(r.DeathCount)
	}

	return &GetLeaderboardResponse{
		Leaderboard: leaderboard,
	}, nil
}
