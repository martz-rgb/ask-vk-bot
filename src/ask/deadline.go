package ask

import (
	"database/sql/driver"
	"errors"
	"time"

	"github.com/hori-ryota/zaperr"
	"github.com/leporo/sqlf"
	"go.uber.org/zap"
)

type DeadlineCause string

var DeadlineCauses = struct {
	Init   DeadlineCause
	Answer DeadlineCause
	Delay  DeadlineCause
	Rest   DeadlineCause
	Freeze DeadlineCause
	Other  DeadlineCause
}{
	Init:   "Init",
	Answer: "Answer",
	Delay:  "Delay",
	Rest:   "Rest",
	Freeze: "Freeze",
	Other:  "Other",
}

func (c DeadlineCause) Value() (driver.Value, error) {
	return string(c), nil
}

func (c *DeadlineCause) Scan(value interface{}) error {
	if value == nil {
		return errors.New("DeadlineCause is not nullable")
	}
	if str, err := driver.String.ConvertValue(value); err == nil {
		if v, ok := str.(string); ok {
			// check if is valid
			if v != string(DeadlineCauses.Init) &&
				v != string(DeadlineCauses.Answer) &&
				v != string(DeadlineCauses.Delay) &&
				v != string(DeadlineCauses.Rest) &&
				v != string(DeadlineCauses.Freeze) &&
				v != string(DeadlineCauses.Other) {
				return errors.New("value is not valid DeadlineCause value")
			}
			*c = DeadlineCause(v)
			return nil
		}
	}
	return errors.New("failed to scan DeadlineCause")
}

type Deadline struct {
	Member    int           `db:"member"`
	Diff      int           `db:"diff"` // unix time in seconds!
	Kind      DeadlineCause `db:"kind"`
	Cause     string        `db:"cause"`
	Timestamp time.Time     `db:"timestamp"`
}

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
