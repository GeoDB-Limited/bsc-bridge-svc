package bridge

import (
	"github.com/bsc-bridge-svc/internal/config"
	"github.com/sirupsen/logrus"
	"net/http"
)

type Service struct {
	cfg config.Config
	log *logrus.Logger
	client *http.Client
}

func New(cfg config.Config) *Service {
	return &Service{
		cfg: cfg,
		log: cfg.Logger(),
		client: &http.Client{},
	}
}
