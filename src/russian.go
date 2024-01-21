package main

import (
	"time"

	"github.com/rb-go/plural-ru"
)

func PluralNoun(oneWord, twoWord, threeWord string) func(int) string {
	return func(count int) string {
		return plural.Noun(count, oneWord, twoWord, threeWord)
	}
}

func MonthGenitive(month time.Month) string {
	switch month {
	case time.January:
		return "января"
	case time.February:
		return "февраля"
	case time.March:
		return "марта"
	case time.April:
		return "апреля"
	case time.May:
		return "мая"
	case time.June:
		return "июня"
	case time.July:
		return "июля"
	case time.August:
		return "августа"
	case time.September:
		return "сентября"
	case time.October:
		return "октября"
	case time.November:
		return "ноября"
	case time.December:
		return "декабря"
	}

	return ""
}
