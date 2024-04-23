package postponed

import (
	"ask-bot/src/posts"
	"ask-bot/src/schedule"
	"slices"
	"time"
)

func (p *Postponed) update(c *Controls, db *DBInfo, vk *VKInfo) error {
	new, invalid := exclude(vk, db)
	invalid = append(invalid, vk.posts[posts.Kinds.Invalid]...)

	for _, post := range invalid {
		err := c.Vk.DeletePost(post.ID)
		if err != nil {
			// TO-DO maybe do not return
			return err
		}
	}
	delete(vk.posts, posts.Kinds.Invalid)

	busy := calculateBusy(vk.posts)

	posts, err := c.createNew(new, &busy)
	if err != nil {
		return err
	}

	for key := range vk.posts {
		posts[key] = append(posts[key], vk.posts[key]...)
	}

	p.posts = posts
	p.busy = busy

	return nil
}

func calculateBusy(arrays Dictionary) schedule.Schedule {
	length := 0

	for key := range arrays {
		length += len(arrays[key])
	}

	offset := 0
	busy := make([]time.Time, length)

	for key := range arrays {
		for j := range arrays[key] {
			busy[j+offset] = arrays[key][j].Date
		}
		offset += len(arrays[key])
	}

	slices.SortFunc(busy, func(a, b time.Time) int {
		return a.Compare(b)
	})

	return busy
}
