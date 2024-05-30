package form

import (
	"ask-bot/src/dict"
	"ask-bot/src/vk"
)

type Request struct {
	Message *vk.MessageParams
	Options []Option
}

func AlwaysRequest(message *vk.MessageParams, options []Option) func(dict.Dictionary) (*Request, bool, error) {
	return func(d dict.Dictionary) (*Request, bool, error) {
		return &Request{
			Message: message,
			Options: options,
		}, false, nil
	}
}

// TO-DO: yes and no words from templates
func AlwaysConfirm(message_builder func(dict.Dictionary) *vk.MessageParams) func(dict.Dictionary) (*Request, bool, error) {
	options := []Option{
		{
			ID:    "true",
			Label: "Да",
			Color: vk.PositiveColor,
			Value: true,
		},
		{
			ID:    "false",
			Label: "Нет",
			Color: vk.NegativeColor,
			Value: false,
		},
	}

	return func(d dict.Dictionary) (*Request, bool, error) {
		return &Request{
			Message: message_builder(d),
			Options: options,
		}, false, nil
	}
}
