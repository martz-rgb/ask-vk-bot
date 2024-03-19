package ask

import (
	"database/sql"
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

type Poll struct {
	Role

	Participants VkIDs         `db:"participants"`
	Greetings    string        `db:"greetings"`
	Post         sql.NullInt32 `db:"post"`
}

func (a *Ask) PendingPolls() ([]Poll, error) {
	var polls []Poll

	query := sqlf.From("polls").
		Bind(&Poll{})

	err := a.db.Select(&polls, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get pending polls",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return polls, nil
}

func (a *Ask) Polls() ([]Poll, error) {
	var polls []Poll

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

func (a *Ask) AddOngoingPoll(post int, role string) error {
	query := sqlf.InsertInto("ongoing_polls").
		Set("post", post).
		Set("role", role)

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to add ongoing poll",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return nil
}
