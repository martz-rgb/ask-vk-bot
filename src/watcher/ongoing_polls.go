package watcher

import (
	"ask-bot/src/ask"
	"time"

	"github.com/SevereCloud/vksdk/v2/object"
)

func (c *Controls) CheckOngoingPolls() error {
	// get post ids
	polls, err := c.Ask.OngoingPolls()
	if err != nil {
		return err
	}

	if len(polls) == 0 {
		return nil
	}

	// get vk posts
	ids := make([]int, len(polls))
	for i := range polls {
		ids[i] = int(polls[i].Post)
	}

	posts, err := c.Admin.PostsByIds(ids)
	if err != nil {
		return err
	}

	// check polls
	for i := range posts {
		for _, attachment := range posts[i].Attachments {
			if attachment.Type != object.AttachmentTypePoll {
				continue
			}

			if time.Now().After(time.Unix(int64(attachment.Poll.EndDate), 0)) {
				err := c.endPoll(&posts[i], &polls[i])
				if err != nil {
					return err
				}
			}
		}
	}

	// if finished then create greeting
	return nil
}

// TO-DO: finish
func (c *Controls) endPoll(post *object.WallWallpost, poll *ask.OngoingPoll) error {
	return nil
}
