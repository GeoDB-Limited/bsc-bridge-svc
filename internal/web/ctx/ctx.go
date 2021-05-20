package ctx

import (
	"context"
	"github.com/bsc-bridge-svc/internal/config"
	"github.com/bsc-bridge-svc/internal/data/postgres"
	"github.com/bsc-bridge-svc/internal/services/bridge"
	"github.com/sirupsen/logrus"
	"net/http"
)

const (
	ctxLog       = "ctxLog"
	ctxConfig    = "ctxConfig"
	ctxUsers     = "ctxUsers"
	ctxBridge    = "ctxBridge"
	ctxTransfers = "ctxTransfer"
)

// context getters and setters
// allows to set and retrieve context objects

func CtxConfig(cfg config.Config) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, ctxConfig, cfg)
	}
}

func Config(r *http.Request) config.Config {
	return r.Context().Value(ctxConfig).(config.Config)
}

func CtxLog(log *logrus.Logger) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, ctxLog, log)
	}
}

func Log(r *http.Request) *logrus.Logger {
	return r.Context().Value(ctxLog).(*logrus.Logger)
}

func CtxUsers(users postgres.Users) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, ctxUsers, users)
	}
}

func Users(r *http.Request) postgres.Users {
	return r.Context().Value(ctxUsers).(postgres.Users).New()
}

func CtxBridge(bridge *bridge.Service) func(ctx context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, ctxBridge, bridge)
	}
}

func Bridge(r *http.Request) *bridge.Service {
	return r.Context().Value(ctxBridge).(*bridge.Service)
}

func CtxTransfers(transfers postgres.Transfers) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, ctxTransfers, transfers)
	}
}

func Transfers(r *http.Request) postgres.Transfers {
	return r.Context().Value(ctxTransfers).(postgres.Transfers)
}
