package ask

import (
	"strings"

	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/hori-ryota/zaperr"
	"github.com/leporo/sqlf"
	"go.uber.org/zap"
)

type OngoingPoll struct {
	Role string `db:"role"`
	Post int    `db:"post"`
}

const (
	PollAnswerNone   = 0
	PollAnswerIgnore = -1
)

type PollCache struct {
	Answers []PollAnswer
}

// Decision is vk id of participants or PollAnswerNone or PollAnswerIgnore
type PollAnswer struct {
	PollID   int `db:"poll_id"`
	AnswerID int `db:"answer_id"`
	Decision int `db:"decision"`
}

func (a *Ask) OngoingPolls() ([]OngoingPoll, error) {
	var polls []OngoingPoll

	query := sqlf.From("ongoing_polls").
		Bind(&OngoingPoll{})

	err := a.db.Select(&polls, query.String(), query.Args()...)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get ongoing polls",
			zap.String("query", query.String()),
			zap.Any("args", query.Args()))
	}

	return polls, nil
}

func ConvertAnswers(labels []string, decision []int, poll *object.PollsPoll) *PollCache {
	var cache []PollAnswer

	for i, l := range labels {
		for _, a := range poll.Answers {
			if strings.Compare(l, a.Text) == 0 {
				cache = append(cache, PollAnswer{
					PollID:   poll.ID,
					AnswerID: a.ID,
					Decision: decision[i],
				})
			}

		}
	}

	return &PollCache{
		Answers: cache,
	}

}

func (a *Ask) AddPoll(role string, cache *PollCache) error {
	query := sqlf.InsertInto("poll_answers_cache")

	for _, answer := range cache.Answers {
		query.NewRow().
			Set("poll_id", answer.PollID).
			Set("answer_id", answer.AnswerID).
			Set("decision", answer.Decision)
	}

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
