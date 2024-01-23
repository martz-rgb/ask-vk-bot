package main

import (
	"github.com/SevereCloud/vksdk/v2/events"
	"go.uber.org/zap"
)

type FAQNode struct{}

func (node *FAQNode) ID() string {
	return "faq"
}

func (node *FAQNode) Entry(user_id int, ask *Ask, vk *VK, params Params) error {
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

	_, err := vk.SendMessage(user_id, message, CreateKeyboard(node, buttons), nil)
	return err
}

func (node *FAQNode) Do(user_id int, ask *Ask, vk *VK, input interface{}) (StateNode, error) {
	switch obj := input.(type) {

	case events.MessageEventObject:
		payload, err := UnmarshalPayload(node, obj.Payload)
		if err != nil {
			return nil, err
		}

		return node.KeyboardEvent(ask, vk, obj.UserID, payload), nil

	default:
		zap.S().Infow("failed to parse vk response to message event object",
			"object", obj)
	}

	return nil, nil
}

func (node *FAQNode) KeyboardEvent(ask *Ask, vk *VK, user_id int, payload *CallbackPayload) StateNode {
	switch payload.Command {
	case "who":
		vk.SendMessage(user_id, "Я подрядчик этого дома.", "", nil)
		return nil
	case "what":
		vk.SendMessage(user_id, "Я умею отвечать на ваши сообщения и управлять этим домом.", "", nil)
		return nil
	case "back":
		return &InitNode{}
	}

	return nil
}
