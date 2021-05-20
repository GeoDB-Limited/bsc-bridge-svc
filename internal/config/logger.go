package config

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Logger interface {
	Logger() *logrus.Logger
}

type logger struct {
	Level string `yaml:"level"`
}

func (l *logger) Logger() *logrus.Logger {
	level, err := logrus.ParseLevel(l.Level)
	if err != nil {
		panic(errors.Wrapf(err, "failed to parse logging Level %s", l.Level))
	}

	log := logrus.New()
	log.SetLevel(level)

	return log
}
