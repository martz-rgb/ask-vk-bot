package schedule

import (
	"slices"
	"time"
)

type Timeslot interface {
	// Slots returnes sorted slice of normalised dates
	// (only date is meaningful, time set to zero UTC)
	// begin and end should be normalised dates too
	Slots(begin time.Time, end time.Time) []time.Time
}

type EverydayTimeslot struct{}

func (m *EverydayTimeslot) Slots(begin time.Time, end time.Time) (dates []time.Time) {
	date := time.Date(begin.Year(), begin.Month(), begin.Day(), 0, 0, 0, 0, time.UTC)

	for date.Compare(end) < 0 {
		dates = append(dates, date)
		date = date.AddDate(0, 0, 1)
	}

	return dates
}

type EveryTimeslot struct {
	Weekday int
}

func (t *EveryTimeslot) Slots(begin time.Time, end time.Time) (dates []time.Time) {
	diff := t.Weekday - int(begin.Weekday())
	if diff < 0 {
		diff += 7
	}

	date := begin.AddDate(0, 0, diff)

	for date.Compare(end) < 0 {
		dates = append(dates, date)
		date = date.AddDate(0, 0, 7)
	}

	return dates
}

type NumberTimeslot struct {
	Number int
}

func (t *NumberTimeslot) findDate(year int, month time.Month) time.Time {
	number := t.Number
	if number < 0 {
		number += 1
		month += 1
	}

	return time.Date(year, month, number, 0, 0, 0, 0, time.UTC)
}

func (t *NumberTimeslot) Slots(begin time.Time, end time.Time) (dates []time.Time) {
	year, month, _ := begin.Date()
	date := t.findDate(year, month)

	for date.Compare(end) < 0 {
		if date.Compare(begin) >= 0 {
			dates = append(dates, date)
		}

		month++
		date = t.findDate(year, month)
	}

	return dates
}

type OrderTimeslot struct {
	Weekday int
	Order   int
}

func (t *OrderTimeslot) findDate(year int, month time.Month) (dates []time.Time) {
	if t.Order > 0 {
		// find first one
		diff := t.Weekday - int(time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).Weekday())
		if diff < 0 {
			diff += 7
		}

		date := time.Date(year, month, 1+diff+(t.Order-1)*7, 0, 0, 0, 0, time.UTC)
		if date.Month() == month {
			dates = append(dates, date)
		}
	} else if t.Order < 0 {
		// find last one
		last_day := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC)
		diff := t.Weekday - int(last_day.Weekday())
		if diff > 0 {
			diff -= 7
		}

		date := last_day.AddDate(0, 0, diff+(t.Order+1)*7)
		if date.Month() == month {
			dates = append(dates, date)
		}
	}

	return dates
}

func (t *OrderTimeslot) Slots(begin time.Time, end time.Time) (dates []time.Time) {
	year, month, _ := begin.Date()
	iter := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

	for iter.Before(end) {
		date := t.findDate(iter.Year(), iter.Month())
		for _, d := range date {
			if d.Compare(begin) >= 0 && d.Compare(end) < 0 {
				dates = append(dates, d)
			}
		}
		iter = iter.AddDate(0, 0, monthDays(iter.Year(), iter.Month()))
	}

	return dates
}

type MultipleWeekdayTimeslot struct {
	Weekday int
	Module  int
	Residue []int
}

func (t *MultipleWeekdayTimeslot) Slots(begin time.Time, end time.Time) (dates []time.Time) {
	diff := t.Weekday - int(begin.Weekday())
	if diff < 0 {
		diff += 7
	}
	date := begin.AddDate(0, 0, diff)

	for date.Compare(end) < 0 {
		_, week := date.ISOWeek()

		if slices.Contains(t.Residue, week%t.Module) {
			dates = append(dates, date)
		}

		date = date.AddDate(0, 0, 7)
	}

	return dates
}

type MultipleDailyKind int

var MultipleKinds = struct {
	Day  MultipleDailyKind
	Date MultipleDailyKind
}{
	Day:  0,
	Date: 1,
}

type MultipleDailyTimeslot struct {
	Kind MultipleDailyKind

	Module  int
	Residue []int
}

func (t *MultipleDailyTimeslot) Slots(begin time.Time, end time.Time) (dates []time.Time) {
	var index func(time.Time) int

	switch t.Kind {
	case MultipleKinds.Date:
		index = func(date time.Time) int {
			return date.Day()
		}
	case MultipleKinds.Day:
		index = func(date time.Time) int {
			return date.YearDay()
		}
	}

	date := time.Date(begin.Year(), begin.Month(), begin.Day(), 0, 0, 0, 0, time.UTC)

	for date.Compare(end) < 0 {
		if slices.Contains(t.Residue, index(date)%t.Module) {
			dates = append(dates, date)
		}

		date = date.AddDate(0, 0, 1)
	}

	return dates
}

func monthDays(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
