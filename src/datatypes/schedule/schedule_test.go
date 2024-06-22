package schedule

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

// TO-DO: tests for merge and exclude schedules?
func TestLexemesWeekday(t *testing.T) {
	query := "first Mondays"
	expected := []Lexem{
		{
			Kind:   LexemKinds.Order,
			Params: []int{1},
		},
		{
			Kind:   LexemKinds.Weekday,
			Params: []int{1},
		},
	}

	err := test_query(query, expected)
	if err != nil {
		t.Fatal(err)
	}
}

// check lexemes
func TestLexemesWithExcept(t *testing.T) {
	query := "every day except Sundays"
	expected := []Lexem{
		{
			Kind: LexemKinds.All,
		},
		{
			Kind: LexemKinds.Day,
		},
		{
			Kind: LexemKinds.Except,
		},
		{
			Kind:   LexemKinds.Weekday,
			Params: []int{0},
		},
	}

	err := test_query(query, expected)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLexemesList(t *testing.T) {
	query := "Tuesdays, Thursdays, Saturdays except odd date, -1"
	expected := []Lexem{
		{
			Kind:   LexemKinds.Weekday,
			Params: []int{2},
		},
		{
			Kind:   LexemKinds.Weekday,
			Params: []int{4},
		},
		{
			Kind:   LexemKinds.Weekday,
			Params: []int{6},
		},
		{
			Kind: LexemKinds.Except,
		},
		{
			Kind:   LexemKinds.Multiple,
			Params: []int{2, 1},
		},
		{
			Kind: LexemKinds.Date,
		},
		{
			Kind:   LexemKinds.Number,
			Params: []int{-1},
		},
	}

	err := test_query(query, expected)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLexemesOrder(t *testing.T) {
	query := "last Mondays, third from end Tuesdays, second Wednesdays"
	expected := []Lexem{
		{
			Kind:   LexemKinds.Order,
			Params: []int{-1},
		},
		{
			Kind:   LexemKinds.Weekday,
			Params: []int{1},
		},
		{
			Kind:   LexemKinds.Order,
			Params: []int{-3},
		},
		{
			Kind:   LexemKinds.Weekday,
			Params: []int{2},
		},
		{
			Kind:   LexemKinds.Order,
			Params: []int{2},
		},
		{
			Kind:   LexemKinds.Weekday,
			Params: []int{3},
		},
	}

	err := test_query(query, expected)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLexemesMultiple(t *testing.T) {
	query := "module 3 residue 1, 2 day"
	expected := []Lexem{
		{
			Kind:   LexemKinds.Multiple,
			Params: []int{3, 1, 2},
		},
		{
			Kind: LexemKinds.Day,
		},
	}

	err := test_query(query, expected)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLexemesShortMultiple(t *testing.T) {
	query := "module 2 day"
	expected := []Lexem{
		{
			Kind:   LexemKinds.Multiple,
			Params: []int{2, 0},
		},
		{
			Kind: LexemKinds.Day,
		},
	}

	err := test_query(query, expected)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLexemesErrorUnknown(t *testing.T) {
	query := "module 2 day, everyday"

	_, err := toLexemes(query)
	if err == nil {
		t.Fatal("should be incorrect")
	}

	if err.Error() != "unknown lexem 'everyday'" {
		t.Fatal("wrong err", err)
	}
}

func TestLexemesErrorNoModule(t *testing.T) {
	query := "module Fridays"

	_, err := toLexemes(query)
	if err == nil {
		t.Fatal("should be incorrect")
	}
}

func TestLexemesErrorNoModuleTooShort(t *testing.T) {
	query := "even Fridays, module"

	_, err := toLexemes(query)
	if err == nil {
		t.Fatal("should be incorrect")
	}
}

func TestLexemesErrorWrongModule(t *testing.T) {
	query := "module 0 Fridays"

	_, err := toLexemes(query)
	if err == nil {
		t.Fatal("should be incorrect")
	}
}
func TestLexemesErrorNoResidue(t *testing.T) {
	query := "module 3 residue Fridays"

	_, err := toLexemes(query)
	if err == nil {
		t.Fatal("should be incorrect")
	}
}

func TestLexemesErrorResidue(t *testing.T) {
	query := "module 32 residue 1, 5, -3 Fridays"

	_, err := toLexemes(query)
	if err == nil {
		t.Fatal("should be incorrect")
	}
}

func TestLexemesErrorFromEnd(t *testing.T) {
	query := "fifth from ent Fridays"

	_, err := toLexemes(query)
	if err == nil {
		t.Fatal("should be incorrect")
	}
}

func test_query(query string, expected []Lexem) error {
	lexems, err := toLexemes(query)
	if err != nil {
		return err
	}

	if len(lexems) != len(expected) {
		return errors.New(fmt.Sprintf("wrong length: %v, %v", lexems, expected))
	}

	for i := range expected {
		if !equal_lexem(&expected[i], &lexems[i]) {

			return errors.New(fmt.Sprintf("wrong lexem, expected: %v, actual: %v", expected[i], lexems[i]))
		}
	}

	return nil
}

func equal_lexem(a *Lexem, b *Lexem) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if a.Kind != b.Kind {
		return false
	}

	if len(a.Params) != len(b.Params) {
		return false
	}

	for i := range a.Params {
		if a.Params[i] != b.Params[i] {
			return false
		}
	}

	return true
}

// check syntax
func TestSyntaxList(t *testing.T) {
	query := "Tuesdays, Thursdays, Saturdays except odd date, -1"
	include := []Timeslot{
		&EveryTimeslot{
			Weekday: 2,
		},
		&EveryTimeslot{
			Weekday: 4,
		},
		&EveryTimeslot{
			Weekday: 6,
		},
	}
	exclude := []Timeslot{
		&MultipleDailyTimeslot{
			Kind:    MultipleKinds.Date,
			Module:  2,
			Residue: []int{1},
		},
		&NumberTimeslot{
			Number: -1,
		},
	}

	err := test_syntax(query, include, exclude)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSyntaxWithExcept(t *testing.T) {
	query := "every day except Sundays"
	include := []Timeslot{
		&EverydayTimeslot{},
	}
	exclude := []Timeslot{
		&EveryTimeslot{
			Weekday: 0,
		},
	}

	err := test_syntax(query, include, exclude)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSyntaxGlobalModificators(t *testing.T) {
	query := "penultimate Fridays, Sundays"
	include := []Timeslot{
		&OrderTimeslot{
			Weekday: 5,
			Order:   -2,
		},
		&OrderTimeslot{
			Weekday: 0,
			Order:   -2,
		},
	}
	exclude := []Timeslot{}

	err := test_syntax(query, include, exclude)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSyntaxEveryModificator(t *testing.T) {
	query := "every day, every date"
	include := []Timeslot{
		&EverydayTimeslot{},
		&EverydayTimeslot{},
	}
	exclude := []Timeslot{}

	err := test_syntax(query, include, exclude)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSyntaxOrderModificator(t *testing.T) {
	query := "first Mondays"
	include := []Timeslot{
		&OrderTimeslot{
			Weekday: 1,
			Order:   1,
		},
	}
	exclude := []Timeslot{}

	err := test_syntax(query, include, exclude)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSyntaxErrorExceptEveryday(t *testing.T) {
	query := "Mondays except every day"

	lexems, err := toLexemes(query)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = toTimeslots(lexems)
	if err == nil {
		t.Fatal("should be incorrect")
	}
}

func TestSyntaxModuleModificator(t *testing.T) {
	query := "module 3 residue 0, 1 day, even Sundays"
	include := []Timeslot{
		&MultipleDailyTimeslot{
			Kind:    MultipleKinds.Day,
			Module:  3,
			Residue: []int{0, 1},
		},
		&MultipleWeekdayTimeslot{
			Weekday: 0,
			Module:  2,
			Residue: []int{0},
		},
	}
	exclude := []Timeslot{}

	err := test_syntax(query, include, exclude)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSyntaxErrorExceptEverydate(t *testing.T) {
	query := "Mondays except every date"

	lexems, err := toLexemes(query)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = toTimeslots(lexems)
	if err == nil {
		t.Fatal("should be incorrect")
	}
}

func TestSyntaxErrorMoreThanOneExcept(t *testing.T) {
	query := "Mondays except Tuesdays except Wednesdays"

	lexems, err := toLexemes(query)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = toTimeslots(lexems)
	if err == nil {
		t.Fatal("should be incorrect")
	}
}
func test_syntax(query string, expected_include []Timeslot, expected_exclude []Timeslot) error {
	lexems, err := toLexemes(query)
	if err != nil {
		return err
	}

	include, exclude, err := toTimeslots(lexems)
	if err != nil {
		return err
	}

	if len(include) != len(expected_include) {
		return errors.New(fmt.Sprintf("wrong include length: %v, %v", include, expected_include))
	}
	if len(exclude) != len(expected_exclude) {
		return errors.New(fmt.Sprintf("wrong include length: %v, %v", exclude, expected_exclude))
	}

	for i := range include {
		if !equal_timeslot(include[i], expected_include[i]) {
			return errors.New(fmt.Sprintf("wrong timeslot: %v, %v", include[i], expected_include[i]))
		}
	}

	for i := range exclude {
		if !equal_timeslot(exclude[i], expected_exclude[i]) {
			return errors.New(fmt.Sprintf("wrong timeslot: %v, %v", exclude[i], expected_exclude[i]))
		}
	}

	return nil
}

func equal_timeslot(a Timeslot, b Timeslot) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	switch a_slot := a.(type) {
	case *EverydayTimeslot:
		_, ok := b.(*EverydayTimeslot)
		if !ok {
			return false
		}

	case *EveryTimeslot:
		b_slot, ok := b.(*EveryTimeslot)
		if !ok {
			return false
		}
		if a_slot.Weekday != b_slot.Weekday {
			return false
		}

	case *NumberTimeslot:
		b_slot, ok := b.(*NumberTimeslot)
		if !ok {
			return false
		}
		if a_slot.Number != b_slot.Number {
			return false
		}

	case *OrderTimeslot:
		b_slot, ok := b.(*OrderTimeslot)
		if !ok {
			return false
		}
		if a_slot.Weekday != b_slot.Weekday {
			return false
		}
		if a_slot.Order != b_slot.Order {
			return false
		}

	case *MultipleWeekdayTimeslot:
		b_slot, ok := b.(*MultipleWeekdayTimeslot)
		if !ok {
			return false
		}
		if a_slot.Weekday != b_slot.Weekday {
			return false
		}
		if a_slot.Module != b_slot.Module {
			return false
		}
		if len(a_slot.Residue) != len(b_slot.Residue) {
			return false
		}

		for i := range a_slot.Residue {
			if a_slot.Residue[i] != b_slot.Residue[i] {
				return false
			}
		}

	case *MultipleDailyTimeslot:
		b_slot, ok := b.(*MultipleDailyTimeslot)
		if !ok {
			return false
		}
		if a_slot.Kind != b_slot.Kind {
			return false
		}
		if a_slot.Module != b_slot.Module {
			return false
		}
		if len(a_slot.Residue) != len(b_slot.Residue) {
			return false
		}

		for i := range a_slot.Residue {
			if a_slot.Residue[i] != b_slot.Residue[i] {
				return false
			}
		}
	}

	return true
}

// check schedule

func TestTimeslotEveryday(t *testing.T) {
	begin := time.Date(2024, 3, 30, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 4, 2, 0, 0, 0, 0, time.UTC)

	timeslot := &EverydayTimeslot{}

	expected := []time.Time{
		time.Date(2024, 3, 30, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
	}

	dates := timeslot.Slots(begin, end)
	if !equal_dates(dates, expected) {
		t.Fatal("not equal timeslots", dates, expected)
	}
}

func TestTimeslotEvery(t *testing.T) {
	begin := time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 4, 5, 0, 0, 0, 0, time.UTC)

	timeslot := &EveryTimeslot{
		Weekday: 1,
	}

	expected := []time.Time{
		time.Date(2024, 3, 25, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
	}

	dates := timeslot.Slots(begin, end)
	if !equal_dates(dates, expected) {
		t.Fatal("not equal timeslots", dates, expected)
	}
}

func TestTimeslotNumberPositive(t *testing.T) {
	begin := time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC)

	timeslot := &NumberTimeslot{
		Number: 26,
	}

	expected := []time.Time{
		time.Date(2024, 3, 26, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 4, 26, 0, 0, 0, 0, time.UTC),
	}

	dates := timeslot.Slots(begin, end)
	if !equal_dates(dates, expected) {
		t.Fatal("not equal timeslots", dates, expected)
	}
}

func TestTimeslotNumberNegative(t *testing.T) {
	begin := time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)

	timeslot := &NumberTimeslot{
		Number: -1,
	}

	expected := []time.Time{
		time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
	}

	dates := timeslot.Slots(begin, end)
	if !equal_dates(dates, expected) {
		t.Fatal("not equal timeslots", dates, expected)
	}
}

func TestTimeslotOrderPositive(t *testing.T) {
	begin := time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)

	timeslot := &OrderTimeslot{
		Weekday: 0,
		Order:   2,
	}

	expected := []time.Time{
		time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 4, 14, 0, 0, 0, 0, time.UTC),
	}

	dates := timeslot.Slots(begin, end)
	if !equal_dates(dates, expected) {
		t.Fatal("not equal timeslots", dates, expected)
	}
}

func TestTimeslotOrderNegative(t *testing.T) {
	begin := time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)

	timeslot := &OrderTimeslot{
		Weekday: 5,
		Order:   -1,
	}

	expected := []time.Time{
		time.Date(2024, 2, 23, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 29, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 4, 26, 0, 0, 0, 0, time.UTC),
	}

	dates := timeslot.Slots(begin, end)
	if !equal_dates(dates, expected) {
		t.Fatal("not equal timeslots", dates, expected)
	}
}

func TestTimeslotOrderOutOfRange1(t *testing.T) {
	begin := time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)

	timeslot := &OrderTimeslot{
		Weekday: 4,
		Order:   5,
	}

	expected := []time.Time{
		time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
	}

	dates := timeslot.Slots(begin, end)
	if !equal_dates(dates, expected) {
		t.Fatal("not equal timeslots", dates, expected)
	}
}

func TestTimeslotOrderOutOfRange2(t *testing.T) {
	begin := time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)

	timeslot := &OrderTimeslot{
		Weekday: 0,
		Order:   6,
	}

	expected := []time.Time{}

	dates := timeslot.Slots(begin, end)
	if !equal_dates(dates, expected) {
		t.Fatal("not equal timeslots", dates, expected)
	}
}

func TestTimeslotOrderOutOfRange3(t *testing.T) {
	begin := time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)

	timeslot := &OrderTimeslot{
		Weekday: 0,
		Order:   6,
	}

	expected := []time.Time{}

	dates := timeslot.Slots(begin, end)
	if !equal_dates(dates, expected) {
		t.Fatal("not equal timeslots", dates, expected)
	}
}

func TestTimeslotMultipleWeekday1(t *testing.T) {
	begin := time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)

	timeslot := &MultipleWeekdayTimeslot{
		Weekday: 5,
		Module:  2,
		Residue: []int{1},
	}

	expected := []time.Time{
		time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 29, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 4, 12, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 4, 26, 0, 0, 0, 0, time.UTC),
	}

	dates := timeslot.Slots(begin, end)
	if !equal_dates(dates, expected) {
		t.Fatal("not equal timeslots", dates, expected)
	}
}

func TestTimeslotMultipleWeekday2(t *testing.T) {
	begin := time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)

	timeslot := &MultipleWeekdayTimeslot{
		Weekday: 5,
		Module:  5,
		Residue: []int{1, 2, 4},
	}

	expected := []time.Time{
		time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 22, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 4, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 4, 19, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 4, 26, 0, 0, 0, 0, time.UTC),
	}

	dates := timeslot.Slots(begin, end)
	if !equal_dates(dates, expected) {
		t.Fatal("not equal timeslots", dates, expected)
	}
}

func TestTimeslotMultipleDate(t *testing.T) {
	begin := time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 7, 0, 0, 0, 0, time.UTC)

	timeslot := &MultipleDailyTimeslot{
		Kind:    MultipleKinds.Date,
		Module:  2,
		Residue: []int{0},
	}

	expected := []time.Time{
		time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 4, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 6, 0, 0, 0, 0, time.UTC),
	}

	dates := timeslot.Slots(begin, end)
	if !equal_dates(dates, expected) {
		t.Fatal("not equal timeslots", dates, expected)
	}
}

func TestTimeslotMultipleDay(t *testing.T) {
	begin := time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 7, 0, 0, 0, 0, time.UTC)

	timeslot := &MultipleDailyTimeslot{
		Kind:    MultipleKinds.Day,
		Module:  2,
		Residue: []int{0},
	}

	expected := []time.Time{
		time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 4, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 6, 0, 0, 0, 0, time.UTC),
	}

	dates := timeslot.Slots(begin, end)
	if !equal_dates(dates, expected) {
		t.Fatal("not equal timeslots", dates, expected)
	}
}

func equal_dates(a []time.Time, b []time.Time) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		a_y, a_m, a_d := a[i].Date()
		b_y, b_m, b_d := b[i].Date()

		if a_y != b_y || a_m != b_m || a_d != b_d {
			return false
		}
	}

	return true
}

// check schedule

func TestScheduleEasy(t *testing.T) {
	query := "every day"
	times := []time.Time{
		time.Date(1, 1, 1, 11, 0, 0, 0, time.UTC),
		time.Date(1, 1, 1, 13, 0, 0, 0, time.UTC),
	}
	begin := time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 2, 0, 0, 0, 0, time.UTC)

	expected := []time.Time{
		time.Date(2024, 2, 28, 11, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 28, 13, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 29, 11, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 29, 13, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 1, 11, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 1, 13, 0, 0, 0, time.UTC),
	}

	schedule, err := Calculate(query, times, begin, end)
	if err != nil {
		t.Fatal(err)
	}

	if !equal_schedule(schedule, expected) {
		t.Fatal("not equal schedule", schedule, expected)
	}
}

func TestScheduleEasyWithBorders(t *testing.T) {
	query := "every day"
	times := []time.Time{
		time.Date(1, 1, 1, 11, 0, 0, 0, time.UTC),
		time.Date(1, 1, 1, 13, 0, 0, 0, time.UTC),
	}
	begin := time.Date(2024, 2, 28, 12, 37, 0, 0, time.UTC)
	end := time.Date(2024, 3, 2, 15, 44, 0, 0, time.UTC)

	expected := []time.Time{
		time.Date(2024, 2, 28, 13, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 29, 11, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 29, 13, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 1, 11, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 1, 13, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 2, 11, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 2, 13, 0, 0, 0, time.UTC),
	}

	schedule, err := Calculate(query, times, begin, end)
	if err != nil {
		t.Fatal(err)
	}

	if !equal_schedule(schedule, expected) {
		t.Fatal("not equal schedule", schedule, expected)
	}
}

func TestScheduleExcept(t *testing.T) {
	query := "every day except Sundays"
	times := []time.Time{
		time.Date(1, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	begin := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 6, 0, 0, 0, 0, time.UTC)

	expected := []time.Time{
		time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 2, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 4, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 5, 12, 0, 0, 0, time.UTC),
	}

	schedule, err := Calculate(query, times, begin, end)
	if err != nil {
		t.Fatal(err)
	}

	if !equal_schedule(schedule, expected) {
		t.Fatal("not equal schedule", schedule, expected)
	}
}

func TestScheduleMerge(t *testing.T) {
	query := "even date, 2"
	times := []time.Time{
		time.Date(1, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	begin := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 6, 0, 0, 0, 0, time.UTC)

	expected := []time.Time{
		time.Date(2024, 3, 2, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 4, 12, 0, 0, 0, time.UTC),
	}

	schedule, err := Calculate(query, times, begin, end)
	if err != nil {
		t.Fatal(err)
	}

	if !equal_schedule(schedule, expected) {
		t.Fatal("not equal schedule", schedule, expected)
	}
}

func TestScheduleComplex(t *testing.T) {
	query := "4, -1, module 5 residue 2 day, first Tuesdays, even Thursdays, every Saturdays except odd date, -5"
	times := []time.Time{
		time.Date(1, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	begin := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)

	expected := []time.Time{
		time.Date(2024, 2, 4, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 6, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 8, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 10, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 16, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 22, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 24, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 26, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 2, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 4, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 12, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 16, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 22, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 30, 12, 0, 0, 0, time.UTC),
	}

	schedule, err := Calculate(query, times, begin, end)
	if err != nil {
		t.Fatal(err)
	}

	if !equal_schedule(schedule, expected) {
		t.Fatal("not equal schedule", schedule, expected)
	}
}

func equal_schedule(a []time.Time, b []time.Time) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !a[i].Equal(b[i]) {
			return false
		}
	}

	return true
}
