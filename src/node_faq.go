package main

import (
	"github.com/SevereCloud/vksdk/v2/events"
	"go.uber.org/zap"
)

type FAQNode struct{}

func (node *FAQNode) ID() string {
	return "faq"
}

func (node *FAQNode) Entry(user_id int, ask *Ask, vk *VK, silent bool) {
	buttons := [][]Button{{
		{
			Label: "Кто ты?",
			Color: "secondary",

			Command: "who",
		},
		{
			Label: "Что ты можешь делать?",
			Color: "secondary",

			Command: "what",
		},
		{
			Label: "Назад",
			Color: "secondary",

			Command: "back",
		},
	}}

	message := "Выберите вопрос, который вас интересует на клавиатуре ниже."

	vk.SendMessage(user_id, message, CreateKeyboard(node, buttons))
}

func (node *FAQNode) Do(user_id int, ask *Ask, vk *VK, input interface{}) StateNode {
	switch obj := input.(type) {

	case events.MessageEventObject:
		payload, err := UnmarshalPayload(node, obj.Payload)
		if err != nil {
			zap.S().Errorw("failed to unmarshal payload",
				"payload", payload)
			return nil
		}

		return node.KeyboardEvent(ask, vk, obj.UserID, payload)

	default:
		zap.S().Warnw("failed to parse vk response to message event object",
			"object", obj)
	}

	return nil
}

func (node *FAQNode) KeyboardEvent(ask *Ask, vk *VK, user_id int, payload *CallbackPayload) StateNode {
	switch payload.Command {
	case "who":
		vk.SendMessage(user_id, "Я подрядчик этого дома.", "")
		return nil
	case "what":
		vk.SendMessage(user_id, "Я умею отвечать на ваши сообщения и управлять этим домом.", "")
		return nil
	case "back":
		return &InitNode{}
	}

	return nil
}
