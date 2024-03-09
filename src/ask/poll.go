package ask

import (
	"database/sql/driver"
	"errors"
	"strconv"
	"strings"

	"github.com/hori-ryota/zaperr"
	"github.com/leporo/sqlf"
	"go.uber.org/zap"
)

type VkIDs []int

func (ids VkIDs) Value() (driver.Value, error) {
	return []int(ids), nil
}

func (ids *VkIDs) Scan(value interface{}) error {
	if value == nil {
		ids = &VkIDs{}
		return nil
	}

	if str, err := driver.String.ConvertValue(value); err == nil {
		if v, ok := str.(string); ok {
			strs := strings.Split(v, ",")
			var ints []int

			for _, s := range strs {
				i, err := strconv.Atoi(s)
				if err != nil {
					return err
				}

				ints = append(ints, i)
			}

			*ids = VkIDs(ints)
			return nil
		}

	}
	return errors.New("failed to scan VkIDs")
}

type PendingPoll struct {
	Role         string `db:"role"`
	Hashtag      string `db:"hashtag"`
	Participants VkIDs  `db:"participants"`
	Greetings    string `db:"greetings"`
}

type Poll struct {
}

func (a *Ask) PendingPolls() ([]PendingPoll, error) {
	var polls []PendingPoll

	query := sqlf.From("polls").
		Bind(&Poll{})

	err := a.db.Select(&polls, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get polls",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return polls, nil
}
