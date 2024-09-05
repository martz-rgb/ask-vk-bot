package watcher

import "ask-bot/src/datatypes/posts"

// make diff tasks like check polls, check
// func (w *Watcher) updatePostponed() error {
// 	return w.p.Update(&postponed.Controls{
// 		Ask: w.c.Ask,
// 		Vk:  w.c.Admin,
// 	})
// }

func (c *Controls) UpdatePostponed() error {
	return c.Postponed.Update(c.PostponedControls())
}

func (c *Controls) DeleteInvalidPostponed() error {
	invalid := c.Postponed.PostsKind(posts.Kinds.Invalid)

	ids := make([]posts.Post, len(invalid))

	for _, post := range invalid {
		ids = append(ids, post)
	}

	return c.Postponed.DeletePosts(c.PostponedControls(), ids)
}
