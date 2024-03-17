package postponed

import (
	"ask-bot/src/ask"
	"ask-bot/src/vk"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type Postponed struct {
	mu     *sync.Mutex
	notify chan bool

	group int
	vk    *vk.VK
	ask   *ask.Ask

	log *zap.SugaredLogger
}

func New(group int, tick time.Duration, v *vk.VK, a *ask.Ask, log *zap.SugaredLogger) (*Postponed, chan bool) {
	p := &Postponed{
		&sync.Mutex{},
		make(chan bool),

		group,
		v,
		a,

		log,
	}

	go p.loop(tick)

	return p, p.notify
}

func (p *Postponed) loop(tick time.Duration) {
	err := p.update()
	if err != nil {
		p.log.Errorw("failed to update postponed on ticker",
			"error", err)
	}

	ticker := time.NewTicker(tick)

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
		}
	}
}

func (p *Postponed) update() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	organization_tags := p.ask.OrganizationHashtags()

	roles_dictionary, err := p.ask.RolesDictionary()
	if err != nil {
		return err
	}

	// pending polls dictionary order by role
	db_polls, err := p.ask.PendingPolls()
	if err != nil {
		return err
	}
	slices.SortFunc[[]ask.PendingPoll](db_polls, func(a, b ask.PendingPoll) int {
		return strings.Compare(a.Role, b.Role)
	})

	// update info about vk postponed posts

	vk_postponed, err := p.vk.PostponedPosts(p.group)
	if err != nil {
		return err
	}
	postponed, err := Parse(roles_dictionary, &organization_tags, vk_postponed)
	if err != nil {
		return err
	}

	for _, post := range postponed {
		switch post.Kind {
		case Poll:
			if len(post.Roles) != 1 {
				p.log.Infow("found invalid poll",
					"poll", post)

				p.vk.DeletePost(p.group, post.Vk.ID)
				if err != nil {
					return err
				}

				p.log.Infow("deleted invalid poll",
					"poll", post)
				continue
			}

			ok := p.markPoll(&db_polls, post.Roles[0].Name)
			if !ok {
				p.log.Infow("found unwanted poll",
					"poll", post)

				err = p.vk.DeletePost(p.group, post.Vk.ID)
				if err != nil {
					return err
				}

				p.log.Infow("deleted unwanted poll",
					"poll", post)
			}
		}
	}

	// create polls which are new
	for _, poll := range db_polls {
		err = p.createPoll(&poll, &organization_tags)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Postponed) markPoll(polls *[]ask.PendingPoll, role string) bool {
	index, ok := slices.BinarySearchFunc(*polls, role, func(poll ask.PendingPoll, role string) int {
		return strings.Compare(poll.Role, role)
	})

	if !ok {
		return false
	}

	*polls = append((*polls)[:index], (*polls)[index+1:]...)
	return true
}

func (p *Postponed) createPoll(poll *ask.PendingPoll, organization_tags *ask.OrganizationHashtags) error {
	//  create vk post

	p.log.Infow("creating new poll",
		"poll", poll)

	message := fmt.Sprintf("%s %s\nГолосование на %s!", organization_tags.PollHashtag, poll.Hashtag, poll.Role)

	var attachments []string

	greetings := strings.Split(poll.Greetings, ";")
	for _, greeting := range greetings {
		images := strings.Split(greeting, ",")

		for _, image := range images {
			file, err := http.Get(image)
			if err != nil {
				return zaperr.Wrap(err, "failed to download image",
					zap.String("url", image))
			}

			photos, err := p.vk.UploadPhotoToWall(p.group, file.Body)
			if err != nil {
				return err
			}

			for _, photo := range photos {
				attachments = append(attachments,
					fmt.Sprintf("photo%d_%d_%s", photo.OwnerID, photo.ID, photo.AccessKey))
			}
		}
	}

	post_time := time.Now().Add(3 * time.Hour)

	// add config for poll duration
	vk_poll, err := p.vk.CreatePoll(p.group, "Берем?", []string{"Конечно!", "Нет."}, true, post_time.Add(24*time.Hour).Unix())
	if err != nil {
		return err
	}
	attachments = append(attachments, fmt.Sprintf("poll%d_%d", vk_poll.OwnerID, vk_poll.ID))

	id, err := p.vk.CreatePost(p.group, message, strings.Join(attachments, ","), false, post_time.Unix())
	if err != nil {
		return err
	}

	p.log.Infow("created new poll",
		"poll", poll,
		"id", id)

	return nil
}
