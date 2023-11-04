package main

import (
	"encoding/json"

	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/object"
)

type FAQNode struct{}

func (node *FAQNode) String() string {
	return "faq"
}

func (node *FAQNode) Init(a *VkApi, d *Db, user_id int, silent bool) {
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

	a.SendMessage(user_id, "Выберите вопрос, который вас интересует на клавиатуре ниже.", keyboard.ToJSON())
}

func (node *FAQNode) Do(a *VkApi, d *Db, event EventType, i interface{}) StateNode {
	if event == ChangeKeyboardEvent {
		obj, ok := i.(events.MessageEventObject)
		if !ok {
			return nil
		}

		var payload CallbackPayload

		err := json.Unmarshal(obj.Payload, &payload)
		if err != nil {
			return nil
		}

		if payload.Id != node.String() {
			return nil
		}

		if payload.Value == "who" {
			a.SendMessage(obj.UserID, "Я подрядчик этого дома.", "")
			return nil
		}

		if payload.Value == "what" {
			a.SendMessage(obj.UserID, "Я умею отвечать на ваши сообщения и управлять этим домом.", "")
			return nil
		}

		if payload.Value == "back" {
			return &InitNode{}
		}
	}

	return nil
}
