package postponed

import (
	"ask-bot/src/ask"
	"ask-bot/src/posts"
)

type DBInfo struct {
	polls []ask.Poll
}

func NewDBInfo(c *Controls) (*DBInfo, error) {
	polls, err := c.Ask.PendingPolls()
	if err != nil {
		return nil, err
	}

	return &DBInfo{
		polls: polls,
	}, nil
}

type VKInfo struct {
	polls  []posts.Post
	others []posts.Post

	invalid []posts.Post
}

func NewVKInfo(c *Controls, cache *Cache) (*VKInfo, error) {
	postponed, err := c.Vk.PostponedPosts()
	if err != nil {
		return nil, err
	}

	postponed_posts := posts.ParseMany(postponed, cache.dictionary, cache.organization)

	var polls []posts.Post
	var others []posts.Post
	var invalid []posts.Post

	for _, post := range postponed_posts {
		switch post.Kind {
		case posts.Poll:
			if len(post.Roles) != 1 {
				invalid = append(invalid, post)
			} else {
				polls = append(polls, post)
			}
		default:
			others = append(others, post)
		}
	}

	return &VKInfo{
		polls:   polls,
		others:  others,
		invalid: invalid,
	}, nil
}
