package ask

import (
	"ask-bot/src/datatypes/schedule"
	"database/sql/driver"
	"errors"
	"slices"
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

func (a *Ask) Schedule(kind TimeslotKind, begin time.Time, end time.Time) (schedule.Schedule, error) {
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

	if len(timeslots) == 0 {
		err = errors.New("no such kind in schedule")
		return nil, zaperr.Wrap(err, "",
			zap.String("kind", string(kind)))
	}

	// format begin\end time
	begin = begin.In(time.UTC).Add(a.timezone)
	end = end.In(time.UTC).Add(a.timezone)

	// get schedules
	var schedules schedule.Schedule
	for _, timeslot := range timeslots {

		var points []time.Time
		for _, s := range strings.Split(timeslot.TimePoints, ",") {
			s = strings.TrimSpace(s)

			point, err := time.Parse(time.TimeOnly, s)
			if err == nil {
				// time in database in correct timezone, but it is considered in UTC
				points = append(points, point)
				continue
			}

			point, err = time.Parse("15:04", s)
			if err == nil {
				points = append(points, point)
				continue
			}

			return nil, zaperr.Wrap(err, "failed to parse time point",
				zap.String("time point", s))
		}
		slices.SortFunc(points, func(a, b time.Time) int {
			if a.After(b) {
				return 1
			} else if a.Before(b) {
				return -1
			}
			return 0
		})

		s, err := schedule.Calculate(timeslot.Query, points, begin, end)
		if err != nil {
			return nil, zaperr.Wrap(err, "failed to get schedule for timeslot",
				zap.String("query", timeslot.Query),
				zap.Any("time points", points),
				zap.Time("begin", begin),
				zap.Time("end", end))
		}

		// from "desired time in local zone but in utc" to real time in utc to be equal in desired location
		for i := range s {
			s[i] = s[i].Add(-a.timezone)
		}

		schedules = schedules.Merge(s)
	}

	return schedules, nil
}
