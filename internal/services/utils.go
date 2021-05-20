package services

import (
	"github.com/pkg/errors"
	"time"
)

func RunWithPeriod(period time.Duration, run func() error) error {
	uptimeTicker := time.NewTicker(period)

	for {
		select {
		case <-uptimeTicker.C:
			err := run()
			if err != nil {
				return errors.Wrap(err, "failed to run task")
			}
		}
	}
}