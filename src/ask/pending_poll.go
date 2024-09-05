package ask

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/hori-ryota/zaperr"
	"github.com/leporo/sqlf"
	"go.uber.org/zap"
)

type VkIDs []int

func (ids VkIDs) Value() (driver.Value, error) {
	return []int(ids), nil
}

func (ids *VkIDs) Scan(value interface{}) error {
	if value == nil {
		return errors.New("VkIDs is not nullable")
	}

	if str, err := driver.String.ConvertValue(value); err == nil {
		if v, ok := str.(string); ok {
			strs := strings.Split(v, ",")
			var ints []int

			for _, s := range strs {
				i, err := strconv.Atoi(s)
				if err != nil {
					return err
				}

				ints = append(ints, i)
			}

			*ids = VkIDs(ints)
			return nil
		}

	}
	return errors.New("failed to scan VkIDs")
}

type Greetings map[uint32]Urls

func (p Greetings) Value() (driver.Value, error) {
	json, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	return json, nil
}

func (p *Greetings) Scan(value interface{}) error {
	if value == nil {
		return errors.New("Greetings is not nullable")
	}

	if str, err := driver.String.ConvertValue(value); err == nil {
		if v, ok := str.(string); ok {
			var participants Greetings
			err := json.Unmarshal([]byte(v), &participants)
			if err != nil {
				return err
			}

			*p = participants
			return nil
		}
	}

	return errors.New("failed to scan Participants")
}

type PendingPoll struct {
	Role

	Count        int       `db:"count"`
	Participants VkIDs     `db:"participants"`
	Greetings    Greetings `db:"greetings"`
}

func (a *Ask) PendingPolls() ([]PendingPoll, error) {
	var polls []PendingPoll

	query := sqlf.From("pending_polls_details").
		Bind(&PendingPoll{}).
		OrderBy("name")

	err := a.db.Select(&polls, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get pending polls",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return polls, nil
}

// value -
// 1) vk id
// 2) -1 as no/no one
// 3) -2 as neutral (really want it a 0 but i should check for undefined value)

// ID is for vk poll answer id
type PollAnswer struct {
	Label string
	ID    int
	Value int
}

// return answers with values
func (poll *PendingPoll) Answers() []PollAnswer {
	// TO-DO plus if neutral variant
	answers := make([]PollAnswer, poll.Count+1)

	for i := range poll.Participants {
		answers[i].Value = poll.Participants[i]
	}

	answers[len(answers)-1].Value = -1

	return answers
}

type PollAnswerCache struct {
	PollID   int `db:"poll_id"`
	AnswerID int `db:"answer_id"`
	Value    int `db:"value"`
}

func (a *Ask) SavePollAnswers(poll_id int, answers []PollAnswer) error {
	query := sqlf.InsertInto("poll_answer_cache")

	for _, answer := range answers {
		query.NewRow().
			Set("poll_id", poll_id).
			Set("answer_id", answer.ID).
			Set("value", answer.Value)
	}

	_, err := a.db.Exec(query.String(), query.Args()...)
	if err != nil {
		return zaperr.Wrap(err, "failed to save poll answers",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return nil
}

func (a *Ask) LoadPollAnswer(poll_id int, answer_id int) (int, error) {
	var value int

	query := sqlf.From("poll_answer_cache").
		Bind(&PollAnswerCache{}).
		Where("poll_id = ?", poll_id).
		Where("answer_id = ?", answer_id)

	err := a.db.Get(&value, query.String(), query.Args()...)
	if err != nil {
		return 0, zaperr.Wrap(err, "failed to load poll answer",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return value, nil
}
