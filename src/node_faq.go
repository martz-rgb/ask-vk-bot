package main

import (
	"encoding/json"

	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/object"
	"go.uber.org/zap"
)

type FAQNode struct{}

func (node *FAQNode) String() string {
	return "faq"
}

func (node *FAQNode) Init(v *VK, d *DB, user_id int, silent bool) {
	keyboard := object.NewMessagesKeyboard(false)

	keyboard.AddRow()

	keyboard.AddCallbackButton("Кто я?", CallbackPayload{
		Id:    node.String(),
		Value: "who",
	}, "secondary")
	keyboard.AddCallbackButton("Что я могу делать?", CallbackPayload{
		Id:    node.String(),
		Value: "what",
	}, "secondary")
	keyboard.AddCallbackButton("Назад", CallbackPayload{
		Id:    node.String(),
		Value: "back",
	}, "primary")

	v.SendMessage(user_id, "Выберите вопрос, который вас интересует на клавиатуре ниже.", keyboard.ToJSON())
}

func (node *FAQNode) Do(v *VK, d *DB, event EventType, i interface{}) StateNode {
	if event == ChangeKeyboardEvent {
		obj, ok := i.(events.MessageEventObject)
		if !ok {
			zap.S().Warnw("failed to parse vk response to message event object",
				"object", obj)
			return nil
		}

		var payload CallbackPayload

		err := json.Unmarshal(obj.Payload, &payload)
		if err != nil {
			zap.S().Errorw("failed to unmarshal payload",
				"payload", payload)
			return nil
		}

		if payload.Id != node.String() {
			zap.S().Infow("payload does not belong to node",
				"node", node.String(),
				"payload", payload)
			return nil
		}

		if payload.Value == "who" {
			v.SendMessage(obj.UserID, "Я подрядчик этого дома.", "")
			return nil
		}

		if payload.Value == "what" {
			v.SendMessage(obj.UserID, "Я умею отвечать на ваши сообщения и управлять этим домом.", "")
			return nil
		}

		if payload.Value == "back" {
			return &InitNode{}
		}
	}

	return nil
}
