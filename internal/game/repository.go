package game

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jace-ys/hot-potato-discord/internal/game/store"
)

type Repository struct {
	db    *sql.DB
	store *store.Queries
}

func NewRepository(database *sql.DB) *Repository {
	return &Repository{
		db:    database,
		store: store.New(database),
	}
}

func (r *Repository) GetGame(ctx context.Context, namespace, channelID string) (*Game, error) {
	game, err := r.store.GetGame(ctx, store.GetGameParams{
		Namespace: namespace,
		ChannelID: channelID,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrGameNotFound
		}
		return nil, err
	}

	return StoreToDomain(game), err
}

func (r *Repository) CreateNewGame(ctx context.Context, namespace, roomID, channelID, potatoKind, startUserID string) (*Game, error) {
	game, err := r.store.UpsertGame(ctx, store.UpsertGameParams{
		Namespace:    namespace,
		RoomID:       roomID,
		ChannelID:    channelID,
		PotatoKind:   potatoKind,
		HeatLevel:    1,
		HolderUserID: startUserID,
		Turns:        0,
		Finished:     false,
	})

	return StoreToDomain(game), err
}

func (r *Repository) NextTurn(ctx context.Context, namespace, channelID, holderUserID string) (*Game, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	game, err := r.store.WithTx(tx).UpdateHolder(ctx, store.UpdateHolderParams{
		Namespace:    namespace,
		ChannelID:    channelID,
		HolderUserID: holderUserID,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrGameNotFound
		}
		return nil, err
	}

	game, err = r.store.WithTx(tx).IncrementTurns(ctx, store.IncrementTurnsParams{
		Namespace: namespace,
		ChannelID: channelID,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrGameNotFound
		}
		return nil, err
	}

	return StoreToDomain(game), tx.Commit()
}

func (r *Repository) IncrementHeatLevel(ctx context.Context, namespace, channelID string) (*Game, error) {
	game, err := r.store.IncreaseHeatLevel(ctx, store.IncreaseHeatLevelParams{
		Namespace: namespace,
		ChannelID: channelID,
		Amount:    1,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrGameNotFound
		}
		return nil, err
	}

	return StoreToDomain(game), err
}

func (r *Repository) EndGame(ctx context.Context, namespace, channelID string) (*Game, error) {
	game, err := r.store.EndGame(ctx, store.EndGameParams{
		Namespace: namespace,
		ChannelID: channelID,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrGameNotFound
		}
		return nil, err
	}

	return StoreToDomain(game), err
}

func StoreToDomain(game store.Game) *Game {
	return &Game{
		Namespace:    game.Namespace,
		RoomID:       game.RoomID,
		ChannelID:    game.ChannelID,
		PotatoKind:   game.PotatoKind,
		HeatLevel:    int(game.HeatLevel),
		HolderUserID: game.HolderUserID,
		Turns:        int(game.Turns),
		Finished:     game.Finished,
	}
}
