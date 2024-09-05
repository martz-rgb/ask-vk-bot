package watcher

import (
	"ask-bot/src/ask"
	"ask-bot/src/datatypes/functional"
	"ask-bot/src/datatypes/posts"
	ts "ask-bot/src/templates"
	"ask-bot/src/vk"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

func (c *Controls) CheckPendingPolls() error {
	polls := c.Postponed.PostsKind(posts.Kinds.Poll)

	// order by name
	pending_polls, err := c.Ask.PendingPolls()
	if err != nil {
		return err
	}

	delete := []posts.Post{}

	for _, poll := range polls {
		i, ok := slices.BinarySearchFunc(pending_polls, poll.Roles[0].Name, func(pp ask.PendingPoll, s string) int {
			return strings.Compare(pp.Name, s)
		})

		if !ok {
			delete = append(delete, poll)
			continue
		}

		pending_polls = append(pending_polls[:i], pending_polls[i+1:]...)
	}

	if len(delete) > 0 {
		err = c.Postponed.DeletePosts(c.PostponedControls(), delete)
		if err != nil {
			return err
		}
	}

	if len(pending_polls) == 0 {
		return nil
	}

	new := []vk.PostParams{}

	// when to post
	begin := time.Now()
	end := begin.Add(14 * 24 * time.Hour)

	slots, err := c.Ask.Schedule(ask.TimeslotKinds.Polls, begin, end)
	if err != nil {
		return err
	}

	slots.Exclude(c.Postponed.Schedule())

	// create polls
	// TO-DO maybe sort polls by timestamp or smth
	for i := 0; i < len(slots) && i < len(pending_polls); i++ {
		poll, err := c.createPoll(pending_polls[i], slots[i])
		if err != nil {
			return err
		}

		new = append(new, poll)
	}

	return c.Postponed.AddPosts(c.PostponedControls(), new)
}

func (c *Controls) createPoll(poll ask.PendingPoll, date time.Time) (vk.PostParams, error) {
	text, err := c.pollText(poll)
	if err != nil {
		return vk.PostParams{}, err
	}

	images, err := c.uploadGreetings(poll.Greetings)
	if err != nil {
		return vk.PostParams{}, err
	}

	vk_poll, err := c.createVkPoll(poll, date)
	if err != nil {
		return vk.PostParams{}, err
	}

	return vk.PostParams{
		Text:        text,
		Attachments: append(images, vk_poll),
		PublishDate: date,
	}, nil
}

func (c *Controls) pollText(poll ask.PendingPoll) (string, error) {
	return ts.ParseTemplate(
		ts.PostPoll,
		ts.PostPollData{
			PollHashtag: c.Ask.OrganizationHashtags().PollHashtag,
			Poll:        poll,
		},
	)
}

// TO-DO not sure if it is fast way, maybe some kind of bulk upload is useful
func (c *Controls) uploadGreetings(greetings ask.Greetings) ([]string, error) {
	var attachments []string

	for _, greeting := range greetings {
		for _, image := range greeting {
			file, err := http.Get(image)
			if err != nil {
				return nil, zaperr.Wrap(err, "failed to download image",
					zap.String("url", image))
			}

			// TO-DO admin or group?
			photos, err := c.Admin.UploadPhotoToWall(file.Body)
			if err != nil {
				return nil, err
			}
			// TO-DO: more normal way to convert
			for _, photo := range photos {
				attachments = append(attachments,
					fmt.Sprintf("photo%d_%d_%s", photo.OwnerID, photo.ID, photo.AccessKey))
			}
		}
	}

	return attachments, nil
}

func (c *Controls) createVkPoll(poll ask.PendingPoll, date time.Time) (string, error) {
	label, err := ts.ParseTemplate(
		ts.PostPollLabel,
		ts.PostPollLabelData{},
	)
	if err != nil {
		return "", err
	}

	answers := poll.Answers()

	for i := range answers {
		answer, err := ts.ParseTemplate(
			ts.PostPollAnswer,
			ts.PostPollAnswerData{
				Index: i,
				Count: len(answers),
				Value: answers[i].Value,
			},
		)

		if err != nil {
			return "", err
		}

		answers[i].Label = answer
	}

	// add config for poll duration
	vk_poll, err := c.Admin.CreatePoll(label,
		functional.Map(answers, func(a ask.PollAnswer) string { return a.Label }),
		true,
		date.Add(24*time.Hour).Unix())

	if err != nil {
		return "", err
	}

	// add to db info about poll
	for _, vk_answer := range vk_poll.Answers {

		for j := range answers {
			if strings.Compare(answers[j].Label, vk_answer.Text) == 0 {
				answers[j].ID = vk_answer.ID
				break
			}
		}
	}

	err = c.Ask.SavePollAnswers(vk_poll.ID, answers)
	if err != nil {
		return "", err
	}

	return vk_poll.ToAttachment(), nil
}
