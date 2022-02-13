package room

import (
	"context"
	"errors"
)

var (
	ErrRoomAlreadyExists = errors.New("room for guild already exists")
	ErrRoomNotFound      = errors.New("room for guild not found")
)

type RoomRepository interface {
	GetRoom(ctx context.Context, namespace, roomID string) (*Room, error)
	CreateRoom(ctx context.Context, namespace, roomID string) (*Room, error)
	IncrementDeaths(ctx context.Context, namespace, roomID, userID string) error
}

type Room struct {
	Namespace  string
	ID         string
	DeathCount []DeathCounter
}

type DeathCounter struct {
	UserID string
	Count  int
}
