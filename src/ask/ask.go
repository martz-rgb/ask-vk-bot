package ask

import (
	"ask-bot/src/ask/db"
	"time"

	"github.com/hori-ryota/zaperr"
	"github.com/leporo/sqlf"
)

type Ask struct {
	config *Config
	db     *db.DB

	timezone time.Duration
}

func New(config *Config) *Ask {
	sqlf.SetDialect(sqlf.NoDialect)

	return &Ask{
		config:   config,
		timezone: time.Duration(config.Timezone) * time.Hour,
	}
}

func (a *Ask) Init(path string, schema string, allow_deletion bool) error {
	d, err := db.NewDB(path)
	if err != nil {
		return zaperr.Wrap(err, "failed to create new db")
	}
	if err = d.Init(schema, allow_deletion); err != nil {
		return zaperr.Wrap(err, "failed to create new db")
	}

	a.db = d

	return nil
}
