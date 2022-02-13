package game

import (
	"context"
	"errors"
)

var (
	ErrGameAlreadyExists = errors.New("game for channel already exists")
	ErrGameNotFound      = errors.New("game for channel not found")
)

type GameRepository interface {
	GetGame(ctx context.Context, namespace, channelID string) (*Game, error)
	CreateNewGame(ctx context.Context, namespace, roomID, channelID, potatoKind, startUserID string) (*Game, error)
	NextTurn(ctx context.Context, namespace, channelID, holderUserID string) (*Game, error)
	IncrementHeatLevel(ctx context.Context, namespace, channelID string) (*Game, error)
	EndGame(ctx context.Context, namespace, channelID string) (*Game, error)
}

type Game struct {
	Namespace    string
	RoomID       string
	ChannelID    string
	PotatoKind   string
	HeatLevel    int
	HolderUserID string
	Turns        int
	Finished     bool
}
