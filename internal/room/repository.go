package room

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jace-ys/hot-potato-discord/internal/room/store"
	"github.com/lib/pq"
)

type Repository struct {
	db    *sql.DB
	store *store.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db:    db,
		store: store.New(db),
	}
}

func (r *Repository) GetRoom(ctx context.Context, namespace, roomID string) (*Room, error) {
	room, err := r.store.GetRoom(ctx, store.GetRoomParams{
		Namespace: namespace,
		ID:        roomID,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRoomNotFound
		}
		return nil, err
	}

	deaths, err := r.store.ListDeathCount(ctx, store.ListDeathCountParams{
		Namespace: namespace,
		RoomID:    roomID,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRoomNotFound
		}
		return nil, err
	}

	return StoreToDomain(room, deaths), nil
}

func (r *Repository) CreateRoom(ctx context.Context, namespace, roomID string) (*Room, error) {
	room, err := r.store.InsertRoom(ctx, store.InsertRoomParams{
		Namespace: namespace,
		ID:        roomID,
	})
	if err != nil {
		var pqErr *pq.Error
		switch {
		case errors.As(err, &pqErr) && pqErr.Code.Name() == "unique_violation":
			return nil, ErrRoomAlreadyExists
		}
		return nil, err
	}

	return r.GetRoom(ctx, room.Namespace, room.ID)
}

func (r *Repository) IncrementDeaths(ctx context.Context, namespace, roomID, userID string) error {
	err := r.store.IncrementDeathCount(ctx, store.IncrementDeathCountParams{
		Namespace: namespace,
		RoomID:    roomID,
		UserID:    userID,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRoomNotFound
		}
		return err
	}

	return err
}

func StoreToDomain(room store.Room, deaths []store.Death) *Room {
	r := &Room{
		Namespace:  room.Namespace,
		ID:         room.ID,
		DeathCount: make([]DeathCounter, len(deaths)),
	}

	for i, row := range deaths {
		r.DeathCount[i] = DeathCounter{UserID: row.UserID, Count: int(row.Count.Int32)}
	}

	return r
}
