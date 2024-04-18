package postponed

import (
	"ask-bot/src/ask"
	"ask-bot/src/posts"
	"ask-bot/src/schedule"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type CachePolls struct {
	existing []posts.Post

	new    []ask.Poll
	remove []posts.Post
}

func NewCachePolls(db_info *DBInfo, vk_info *VKInfo) *CachePolls {
	var result []posts.Post

	// sorting
	new := slices.Clone(db_info.polls)
	slices.SortFunc(new, func(a, b ask.Poll) int {
		return strings.Compare(a.Name, b.Name)
	})

	remove := vk_info.invalid
	for i := range vk_info.polls {
		index, ok := slices.BinarySearchFunc(new, vk_info.polls[i].Roles[0].Name, func(p ask.Poll, s string) int {
			return strings.Compare(p.Name, s)
		})

		if !ok {
			remove = append(remove, vk_info.polls[i])
		} else {
			new = append(new[:index], new[index+1:]...)
			result = append(result, vk_info.polls[i])
		}
	}

	return &CachePolls{
		existing: result,
		new:      new,
		remove:   remove,
	}
}

func (polls *CachePolls) AddNew(c *Controls, organization *ask.OrganizationHashtags, busy *([]time.Time)) (err error) {
	if len(polls.new) == 0 {
		return nil
	}

	begin := time.Now()
	end := begin

	slots := schedule.Schedule{}
	// 14, 70 are heurustic values
	for len(slots) < len(polls.new) && end.YearDay()-begin.YearDay() < 70 {
		end = end.AddDate(0, 0, 14)

		slots, err = c.Ask.Schedule(ask.TimeslotKinds.Polls, begin, end)
		if err != nil {
			return err
		}

		slots = slots.Exclude(*busy)
	}

	// if too few slots, just create some and left others on later
	for i := 0; i < len(polls.new) && i < len(slots); i++ {
		post, err := addNewPoll(c, organization, &polls.new[i], (slots)[i])
		if err != nil {
			return err
		}

		//add to existing
		polls.existing = append(polls.existing, *post)

		// add to busy
		index, _ := slices.BinarySearchFunc(*busy, slots[i], func(t1, t2 time.Time) int {
			return t1.Compare(t2)
		})
		if index == len(*busy) {
			*busy = append(*busy, slots[i])
		} else {
			*busy = append((*busy)[:index+1], (*busy)[index:]...)
			(*busy)[index] = slots[i]
		}
	}

	return nil
}

func (polls *CachePolls) Remove(c *Controls) error {
	for i := range polls.remove {
		err := c.Vk.DeletePost(polls.remove[i].ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (polls *CachePolls) Posts() []posts.Post {
	return polls.existing
}

func addNewPoll(c *Controls, organization *ask.OrganizationHashtags, poll *ask.Poll, date time.Time) (*posts.Post, error) {
	message := fmt.Sprintf("%s %s\nГолосование на %s!", organization.PollHashtag, poll.Hashtag, poll.Role.Name)

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

	// add config for poll duration
	vk_poll, err := c.Vk.CreatePoll("Берем?", []string{"Конечно!", "Нет."}, true, date.Add(24*time.Hour).Unix())
	if err != nil {
		return nil, err
	}
	attachments = append(attachments, fmt.Sprintf("poll%d_%d", vk_poll.OwnerID, vk_poll.ID))

	id, err := c.Vk.CreatePost(message, strings.Join(attachments, ","), false, date.Unix())
	if err != nil {
		return nil, err
	}

	return &posts.Post{
		Tags:  []string{organization.PollHashtag, poll.Hashtag},
		Kind:  posts.Poll,
		Roles: []ask.Role{poll.Role},
		ID:    id,
		Date:  date,
	}, nil
}
