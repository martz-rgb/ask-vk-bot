package postponed

import (
	"ask-bot/src/ask"
	"ask-bot/src/ask/schedule"
	"ask-bot/src/posts"
	"ask-bot/src/vk"
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type Controls struct {
	Vk  *vk.VK
	Ask *ask.Ask
}

type Postponed struct {
	mu     *sync.Mutex
	notify chan bool

	c    *Controls
	tick time.Duration

	cache []posts.Post

	log *zap.SugaredLogger
}

func New(c *Controls, tick time.Duration, log *zap.SugaredLogger) (*Postponed, chan bool) {
	p := &Postponed{
		&sync.Mutex{},
		make(chan bool),
		c,
		tick,
		nil,
		log,
	}
	return p, p.notify
}

func (p *Postponed) Run(ctx context.Context, wg *sync.WaitGroup) {
	err := p.update()
	if err != nil {
		p.log.Errorw("failed to update postponed on ticker",
			"error", err)
	}

	ticker := time.NewTicker(p.tick)

	for {
		select {
		case <-ticker.C:
			err := p.update()
			if err != nil {
				p.log.Errorw("failed to update postponed on ticker",
					"error", err)
			}

		case <-p.notify:
			err := p.update()
			if err != nil {
				p.log.Errorw("failed to update postponed on notify",
					"error", err)
			}
		case <-ctx.Done():
			wg.Done()
			return
		}
	}
}

func (postponed *Postponed) update() error {
	postponed.mu.Lock()
	defer postponed.mu.Unlock()

	organization := postponed.c.Ask.OrganizationHashtags()

	dictionary, err := postponed.c.Ask.RolesDictionary()
	if err != nil {
		return err
	}

	// pending polls dictionary order by role
	db_polls, err := postponed.c.Ask.PendingPolls()
	if err != nil {
		return err
	}
	slices.SortFunc(db_polls, func(a, b ask.Poll) int {
		return strings.Compare(a.Role.Name, b.Role.Name)
	})

	// info about vk postponed posts
	vk_postponed, err := postponed.c.Vk.PostponedPosts()
	if err != nil {
		return err
	}
	postponed_posts := posts.ParseMany(vk_postponed, dictionary, &organization)
	postponed.cache = postponed_posts

	for _, post := range postponed_posts {
		switch post.Kind {
		case posts.Poll:
			if len(post.Roles) != 1 {
				postponed.log.Infow("found invalid poll",
					"poll", post)

				postponed.c.Vk.DeletePost(post.ID)
				if err != nil {
					return err
				}

				postponed.log.Infow("deleted invalid poll",
					"poll", post)
				continue
			}

			ok := postponed.markPoll(&db_polls, post.Roles[0].Name)
			if !ok {
				postponed.log.Infow("found unwanted poll",
					"poll", post)

				err = postponed.c.Vk.DeletePost(post.ID)
				if err != nil {
					return err
				}

				postponed.log.Infow("deleted unwanted poll",
					"poll", post)
			}
		}
	}

	// create polls which are new

	// find enough slots
	begin := time.Now()
	end := begin.AddDate(0, 0, 14)

	slots := []time.Time{}
	for len(slots) < len(db_polls) {
		slots, err = postponed.c.Ask.Schedule(ask.TimeslotKinds.Polls, begin, end)
		if err != nil {
			return err
		}
		slots = schedule.ExcludeSchedule(slots, posts.ToTime(postponed.cache))
		end = end.AddDate(0, 0, 14)
	}

	for i, poll := range db_polls {
		id, err := postponed.createPoll(&poll, slots[i], &organization)
		if err != nil {
			return err
		}

		// add to cache
		postponed.cache = append(postponed.cache, posts.Post{
			Tags:  []string{organization.PollHashtag, poll.Hashtag},
			Kind:  posts.Poll,
			Roles: []ask.Role{poll.Role},
			ID:    id,
			Date:  slots[i],
		})
	}

	return nil
}

func (postponed *Postponed) markPoll(polls *[]ask.Poll, role string) bool {
	index, ok := slices.BinarySearchFunc(*polls, role, func(poll ask.Poll, role string) int {
		return strings.Compare(poll.Role.Name, role)
	})

	if !ok {
		return false
	}

	*polls = append((*polls)[:index], (*polls)[index+1:]...)
	return true
}

func (postponed *Postponed) createPoll(poll *ask.Poll, date time.Time, organization *ask.OrganizationHashtags) (int, error) {
	//  create vk post

	postponed.log.Infow("creating new poll",
		"poll", poll)

	message := fmt.Sprintf("%s %s\nГолосование на %s!", organization.PollHashtag, poll.Hashtag, poll.Role.Name)

	var attachments []string

	greetings := strings.Split(poll.Greetings, ";")
	for _, greeting := range greetings {
		images := strings.Split(greeting, ",")

		for _, image := range images {
			file, err := http.Get(image)
			if err != nil {
				return 0, zaperr.Wrap(err, "failed to download image",
					zap.String("url", image))
			}

			photos, err := postponed.c.Vk.UploadPhotoToWall(file.Body)
			if err != nil {
				return 0, err
			}

			for _, photo := range photos {
				attachments = append(attachments,
					fmt.Sprintf("photo%d_%d_%s", photo.OwnerID, photo.ID, photo.AccessKey))
			}
		}
	}

	// add config for poll duration
	vk_poll, err := postponed.c.Vk.CreatePoll("Берем?", []string{"Конечно!", "Нет."}, true, date.Add(24*time.Hour).Unix())
	if err != nil {
		return 0, err
	}
	attachments = append(attachments, fmt.Sprintf("poll%d_%d", vk_poll.OwnerID, vk_poll.ID))

	id, err := postponed.c.Vk.CreatePost(message, strings.Join(attachments, ","), false, date.Unix())
	if err != nil {
		return 0, err
	}

	postponed.log.Infow("created new poll",
		"poll", poll,
		"id", id)

	return id, nil
}
