package vk

import (
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

func (v *VK) CreateBoard(title string, text string, attachments string) (int, error) {
	id := v.id
	if id < 0 {
		id = -id
	}

	params := api.Params{
		"group_id":    id,
		"title":       title,
		"text":        text,
		"attachments": attachments,
		"from_group":  1,
	}

	response, err := v.api.BoardAddTopic(params)
	if err != nil {
		return -1, zaperr.Wrap(err, "failed to create board",
			zap.Any("params", params),
			zap.Int("response", response))
	}

	zap.S().Debugw("successfully created board",
		"params", params,
		"response", response)

	return response, nil
}
