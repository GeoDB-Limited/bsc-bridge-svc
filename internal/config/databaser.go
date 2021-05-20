package config

import (
	"database/sql"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sync"
)

type Databaser interface {
	DB() *sql.DB
}

type databaser struct {
	Url string `yaml:"url"`

	cache struct {
		db *sql.DB
	}

	log *logrus.Logger
	sync.Once
}

// DB reads database when called once
func (d *databaser) DB() *sql.DB {
	d.Once.Do(func() {
		var err error
		d.cache.db, err = sql.Open("postgres", d.Url)
		if err != nil {
			panic(err)
		}

		if err := d.cache.db.Ping(); err != nil {
			panic(errors.Wrap(err, "database unavailable"))
		}
	})
	return d.cache.db
}
