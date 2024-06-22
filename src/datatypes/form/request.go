package form

import (
	"ask-bot/src/datatypes/dict"
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
func AlwaysConfirm(message *vk.MessageParams) func(dict.Dictionary) (*Request, bool, error) {
	options := []Option{
		{
			ID:    "true",
			Label: "Да",
			Color: vk.PrimaryColor,
			Value: true,
		},
		{
			ID:    "false",
			Label: "Нет",
			Color: vk.SecondaryColor,
			Value: false,
		},
	}

	return func(d dict.Dictionary) (*Request, bool, error) {
		return &Request{
			Message: message,
			Options: options,
		}, false, nil
	}
}
