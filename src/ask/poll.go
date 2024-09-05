package ask

import (
	"github.com/hori-ryota/zaperr"
	"github.com/leporo/sqlf"
	"go.uber.org/zap"
)

type OngoingPoll struct {
	Role string `db:"role"`
	Post int    `db:"post"`
}

func (a *Ask) OngoingPolls() ([]OngoingPoll, error) {
	var polls []OngoingPoll

	query := sqlf.From("ongoing_polls").
		Bind(&OngoingPoll{})

	err := a.db.Select(&polls, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get ongoing polls",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return polls, nil
}

func (a *Ask) AddOngoingPoll(role string, post int) error {
	query := sqlf.InsertInto("ongoing_polls").
		Set("role", role).
		Set("post", post)

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to add ongoing poll",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return nil
}

type Poll struct {
	PendingPoll

	Post int `db:"post"`
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
