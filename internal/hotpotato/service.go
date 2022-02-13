package hotpotato

import (
	"context"
	"errors"
)

type Service interface {
	Toss(ctx context.Context, req *TossRequest) (*TossResponse, error)
	Steal(ctx context.Context, req *StealRequest) (*StealResponse, error)
	Cook(ctx context.Context, req *CookRequest) (*CookResponse, error)
	GetHolder(ctx context.Context, req *GetHolderRequest) (*GetHolderResponse, error)
	GetLeaderboard(ctx context.Context, req *GetLeaderboardRequest) (*GetLeaderboardResponse, error)
}

type TossRequest struct {
	Namespace    string
	RoomID       string
	ChannelID    string
	ActorUserID  string
	TargetUserID string
}

func (r *TossRequest) Validate() error {
	switch {
	case r.Namespace == "":
		return errors.New("missing namespace")
	case r.RoomID == "":
		return errors.New("missing room ID")
	case r.ChannelID == "":
		return errors.New("missing channel ID")
	case r.ActorUserID == "":
		return errors.New("missing actor user ID")
	case r.TargetUserID == "":
		return errors.New("missing target user ID")
	default:
		return nil
	}
}

type TossResponse struct {
	Turn         int
	Potato       Potato
	HolderUserID string
	Exploded     bool
}

type StealRequest struct {
	Namespace    string
	RoomID       string
	ChannelID    string
	ActorUserID  string
	TargetUserID string
}

func (r *StealRequest) Validate() error {
	switch {
	case r.Namespace == "":
		return errors.New("missing namespace")
	case r.RoomID == "":
		return errors.New("missing room ID")
	case r.ChannelID == "":
		return errors.New("missing channel ID")
	case r.ActorUserID == "":
		return errors.New("missing actor user ID")
	case r.TargetUserID == "":
		return errors.New("missing target user ID")
	default:
		return nil
	}
}

type StealResponse struct {
	Turn         int
	Potato       Potato
	HolderUserID string
	Exploded     bool
}

type CookRequest struct {
	Namespace   string
	RoomID      string
	ChannelID   string
	ActorUserID string
}

func (r *CookRequest) Validate() error {
	switch {
	case r.Namespace == "":
		return errors.New("missing namespace")
	case r.RoomID == "":
		return errors.New("missing room ID")
	case r.ChannelID == "":
		return errors.New("missing channel ID")
	case r.ActorUserID == "":
		return errors.New("missing actor user ID")
	default:
		return nil
	}
}

type CookResponse struct {
	Turn         int
	Potato       Potato
	HeatLevel    int
	HolderUserID string
	Exploded     bool
}

type GetHolderRequest struct {
	Namespace string
	RoomID    string
	ChannelID string
}

func (r *GetHolderRequest) Validate() error {
	switch {
	case r.Namespace == "":
		return errors.New("missing namespace")
	case r.RoomID == "":
		return errors.New("missing room ID")
	case r.ChannelID == "":
		return errors.New("missing channel ID")
	default:
		return nil
	}
}

type GetHolderResponse struct {
	Potato       Potato
	HolderUserID string
}

type GetLeaderboardRequest struct {
	Namespace string
	RoomID    string
	Top       int
}

func (r *GetLeaderboardRequest) Validate() error {
	switch {
	case r.Namespace == "":
		return errors.New("missing namespace")
	case r.RoomID == "":
		return errors.New("missing room ID")
	case r.Top < 0:
		return errors.New("top cannot be negative")
	default:
		return nil
	}
}

type GetLeaderboardResponse struct {
	Leaderboard Scoreboard
}
