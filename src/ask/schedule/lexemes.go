package schedule

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type LexemKind int

var LexemKinds = struct {
	Unknown  LexemKind
	Number   LexemKind
	Day      LexemKind
	Date     LexemKind
	Weekday  LexemKind
	All      LexemKind
	Order    LexemKind
	Multiple LexemKind
	Except   LexemKind
}{
	Unknown:  0,
	Number:   1, // 1 params -- date
	Day:      2, // no params
	Date:     3, // no params
	Weekday:  4, // 1 params -- index of day
	All:      5, // no params
	Order:    6, // 1 params -- index
	Multiple: 7, // 2+ params -- module & residues
	Except:   8, // no params
}

const (
	Day  = "day"
	Date = "date"

	Mondays    = "mondays"
	Tuesdays   = "tuesdays"
	Wednesdays = "wednesdays"
	Thursdays  = "thursdays"
	Fridays    = "fridays"
	Saturdays  = "saturdays"
	Sundays    = "sundays"

	Every = "every"

	First       = "first"
	Second      = "second"
	Third       = "third"
	Forth       = "forth"
	Fifth       = "fifth"
	Last        = "last"
	Penultimate = "penultimate"

	From = "from"
	End  = "end"

	Odd     = "odd"
	Even    = "even"
	Module  = "module"
	Residue = "residue"

	Except = "except"
)

type Lexem struct {
	Kind   LexemKind
	Params []int
}

func toLexemes(sentence string) ([]Lexem, error) {
	// normalize
	sentence = strings.ToLower(sentence)
	// remove commas
	sentence = strings.ReplaceAll(sentence, ",", "")
	words := strings.Split(sentence, " ")

	lexems := []Lexem{}

	index := 0
	for index < len(words) {
		kind := wordKind(words[index])

		switch kind {
		case LexemKinds.Unknown:
			return nil, errors.New(fmt.Sprintf("unknown lexem '%s'", words[index]))

		case LexemKinds.Number:
			num, _ := strconv.Atoi(words[index])
			lexems = append(lexems, Lexem{
				Kind:   kind,
				Params: []int{num},
			})
			index++

		case LexemKinds.Weekday:
			lexems = append(lexems,
				Lexem{
					Kind:   kind,
					Params: []int{dayOfWeekIndex(words[index])},
				})
			index++

		case LexemKinds.Day, LexemKinds.Date, LexemKinds.All, LexemKinds.Except:
			lexems = append(lexems, Lexem{
				Kind: kind,
			})
			index++

		case LexemKinds.Order:
			if words[index] == Last {
				lexems = append(lexems, Lexem{
					Kind:   kind,
					Params: []int{-1},
				})
				index++
				continue
			}

			if words[index] == Penultimate {
				lexems = append(lexems, Lexem{
					Kind:   kind,
					Params: []int{-2},
				})
				index++
				continue
			}

			order := orderIndex(words[index])

			if index+1 < len(words) && words[index+1] == From {
				if index+2 < len(words) && words[index+2] == End {
					order = -order
					index += 2
				} else {
					return nil, errors.New("after keyword 'from' should go keyword 'end'")
				}
			}

			lexems = append(lexems, Lexem{
				Kind:   kind,
				Params: []int{order},
			})
			index++
			continue

		case LexemKinds.Multiple:
			if words[index] == Even {
				lexems = append(lexems, Lexem{
					Kind:   kind,
					Params: []int{2, 0},
				})
				index++
				continue
			}

			if words[index] == Odd {
				lexems = append(lexems, Lexem{
					Kind:   kind,
					Params: []int{2, 1},
				})
				index++
				continue
			}

			if index+1 >= len(words) {
				return nil, errors.New("no molule for multiple modifier")
			}

			module, err := strconv.Atoi(words[index+1])
			if err != nil {
				return nil, err
			}

			if module < 2 {
				return nil, errors.New(fmt.Sprintf("wrong module: %d", module))
			}

			residue := []int{}

			if index+2 < len(words) && words[index+2] == Residue {
				for index+3+len(residue) < len(words) {
					r, err := strconv.Atoi(words[index+3+len(residue)])
					if err != nil {
						break
					}

					if r < 0 || r >= module {
						return nil, errors.New(fmt.Sprintf("residue by module %d is incorrect: %d", module, r))
					}

					residue = append(residue, r)
				}

				if len(residue) == 0 {
					return nil, errors.New("no residue for multiple modifier with residue")
				}

				index += 1 + len(residue)
			} else {
				residue = []int{0}
			}

			lexems = append(lexems, Lexem{
				Kind:   kind,
				Params: append([]int{module}, residue...),
			})
			index += 2
			continue
		}
	}

	return lexems, nil
}

func wordKind(word string) LexemKind {
	switch word {
	case Day:
		return LexemKinds.Day
	case Date:
		return LexemKinds.Date
	case Mondays, Tuesdays, Wednesdays, Thursdays, Fridays, Saturdays, Sundays:
		return LexemKinds.Weekday
	case Every:
		return LexemKinds.All
	case First, Second, Third, Forth, Fifth, Last, Penultimate:
		return LexemKinds.Order
	case Odd, Even, Module:
		return LexemKinds.Multiple
	case Except:
		return LexemKinds.Except
	}

	_, err := strconv.Atoi(word)
	if err == nil {
		return LexemKinds.Number
	}

	return LexemKinds.Unknown
}

// respectfully to Weekday from time package
func dayOfWeekIndex(word string) int {
	switch word {
	case Sundays:
		return 0
	case Mondays:
		return 1
	case Tuesdays:
		return 2
	case Wednesdays:
		return 3
	case Thursdays:
		return 4
	case Fridays:
		return 5
	case Saturdays:
		return 6
	}

	return -1
}

func orderIndex(word string) int {
	switch word {
	case First:
		return 1
	case Second:
		return 2
	case Third:
		return 3
	case Forth:
		return 4
	case Fifth:
		return 5
	}

	return 0
}
