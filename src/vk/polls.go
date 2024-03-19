package vk

import (
	"encoding/json"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

func (v *VK) CreatePoll(question string, answers []string, is_anonymous bool, end_date int64) (*object.PollsPoll, error) {
	json_answers, err := json.Marshal(answers)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to marshal answers",
			zap.Any("answers", answers))
	}

	params := api.Params{
		"owner_id":     v.id,
		"question":     question,
		"add_answers":  string(json_answers),
		"is_anonymous": is_anonymous,
		"end_date":     end_date,
	}

	response, err := v.api.PollsCreate(params)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to create a poll",
			zap.Any("params", params),
			zap.Any("reponse", response))
	}

	zap.S().Debugw("successfully created a poll",
		"params", params,
		"response", response)

	poll := object.PollsPoll(response)
	return &poll, nil
}
