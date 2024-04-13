package schedule

import (
	"time"
)

type Schedule []time.Time

func Calculate(timeslot string, time_points []time.Time, begin time.Time, end time.Time) (Schedule, error) {
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
			if dates[i].Compare(slots[j]) == 0 {
				i, j = i+1, j+1
				continue
			}

			if dates[i].Compare(slots[j]) < 0 {
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
			if dates[i].Compare(slots[j]) == 0 {
				slots = append(slots[:j], slots[j+1:]...)
				i = i + 1
				continue
			}

			if dates[i].Compare(slots[j]) < 0 {
				i = i + 1
			} else {
				j = j + 1
			}
		}
	}

	var schedule Schedule
	for i := range slots {
		for j := range time_points {
			slot := MergeDatetime(slots[i], time_points[j])

			if slot.Compare(begin) >= 0 && slot.Compare(end) < 0 {
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

func (schedule Schedule) Merge(other Schedule) (result Schedule) {
	i, j := 0, 0

	for i < len(schedule) && j < len(other) {
		if schedule[i].Compare(other[j]) == 0 {
			result = append(result, schedule[i])
			i, j = i+1, j+1
			continue
		}

		if schedule[i].Compare(other[j]) < 0 {
			result = append(result, schedule[i])
			i = i + 1
		} else {
			result = append(result, other[j])
			j = j + 1
		}
	}

	if i < len(schedule) {
		result = append(result, schedule[i:]...)
	}

	if j < len(other) {
		result = append(result, other[j:]...)
	}

	return result
}

func (schedule Schedule) Exclude(other []time.Time) (result Schedule) {
	i, j := 0, 0

	for i < len(schedule) && j < len(other) {
		if schedule[i].Compare(other[j]) == 0 {
			i, j = i+1, j+1
			continue
		}

		if schedule[i].Compare(other[j]) < 0 {
			result = append(result, schedule[i])
			i = i + 1
		} else {
			j = j + 1
		}
	}

	if i < len(schedule) {
		result = append(result, schedule[i:]...)
	}

	return result
}
