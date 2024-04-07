package postponed

import (
	"ask-bot/src/ask"
	"ask-bot/src/posts"
	"slices"
	"time"
)

type CachePosts interface {
	AddNew(c *Controls, organization *ask.OrganizationHashtags, busy *([]time.Time)) error
	Remove(c *Controls) error
	Posts() []posts.Post
}

type Cache struct {
	organization *ask.OrganizationHashtags
	dictionary   []ask.Role

	polls  []posts.Post
	others []posts.Post

	busy []time.Time
}

func NewCache(c *Controls) *Cache {
	organization := c.Ask.OrganizationHashtags()

	return &Cache{
		organization: &organization,
	}
}

func (cache *Cache) internalUpdate(c *Controls) error {
	// TO-DO: maybe check if roles do not change somehow
	dictionary, err := c.Ask.RolesDictionary()
	if err != nil {
		return err
	}
	cache.dictionary = dictionary
	return nil
}

func (cache *Cache) update(c *Controls, db *DBInfo, vk *VKInfo) error {
	// init posts
	polls := NewCachePolls(db, vk)
	others := NewCacheOthers(vk)

	busy, err := cache.handle(c, polls, others)
	if err != nil {
		return err
	}

	// update info struct
	cache.polls = polls.Posts()
	cache.others = others.Posts()
	cache.busy = busy

	return nil
}

func calculateBusy(arrays ...[]posts.Post) []time.Time {
	length := 0

	for i := range arrays {
		length += len(arrays[i])
	}

	offset := 0
	busy := make([]time.Time, length)

	for i := range arrays {
		for j := range arrays[i] {
			busy[j+offset] = arrays[i][j].Date
		}
		offset += len(arrays[i])
	}

	slices.SortFunc(busy, func(a, b time.Time) int {
		return a.Compare(b)
	})

	return busy
}

func (cache *Cache) handle(c *Controls, kinds ...CachePosts) ([]time.Time, error) {
	// remove
	for i := range kinds {
		if err := kinds[i].Remove(c); err != nil {
			return nil, err
		}
	}

	current_posts := [][]posts.Post{}
	for i := range kinds {
		current_posts = append(current_posts, kinds[i].Posts())
	}
	busy := calculateBusy(current_posts...)

	// add new
	for i := range kinds {
		if err := kinds[i].AddNew(c, cache.organization, &busy); err != nil {
			return nil, err
		}
	}

	return busy, nil
}
