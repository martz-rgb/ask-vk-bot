package vk

import (
	"regexp"
	"slices"
	"time"

	"github.com/SevereCloud/vksdk/v2/object"
)

type Post struct {
	ID int

	Tags []string
}

func Parse(post *object.WallWallpost) Post {
	// find hashtags
	tags := regexp.MustCompile(`#([\w@]+)`).FindAllString(post.Text, -1)

	return Post{
		ID:   post.ID,
		Tags: tags,
	}
}

type Wall struct {
	group int
	vk    *VK

	updated int64

	// sorted by id
	posts []Post
}

func NewWall(group int, vk *VK) *Wall {
	return &Wall{
		group: group,
		vk:    vk,
	}
}

func (w *Wall) Update() error {
	posts, err := w.vk.PostponedWallPosts(w.group)
	if err != nil {
		return err
	}
	slices.SortFunc[[]object.WallWallpost, object.WallWallpost](
		posts,
		func(a, b object.WallWallpost) int {
			return b.ID - a.ID
		})

	i, j := 0, 0
	for i < len(w.posts) && j < len(posts) {
		if w.posts[i].ID == posts[j].ID {
			if int64(posts[j].Edited) > w.updated {
				w.posts[i] = Parse(&posts[j])
			}
			i, j = i+1, j+1
			continue
		}

		if w.posts[i].ID < posts[j].ID {
			w.posts = append(w.posts[:i], w.posts[i+1:]...)
		} else {
			w.posts = append(w.posts[:i+1], w.posts[i:]...)
			w.posts[i] = Parse(&posts[j])

			i, j = i+1, j+1
		}
	}

	if i < len(w.posts) {
		w.posts = w.posts[:i+1]
	}
	for ; j < len(posts); j++ {
		w.posts = append(w.posts, Parse(&posts[j]))
	}

	w.updated = time.Now().Unix()

	return nil
}
