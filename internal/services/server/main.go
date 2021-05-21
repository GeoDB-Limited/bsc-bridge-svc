package server

import (
	"context"
	"github.com/bsc-bridge-svc/internal/config"
	"github.com/bsc-bridge-svc/internal/data/postgres"
	"github.com/bsc-bridge-svc/internal/services"
	"github.com/bsc-bridge-svc/internal/services/bridge"
	"github.com/bsc-bridge-svc/internal/services/sender"
	"github.com/bsc-bridge-svc/internal/web/ctx"
	"github.com/bsc-bridge-svc/internal/web/handlers"
	"github.com/bsc-bridge-svc/internal/web/logging"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Service struct {
	cfg    config.Config
	ctx    context.Context
	log    *logrus.Logger
	bridge *bridge.Service
	sender *sender.Service
}

func New(cfg config.Config, ctx context.Context) *Service {
	return &Service{
		cfg:    cfg,
		ctx:    ctx,
		log:    cfg.Logger(),
		bridge: bridge.New(cfg),
		sender: sender.New(cfg, ctx),
	}
}

func (s *Service) Run() error {
	defer func() {
		// recover if something has broken
		if rvr := recover(); rvr != nil {
			s.log.Error("app panicked\n", rvr)
		}
	}()

	go services.RunWithPeriod(s.log, 5*time.Second, s.sender.Send)
	go services.RunWithPeriod(s.log, 10*time.Second, s.sender.Refund)

	s.log.WithField("port", s.cfg.Address()).Info("Starting server")
	if err := http.ListenAndServe(s.cfg.Address(), s.router()); err != nil {
		return errors.Wrap(err, "listener failed")
	}
	return nil
}

func (s *Service) router() chi.Router {
	router := chi.NewRouter()

	router.Use(
		logging.Middleware(s.log),
		ctx.Middleware(
			ctx.CtxLog(s.log),
			ctx.CtxConfig(s.cfg),
			ctx.CtxUsers(postgres.NewUsers(s.cfg)),
			ctx.CtxTransfers(postgres.NewTransfers(s.cfg)),
			ctx.CtxBridge(s.bridge),
		),
	)

	// routes of the service
	router.Route("/bsc/users", func(r chi.Router) {
		r.Get("/", handlers.GetUser)
	})

	return router
}
