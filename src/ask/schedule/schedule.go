package schedule

import (
	"time"
)

type TimeslotError error

func Schedule(timeslot string, time_points []time.Time, begin time.Time, end time.Time) ([]time.Time, error) {
	lexems, err := toLexemes(timeslot)
	if err != nil {
		return nil, err
	}

	include, exclude, err := toTimeslots(lexems)
	if err != nil {
		return nil, err
	}

	begin_date := time.Date(begin.Year(), begin.Month(), begin.Day(), 0, 0, 0, 0, time.UTC)
	// +1 to move border
	end_date := time.Date(end.Year(), end.Month(), end.Day()+1, 0, 0, 0, 0, time.UTC)

	slots := []time.Time{}
	for _, add := range include {
		dates := add.Slots(begin_date, end_date)

		i, j := 0, 0

		for i < len(dates) && j < len(slots) {
			if compare(dates[i], slots[j]) == 0 {
				i, j = i+1, j+1
				continue
			}

			if compare(dates[i], slots[j]) < 0 {
				slots = append(slots[:j+1], slots[j:]...)
				slots[j] = dates[i]
				i, j = i+1, j+1
			} else {
				j = j + 1
			}
		}

		if i < len(dates) {
			slots = append(slots, dates[i:]...)
		}
	}

	for _, del := range exclude {
		dates := del.Slots(begin_date, end_date)

		i, j := 0, 0

		for i < len(dates) && j < len(slots) {
			if compare(dates[i], slots[j]) == 0 {
				slots = append(slots[:j], slots[j+1:]...)
				i = i + 1
				continue
			}

			if compare(dates[i], slots[j]) < 0 {
				i = i + 1
			} else {
				j = j + 1
			}
		}
	}

	var schedule []time.Time
	for i := range slots {
		for j := range time_points {
			slot := MergeDatetime(slots[i], time_points[j])

			if compare(slot, begin) >= 0 && compare(slot, end) < 0 {
				schedule = append(schedule, slot)
			}
		}
	}

	return schedule, nil
}

func MergeDatetime(date time.Time, moment time.Time) time.Time {
	year, month, day := date.Date()
	hour, minute, second := moment.Clock()

	return time.Date(year, month, day, hour, minute, second, 0, time.UTC)
}

func compare(a time.Time, b time.Time) int {
	if a.After(b) {
		return 1
	} else if a.Before(b) {
		return -1
	}

	return 0
}

func MergeSchedules(a []time.Time, b []time.Time) (result []time.Time) {
	i, j := 0, 0

	for i < len(a) && j < len(b) {
		if compare(a[i], b[j]) == 0 {
			result = append(result, a[i])
			i, j = i+1, j+1
			continue
		}

		if compare(a[i], b[j]) < 0 {
			result = append(result, a[i])
			i = i + 1
		} else {
			result = append(result, b[j])
			j = j + 1
		}
	}

	if i < len(a) {
		result = append(result, a[i:]...)
	}

	if j < len(b) {
		result = append(result, b[j:]...)
	}

	return result
}

func ExcludeSchedule(a []time.Time, b []time.Time) (result []time.Time) {
	i, j := 0, 0

	for i < len(a) && j < len(b) {
		if compare(a[i], b[j]) == 0 {
			i, j = i+1, j+1
			continue
		}

		if compare(a[i], b[j]) < 0 {
			result = append(result, a[i])
			i = i + 1
		} else {
			j = j + 1
		}
	}

	if i < len(a) {
		result = append(result, a[i:]...)
	}

	return result
}
