package vk

import (
	"fmt"
	"strings"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

// i honestly don't know when reply and copy types are used
const (
	PostponedPost = "postpone"
	SuggestedPost = "suggest"
	ReplyPost     = "reply"
	CopyPost      = "copy"
	NewPost       = "post"
)

type PostParams struct {
	Text        string
	Attachments []string
	Signed      bool
	PublishDate time.Time
}

func (v *VK) PostLink(post int) string {
	return fmt.Sprintf("https://vk.com/wall%d_%d", v.id, post)
}

func (v *VK) CreatePost(text string, attachments string, signed bool, publish_date int64) (int, error) {
	params := api.Params{
		"owner_id":     v.id,
		"from_group":   1,
		"message":      text,
		"attachments":  attachments,
		"signed":       signed,
		"publish_date": publish_date,
	}

	response, err := v.api.WallPost(params)
	if err != nil {
		return 0, zaperr.Wrap(err, "failed to create post",
			zap.Any("params", params),
			zap.Any("response", response))
	}

	zap.S().Debugw("successfully created post",
		"params", params,
		"response", response)

	return response.PostID, nil
}

func (v *VK) CreatePostByParams(post *PostParams) (int, error) {
	params := api.Params{
		"owner_id":     v.id,
		"from_group":   1,
		"message":      post.Text,
		"attachments":  strings.Join(post.Attachments, ","),
		"signed":       post.Signed,
		"publish_date": post.PublishDate.Unix(),
	}

	response, err := v.api.WallPost(params)
	if err != nil {
		return 0, zaperr.Wrap(err, "failed to create post by params",
			zap.Any("params", params),
			zap.Any("response", response))
	}

	zap.S().Debugw("successfully created post by params",
		"params", params,
		"response", response)

	return response.PostID, nil
}

func (v *VK) PostsByIds(ids []int) ([]object.WallWallpost, error) {
	formatted_ids := make([]string, len(ids))

	for i := range ids {
		formatted_ids[i] = fmt.Sprintf("%d_%d", v.id, ids[i])
	}

	params := api.Params{
		"posts": formatted_ids,
	}

	response, err := v.api.WallGetByID(params)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get posts by ids",
			zap.Any("params", params),
			zap.Any("response", response))
	}

	zap.S().Debugw("successfully get posts by ids",
		"params", params,
		"response", response)

	return response, nil
}

func (v *VK) PostponedPosts() ([]object.WallWallpost, error) {
	max_count := 100

	var posts []object.WallWallpost
	offset := 0

	for {
		params := api.Params{
			"owner_id": v.id,
			"count":    max_count,
			"filter":   "postponed",
			"offset":   offset,
		}

		response, err := v.api.WallGet(params)
		if err != nil {
			return nil, zaperr.Wrap(err, "failed to get postponed wall posts",
				zap.Any("params", params),
				zap.Any("response", response))
		}

		zap.S().Debugw("successfully get postponed wall posts",
			"params", params,
			"response", response)

		posts = append(posts, response.Items...)

		offset += max_count

		if response.Count-offset <= 0 {
			return posts, nil
		}
	}
}

func (v *VK) DeletePost(post_id int) error {
	params := api.Params{
		"owner_id": v.id,
		"post_id":  post_id,
	}

	response, err := v.api.WallDelete(params)
	if err != nil {
		return zaperr.Wrap(err, "failed to delete post",
			zap.Any("params", params),
			zap.Any("response", response))
	}

	zap.S().Debugw("successfully deleted post",
		"params", params,
		"response", response)

	return nil
}
