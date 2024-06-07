package postponed

import (
	"ask-bot/src/ask"
	"ask-bot/src/posts"
	"ask-bot/src/schedule"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

func (c *Controls) createNew(objects *DBInfo, busy *schedule.Schedule) (new Dictionary, err error) {
	new = make(Dictionary)
	// poll, acceptances & leavings

	new[posts.Kinds.Poll], err = c.createNewPolls(objects.polls, busy)
	if err != nil {
		return nil, err
	}

	return new, nil
}

func (c *Controls) createNewPolls(polls []ask.Poll, busy *schedule.Schedule) (new []posts.Post, err error) {
	if len(polls) == 0 {
		return nil, nil
	}

	begin := time.Now()
	end := begin

	// 14, 70 are heurustic values
	slots := schedule.Schedule{}
	for len(slots) < len(polls) && end.YearDay()-begin.YearDay() < 70 {
		end = end.AddDate(0, 0, 14)

		slots, err = c.Ask.Schedule(ask.TimeslotKinds.Polls, begin, end)
		if err != nil {
			return nil, err
		}

		slots = slots.Exclude(*busy)
	}

	// if too few slots, just create some and left others on later
	for i := 0; i < len(polls) && i < len(slots); i++ {
		post, err := c.addNewPoll(polls[i], slots[i])
		if err != nil {
			return nil, err
		}

		//add to existing
		new = append(new, *post)

		// add to busy
		*busy = (*busy).Add(slots[i])
	}

	return new, nil
}

type PollDecision struct {
	Label    string
	Decision int
}

type PollDecisions []PollDecision

func (ds PollDecisions) ToLabels() []string {
	labels := make([]string, len(ds))

	for i := range ds {
		labels[i] = ds[i].Label
	}

	return labels
}

func (c *Controls) addNewPoll(poll ask.Poll, date time.Time) (*posts.Post, error) {
	organization := c.Ask.OrganizationHashtags()

	message := fmt.Sprintf("%s %s\nГолосование на %s!",
		organization.PollHashtag,
		poll.Hashtag,
		poll.Role.Name)

	var attachments []string
	greetings := strings.Split(poll.Greetings, ";")
	for _, greeting := range greetings {
		images := strings.Split(greeting, ",")

		for _, image := range images {
			file, err := http.Get(image)
			if err != nil {
				return nil, zaperr.Wrap(err, "failed to download image",
					zap.String("url", image))
			}

			photos, err := c.Vk.UploadPhotoToWall(file.Body)
			if err != nil {
				return nil, err
			}

			for _, photo := range photos {
				attachments = append(attachments,
					fmt.Sprintf("photo%d_%d_%s", photo.OwnerID, photo.ID, photo.AccessKey))
			}
		}
	}

	// TO-DO: write normal

	labels := []string{"Конечно", "Нет"}
	decisions := make([]int, len(poll.Participants)+1)
	for i := range poll.Participants {
		decisions[i] = poll.Participants[i]
	}
	decisions[len(decisions)-1] = ask.PollAnswerNone

	// add config for poll duration
	vk_poll, err := c.Vk.CreatePoll("Берем?",
		labels,
		true,
		date.Add(24*time.Hour).Unix())
	if err != nil {
		return nil, err
	}
	attachments = append(attachments,
		fmt.Sprintf("poll%d_%d", vk_poll.OwnerID, vk_poll.ID))

	id, err := c.Vk.CreatePost(message, strings.Join(attachments, ","), false, date.Unix())
	if err != nil {
		return nil, err
	}

	// save cache
	err = c.Ask.AddPoll(poll.Role.Name, ask.ConvertAnswers(labels, decisions, vk_poll))
	if err != nil {
		return nil, err
	}

	return &posts.Post{
		Kind:  posts.Kinds.Poll,
		Roles: []ask.Role{poll.Role},
		ID:    id,
		Date:  date,
		Poll: &posts.Poll{
			ID:      vk_poll.ID,
			Closed:  bool(vk_poll.Closed),
			Answers: vk_poll.Answers,
		},
	}, nil
}
