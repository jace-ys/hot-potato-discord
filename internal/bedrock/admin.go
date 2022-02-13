package bedrock

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Admin struct {
	logger log.Logger
	server *http.Server
	health healthcheck.Handler
}

func NewAdmin(logger log.Logger, port int) (*Admin, error) {
	admin := &Admin{
		logger: logger,
		health: healthcheck.NewHandler(),
	}

	admin.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: admin.router(),
	}

	return admin, nil
}

func (a *Admin) router() http.Handler {
	router := mux.NewRouter()

	admin := router.PathPrefix("/admin/").Subrouter()
	admin.Handle("/metrics", promhttp.Handler())
	admin.HandleFunc("/live", a.health.LiveEndpoint)
	admin.HandleFunc("/ready", a.health.ReadyEndpoint)

	return router
}

func (a *Admin) Start(ctx context.Context) error {
	level.Info(a.logger).Log("event", "server.started", "name", "admin", "addr", a.server.Addr)
	defer level.Info(a.logger).Log("event", "server.stopped")

	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start admin server: %w", err)
	}

	return nil
}

func (a *Admin) Stop(ctx context.Context) error {
	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown admin server: %w", err)
	}

	return nil
}
