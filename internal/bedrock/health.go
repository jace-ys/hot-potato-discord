package bedrock

import "github.com/heptiolabs/healthcheck"

type HealthCheckTarget interface {
	LivenessProbes() map[string]healthcheck.Check
	ReadinessProbes() map[string]healthcheck.Check
}

func (a *Admin) RegisterHealthChecks(target HealthCheckTarget) {
	for name, check := range target.LivenessProbes() {
		a.health.AddLivenessCheck(name, check)
	}

	for name, check := range target.ReadinessProbes() {
		a.health.AddReadinessCheck(name, check)
	}
}
