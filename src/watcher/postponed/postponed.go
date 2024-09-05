package postponed

import (
	"ask-bot/src/ask"
	"ask-bot/src/datatypes/posts"
	"ask-bot/src/datatypes/schedule"
	"ask-bot/src/vk"
	"sync"
)

type Controls struct {
	Vk  *vk.VK
	Ask *ask.Ask
}

// Layer of abstraction to work with postponed posts.
// All postponed posts should be added and deleted through Postponed instance.
type Postponed struct {
	mu sync.Mutex

	posts    posts.Posts
	schedule schedule.Schedule
}

func New(v *vk.VK) *Postponed {
	return &Postponed{}
}

// update posts & schedule

// full reupdate of data
func (p *Postponed) Update(c *Controls) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	postponed, err := c.Vk.PostponedPosts()
	if err != nil {
		return err
	}

	dictionary, err := c.Ask.RolesDictionary()
	if err != nil {
		return err
	}

	p.posts = posts.ParseMany(postponed, dictionary, c.Ask.OrganizationHashtags())
	p.schedule = p.posts.Schedule()

	return nil
}

func (p *Postponed) Posts() posts.Posts {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.posts
}

func (p *Postponed) PostsKind(mask posts.Kind) []posts.Post {
	p.mu.Lock()
	defer p.mu.Unlock()

	var result []posts.Post
	kinds := posts.ParseKinds(mask)

	for i := range kinds {
		result = append(result, p.posts[kinds[i]]...)
	}

	return result
}

func (p *Postponed) Schedule() schedule.Schedule {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.schedule
}

// add post to vk & cache
func (p *Postponed) AddPost(c *Controls, params vk.PostParams) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	id, err := c.Vk.CreatePostByParams(&params)
	if err != nil {
		return err
	}

	dictionary, err := c.Ask.RolesDictionary()
	if err != nil {
		return err
	}

	post := posts.ParseFromParams(id, params, dictionary, c.Ask.OrganizationHashtags())

	p.posts[post.Kind] = append(p.posts[post.Kind], *post)
	p.schedule = p.schedule.Add(params.PublishDate)

	return nil
}

func (p *Postponed) DeletePost(c *Controls, post posts.Post) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	err := c.Vk.DeletePost(post.ID)
	if err != nil {
		return err
	}

	for i := range p.posts[post.Kind] {
		if p.posts[post.Kind][i].ID == post.ID {
			p.posts[post.Kind] = append(p.posts[post.Kind][:i], p.posts[post.Kind][i+1:]...)
			break
		}
	}

	p.schedule = p.schedule.Delete(post.Date)

	return nil
}

func (p *Postponed) AddPosts(c *Controls, params []vk.PostParams) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	dictionary, err := c.Ask.RolesDictionary()
	if err != nil {
		return err
	}

	for i := range params {
		id, err := c.Vk.CreatePostByParams(&params[i])
		if err != nil {
			return err
		}

		post := posts.ParseFromParams(id, params[i], dictionary, c.Ask.OrganizationHashtags())

		p.posts[post.Kind] = append(p.posts[post.Kind], *post)
		p.schedule = p.schedule.Add(params[i].PublishDate)
	}

	return nil
}

func (p *Postponed) DeletePosts(c *Controls, posts []posts.Post) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i := range posts {
		err := c.Vk.DeletePost(posts[i].ID)
		if err != nil {
			return err
		}

		for i := range p.posts[posts[i].Kind] {
			if p.posts[posts[i].Kind][i].ID == posts[i].ID {
				p.posts[posts[i].Kind] = append(p.posts[posts[i].Kind][:i], p.posts[posts[i].Kind][i+1:]...)
				break
			}
		}

		p.schedule = p.schedule.Delete(posts[i].Date)
	}

	return nil
}
