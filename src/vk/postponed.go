package vk

import (
	"regexp"
	"slices"
	"time"

	"github.com/SevereCloud/vksdk/v2/object"
)

type Post struct {
	ID      int
	Created time.Time

	Tags       []string
	Text       string
	Attachment []object.WallWallpostAttachment
}

func Parse(post *object.WallWallpost) Post {
	// find hashtags
	tags := regexp.MustCompile(`#([\w@]+)`).FindAllString(post.Text, -1)

	return Post{
		post.ID,
		time.Now(),

		tags,
		post.Text,
		post.Attachments,
	}
}

type Postponed struct {
	group int
	vk    *VK

	updated time.Time

	// sorted by id
	posts []Post
}

func NewPostponed(group int, vk *VK) *Postponed {
	return &Postponed{
		group: group,
		vk:    vk,
	}
}

func (p *Postponed) Update() ([]Post, error) {
	posts, err := p.vk.PostponedWallPosts(p.group)
	if err != nil {
		return nil, err
	}
	slices.SortFunc[[]object.WallWallpost, object.WallWallpost](
		posts,
		func(a, b object.WallWallpost) int {
			return b.ID - a.ID
		})

	i, j := 0, 0
	for i < len(p.posts) && j < len(posts) {
		if p.posts[i].ID == posts[j].ID {
			if time.Unix(int64(posts[j].Edited), 0).After(p.updated) {
				p.posts[i] = Parse(&posts[j])
			}
			i, j = i+1, j+1
			continue
		}

		if p.posts[i].ID < posts[j].ID {
			p.posts = append(p.posts[:i], p.posts[i+1:]...)
		} else {
			p.posts = append(p.posts[:i+1], p.posts[i:]...)
			p.posts[i] = Parse(&posts[j])

			i, j = i+1, j+1
		}
	}

	if i < len(p.posts) {
		p.posts = p.posts[:i+1]
	}
	for ; j < len(posts); j++ {
		p.posts = append(p.posts, Parse(&posts[j]))
	}

	p.updated = time.Now()

	return p.posts, nil
}
