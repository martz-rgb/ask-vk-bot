package postponed

import (
	"ask-bot/src/datatypes/posts"
)

type VKInfo struct {
	posts Dictionary
}

func NewVKInfo(c *Controls) (*VKInfo, error) {
	dictionary, err := c.Ask.RolesDictionary()
	if err != nil {
		return nil, err
	}

	postponed, err := c.Vk.PostponedPosts()
	if err != nil {
		return nil, err
	}

	postponed_posts := posts.ParseMany(postponed, dictionary, c.Ask.OrganizationHashtags())

	info := &VKInfo{
		posts: make(Dictionary),
	}

	for _, post := range postponed_posts {
		info.posts[post.Kind] = append(info.posts[post.Kind], post)
	}

	return info, nil
}
