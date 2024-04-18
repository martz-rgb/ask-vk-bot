package vk

import (
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

func (v *VK) CreateAlbum(title string) (int, error) {
	id := v.id
	if id < 0 {
		id = -id
	}
	params := api.Params{
		"title":                 title,
		"group_id":              id,
		"upload_by_admins_only": 1,
	}

	response, err := v.api.PhotosCreateAlbum(params)
	if err != nil {
		return -1, zaperr.Wrap(err, "failed to create album",
			zap.Any("params", params),
			zap.Any("response", response))
	}

	zap.S().Debugw("successfully created album",
		"params", params,
		"response", response)

	return response.ID, nil
}
