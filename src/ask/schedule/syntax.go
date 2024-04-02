package schedule

import "errors"

// TO-DO: анализ модификаторов
// не должно быть два модификатора рядом (сейчас это не ошибка),
// за исключением модификаторов порядка, которые склеиваются (только те, что идут подряд)
func toTimeslots(lexems []Lexem) ([]Timeslot, []Timeslot, error) {
	include := []Timeslot{}
	exclude := []Timeslot{}

	var slots *[]Timeslot = &include

	var modifier *Lexem

	for index, lexem := range lexems {
		switch lexem.Kind {
		case LexemKinds.Number:
			modifier = nil

			*slots = append(*slots, &NumberTimeslot{
				Number: lexem.Params[0],
			})

		case LexemKinds.Day, LexemKinds.Date:
			if modifier == nil || modifier.Kind != LexemKinds.Multiple {
				modifier = nil

				if slots == &exclude {
					return nil, nil, errors.New("every day can not be excluded timeslot")
				}
				*slots = append(*slots, &EverydayTimeslot{})

				continue
			}

			kind := MultipleKinds.Day
			if lexem.Kind == LexemKinds.Date {
				kind = MultipleKinds.Date
			}

			*slots = append(*slots, &MultipleDailyTimeslot{
				Kind: kind,

				Module:  modifier.Params[0],
				Residue: modifier.Params[1:],
			})

		case LexemKinds.Weekday:
			if modifier == nil {
				// every weekday
				*slots = append(*slots, &EveryTimeslot{
					Weekday: lexem.Params[0],
				})

				continue
			}

			switch modifier.Kind {
			case LexemKinds.Order:
				*slots = append(*slots, &OrderTimeslot{
					Weekday: lexem.Params[0],
					Order:   modifier.Params[0],
				})

			case LexemKinds.Multiple:
				*slots = append(*slots, &MultipleWeekdayTimeslot{
					Weekday: lexem.Params[0],
					Module:  modifier.Params[0],
					Residue: modifier.Params[1:],
				})
			}

		case LexemKinds.All:
			modifier = nil

		case LexemKinds.Order, LexemKinds.Multiple:
			modifier = &lexems[index]

		case LexemKinds.Except:
			modifier = nil
			if slots == &include {
				slots = &exclude
			} else {
				return nil, nil, errors.New("only one except is permitted")
			}
		}
	}

	return include, exclude, nil
}
