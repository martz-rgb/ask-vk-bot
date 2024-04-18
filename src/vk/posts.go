package vk

import (
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
