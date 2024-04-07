package listener

import (
	"ask-bot/src/ask"
	"ask-bot/src/postponed"
	"ask-bot/src/posts"
	"ask-bot/src/vk"
	"context"
	"sync"

	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/object"
	"go.uber.org/zap"
)

type Controls struct {
	Ask *ask.Ask

	Group *vk.VK
	Admin *vk.VK

	Postponed       *postponed.Postponed
	UpdatePostponed chan bool

	NotifyUser chan *vk.MessageParams
}

type Listener struct {
	c *Controls

	log *zap.SugaredLogger
}

func New(
	controls *Controls,
	log *zap.SugaredLogger) *Listener {
	return &Listener{
		c:   controls,
		log: log,
	}
}

func (l *Listener) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	wg.Add(1)
	go l.RunDB(ctx, wg)

	lp, err := l.c.Group.NewLongPoll()
	if err != nil {
		l.log.Errorw("failed to run listener longpoll",
			"error", err,
			"id", l.c.Group.ID())
		return
	}

	lp.WallPostNew(l.WallPostNew)

	lp.RunWithContext(ctx)
}

func (l *Listener) WallPostNew(ctx context.Context, event events.WallPostNewObject) {
	l.log.Info(event)

	vk_post := object.WallWallpost(event)

	err := l.NewPost(&vk_post)
	if err != nil {
		l.log.Errorw("failed to handle new post",
			"post id", event.ID,
			"error", err)
	}

	//l.admin.WallPostNew(l.group_id, "got: "+event.Text, "", false, time.Now().Add(5*time.Minute).Unix())
}

func (l *Listener) NewPost(vk_post *object.WallWallpost) error {
	// i don't care about "copy", "reply" and "postponed"
	// the last two shouldn't go here anyway
	if vk_post.PostType != vk.SuggestedPost &&
		vk_post.PostType != vk.NewPost {
		return nil
	}

	dictionary, err := l.c.Ask.RolesDictionary()
	if err != nil {
		return err
	}
	organization := l.c.Ask.OrganizationHashtags()

	post := posts.Parse(vk_post, dictionary, &organization)

	switch post.Kind {
	case posts.Poll:
		if vk_post.PostType == vk.SuggestedPost {
			// it is wrong
			// should delete it probably
			break
		}

		err := l.c.Ask.AddOngoingPoll(post.ID, post.Roles[0].Name)
		if err != nil {
			return err
		}
	case posts.Answer:
	case posts.FreeAnswer:
	case posts.Leaving:
	case posts.Invalid:
	}

	return nil
}
