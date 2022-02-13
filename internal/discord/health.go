package discord 

import (
	"errors"
	"fmt"
	"time"

	"github.com/heptiolabs/healthcheck"
)

func (b *Bot) LivenessProbes() map[string]healthcheck.Check {
	return map[string]healthcheck.Check{}
}

func (b *Bot) ReadinessProbes() map[string]healthcheck.Check {
	return map[string]healthcheck.Check{
		"server": healthcheck.HTTPGetCheck(fmt.Sprintf("http://%s/ping", b.server.Addr), time.Second),
		"discord": func() error {
			if b.discord.HeartbeatLatency() > time.Minute {
				return errors.New("heartbeat no ack in the last minute")
			}
			return nil
		},
	}
}
