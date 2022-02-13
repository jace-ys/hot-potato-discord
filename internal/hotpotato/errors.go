package hotpotato

import "errors"

var (
	ErrNoOngoingGame      = errors.New("no ongoing game found")
	ErrInvalidPotatoKind  = errors.New("unrecognised potato kind")
	ErrSelfStealUnallowed = errors.New("cannot steal potato from self")
)

type NotHolderError struct {
	HolderUserID string
}

func (e *NotHolderError) Error() string {
	return "user does not current hold the potato"
}
