package ask

import (
	"database/sql"

	"github.com/hori-ryota/zaperr"
	"github.com/leporo/sqlf"
	"go.uber.org/zap"
)

// type PollAnswers map[int]int

// func (answers PollAnswers) Value() (driver.Value, error) {
// 	pairs := make([]string, len(answers))
// 	for key, value := range answers {
// 		pairs = append(pairs, fmt.Sprintf("%d,%d", key, value))
// 	}

// 	return strings.Join(pairs, ";"), nil
// }

// func (answers *PollAnswers) Scan(value interface{}) error {
// 	if value == nil {
// 		return errors.New("PollAnswers is not nullable")
// 	}

// 	if str, err := driver.String.ConvertValue(value); err == nil {
// 		if v, ok := str.(string); ok {
// 			strs := strings.Split(v, ";")

// 			pairs := make(map[int]int, len(strs))

// 			for _, s := range strs {
// 				pair := strings.Split(s, ",")
// 				if len(pair) != 2 {
// 					return errors.New("should be pair")
// 				}

// 				key, err := strconv.Atoi(pair[0])
// 				if err != nil {
// 					return err
// 				}
// 				value, err := strconv.Atoi(pair[1])
// 				if err != nil {
// 					return err
// 				}

// 				pairs[key] = value
// 			}

// 			*answers = PollAnswers(pairs)
// 			return nil
// 		}

// 	}
// 	return errors.New("failed to scan PollAnswers")
// }

type Poll struct {
	Role string `db:"role"`
	// Answers PollAnswers   `db:"answers"`
	Post sql.NullInt32 `db:"post"`
}

func (a *Ask) Polls() ([]Poll, error) {
	var polls []Poll

	query := sqlf.From("polls").
		Bind(&Poll{})

	err := a.db.Select(&polls, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get polls",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return polls, nil
}

func (a *Ask) OngoingPolls() ([]Poll, error) {
	var polls []Poll

	query := sqlf.From("ongoing_polls").
		Bind(&Poll{})

	err := a.db.Select(&polls, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get ongoing polls",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return polls, nil
}

func (a *Ask) AddPoll(role string) error {
	query := sqlf.InsertInto("polls").
		Set("role", role).
		//Set("answers", answers).
		Clause("ON CONFLICT (role) DO UPDATE SET answers = excluded.answers")

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to add poll",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return nil
}

func (a *Ask) AddOngoingPoll(role string, post int) error {
	query := sqlf.Update("ongoing_polls").
		Set("post", post).
		Where("role = ?", role)

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to add ongoing poll",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return nil
}
