package postponed

import (
	"ask-bot/src/ask"
	"ask-bot/src/posts"
	"time"
)

type CacheOthers struct {
	posts []posts.Post
}

func NewCacheOthers(vk *VKInfo) *CacheOthers {
	return &CacheOthers{
		posts: vk.others,
	}
}

func (o *CacheOthers) AddNew(c *Controls, organization *ask.OrganizationHashtags, busy *([]time.Time)) error {
	return nil
}
func (o *CacheOthers) Remove(c *Controls) error {
	return nil
}
func (o *CacheOthers) Posts() []posts.Post {
	return o.posts
}
