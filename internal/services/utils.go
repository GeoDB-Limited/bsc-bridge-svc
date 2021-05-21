package services

import (
	"github.com/sirupsen/logrus"
	"time"
)

func RunWithPeriod(log *logrus.Logger, period time.Duration, run func() error) {
	defer func() {
		// recover if something has broken
		if rvr := recover(); rvr != nil {
			log.Error("app panicked\n", rvr)
		}
	}()

	uptimeTicker := time.NewTicker(period)

	for {
		select {
		case <-uptimeTicker.C:
			err := run()
			if err != nil {
				log.WithError(err).Error("run function finished with error")
			}
		}
	}
}
