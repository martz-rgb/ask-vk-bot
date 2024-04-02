package ask

import (
	"ask-bot/src/ask/schedule"
	"database/sql/driver"
	"errors"
	"strings"
	"time"

	"github.com/hori-ryota/zaperr"
	"github.com/leporo/sqlf"
	"go.uber.org/zap"
)

type TimeslotKind string

var TimeslotKinds = struct {
	Polls       TimeslotKind
	Greetings   TimeslotKind
	Answers     TimeslotKind
	FreeAnswers TimeslotKind
	Leavings    TimeslotKind
}{
	Polls:       "Polls",
	Greetings:   "Greetings",
	Answers:     "Answers",
	FreeAnswers: "Free Answers",
	Leavings:    "Leavings",
}

func (s TimeslotKind) Value() (driver.Value, error) {
	return string(s), nil
}

func (s *TimeslotKind) Scan(value interface{}) error {
	if value == nil {
		return errors.New("TimeslotKind is not nullable")
	}

	if str, err := driver.String.ConvertValue(value); err == nil {
		if v, ok := str.(string); ok {
			if v != string(TimeslotKinds.Polls) &&
				v != string(TimeslotKinds.Greetings) &&
				v != string(TimeslotKinds.Answers) &&
				v != string(TimeslotKinds.FreeAnswers) &&
				v != string(TimeslotKinds.Leavings) {
				return errors.New("value is not valid TimeslotKind value")
			}

			*s = TimeslotKind(v)
			return nil
		}

	}
	return errors.New("failed to scan TimeslotKind")
}

type Timeslot struct {
	Id         int          `db:"id"`
	Kind       TimeslotKind `db:"kind"`
	Query      string       `db:"query"`
	TimePoints string       `db:"time_points"`
}

func (a *Ask) Schedule(kind TimeslotKind, begin time.Time, end time.Time) ([]time.Time, error) {
	// get timeslots
	var timeslots []Timeslot
	query := sqlf.From("schedule").
		Bind(&Timeslot{}).
		Where("kind = ?", kind)

	err := a.db.Select(&timeslots, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get schedule",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	// format begin\end time
	begin = begin.In(time.UTC).Add(a.timezone)
	end = end.In(time.UTC).Add(a.timezone)

	// get schedules
	var schedules []time.Time
	for _, timeslot := range timeslots {

		var points []time.Time
		for _, s := range strings.Split(timeslot.TimePoints, ",") {
			s = strings.TrimSpace(s)
			point, err := time.Parse(time.TimeOnly, s)
			if err != nil {
				return nil, zaperr.Wrap(err, "failed to parse time point",
					zap.String("time point", s))
			}
			points = append(points, point)
		}

		s, err := schedule.Schedule(timeslot.Query, points, begin, end)
		if err != nil {
			return nil, zaperr.Wrap(err, "failed to get schedule for timeslot",
				zap.String("query", timeslot.Query),
				zap.Any("time points", points),
				zap.Time("begin", begin),
				zap.Time("end", end))
		}

		schedules = schedule.MergeSchedules(schedules, s)
	}

	return schedules, nil
}
