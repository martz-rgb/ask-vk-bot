package main

import (
	"ask-bot/src/form"
	"ask-bot/src/vk"
	"errors"
	"strings"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

func ExtractID(message *vk.Message) interface{} {
	if message == nil {
		return nil
	}

	return message.ID
}

func ExtractAttachments(message *vk.Message) interface{} {
	if message == nil {
		return nil
	}

	return vk.ToAttachments(message.Attachments)
}

// get info about user
func InfoAboutValidate(value interface{}) (*vk.MessageParams, error) {
	if value == nil {
		return &vk.MessageParams{
			Text: "Поле обязательно для заполнения.",
		}, nil
	}

	message, ok := value.(int)
	if !ok {
		err := errors.New("failed to convert about value to int")
		return nil, zaperr.Wrap(err, "",
			zap.Any("value", value),
			zap.String("field", "AboutField"))
	}

	if message == 0 {
		return &vk.MessageParams{
			Text: "Поле обязательно для заполнения.",
		}, nil
	}

	return nil, nil
}

// (admin) confirm reservation
var ConfirmReservationOptions = []form.Option{
	{
		ID:    "confirm",
		Label: "Потвердить",
		Value: true,
	},
	{
		ID:    "delete",
		Label: "Удалить",
		Value: false,
	},
}

func ConfirmReservationValidate(value interface{}) (*vk.MessageParams, error) {
	if value == nil {
		return &vk.MessageParams{
			Text: "Поле обязательно для заполнения.",
		}, nil
	}

	if _, ok := value.(bool); !ok {
		err := errors.New("failed to convert about value to bool")
		return nil, zaperr.Wrap(err, "",
			zap.Any("value", value),
			zap.String("field", "ConfirmReservationField"))
	}

	return nil, nil
}

// check for photo attachments
func GreetingValidate(value interface{}) (*vk.MessageParams, error) {
	if value == nil {
		return &vk.MessageParams{
			Text: "Поле обязательно для заполнения.",
		}, nil
	}

	attachments, ok := value.(string)
	if !ok {
		err := errors.New("failed to convert about value to bool")
		return nil, zaperr.Wrap(err, "",
			zap.Any("value", value),
			zap.String("field", "ConfirmReservationField"))
	}

	items := strings.Split(attachments, ",")

	is_photo := false
	for _, item := range items {
		if strings.HasPrefix(item, "photo") {
			is_photo = true
			break
		}
	}

	if !is_photo {
		return &vk.MessageParams{
			Text: "Сообщение должно содержать изображения.",
		}, nil
	}

	return nil, nil
}
