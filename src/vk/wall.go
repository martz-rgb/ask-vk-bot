package vk

import (
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

func (v *VK) WallPostNew(group_id int, message string, attachments string, signed bool, publish_date int64) error {
	params := api.Params{
		"owner_id":     -group_id,
		"from_group":   1,
		"message":      message,
		"attachments":  attachments,
		"signed":       signed,
		"publish_date": publish_date,
	}

	response, err := v.api.WallPost(params)
	if err != nil {
		return zaperr.Wrap(err, "failed to post on wall",
			zap.Any("params", params),
			zap.Any("response", response))
	}

	zap.S().Debugw("successfully posted on wall",
		"params", params,
		"response", response)

	return nil
}

func (v *VK) PostponedWallPosts(group_id int) ([]object.WallWallpost, error) {
	max_count := 100

	var posts []object.WallWallpost
	offset := 0

	for {
		params := api.Params{
			"owner_id": -group_id,
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
