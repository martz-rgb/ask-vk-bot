package ask

import (
	"time"

	"github.com/hori-ryota/zaperr"
	"github.com/leporo/sqlf"
	"go.uber.org/zap"
)

// deadline
func (a *Ask) Deadline(member int) (time.Time, error) {
	var deadline int64

	// should be at least one record
	query := sqlf.From("deadline").
		Select("SUM(diff)").
		Where("member = ?", member)

	err := a.db.Get(&deadline, query.String(), query.Args()...)
	if err != nil {
		return time.Time{}, zaperr.Wrap(err, "failed to get deadline for memeber",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return time.Unix(deadline, 0).Add(a.timezone), nil
}

func (a *Ask) HistoryDeadline(member int) ([]Deadline, error) {
	var history []Deadline

	query := sqlf.From("deadline").
		Bind(&Deadline{}).
		Where("member = ?", member).
		OrderBy("timestamp DESC")

	err := a.db.Select(&history, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get history of deadline for member",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return history, nil
}

// TO-DO maybe another way to insert
func (a *Ask) ChangeDeadline(member int, diff time.Duration, kind DeadlineCause, cause string) error {
	query := sqlf.InsertInto("deadline").
		Set("member", member).
		Set("diff", diff.Seconds()).
		Set("kind", kind).
		Set("cause", cause)

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to insert deadline event",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return nil
}
