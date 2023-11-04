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

	keyboard.AddCallbackButton("Кто я?", "who", "secondary")
	keyboard.AddCallbackButton("Что я могу делать?", "what", "secondary")
	keyboard.AddCallbackButton("Назад", "back", "primary")

	a.SendMessage(user_id, "Выберите вопрос, который вас интересует на клавиатуре ниже.", keyboard.ToJSON())
}

func (node *FAQNode) Do(a *VkApi, d *Db, event EventType, i interface{}) StateNode {
	if event == ChangeKeyboardEvent {
		obj, ok := i.(events.MessageEventObject)
		if !ok {
			return nil
		}

		var payload string

		err := json.Unmarshal(obj.Payload, &payload)
		if err != nil {
			return nil
		}

		if payload == "who" {
			a.SendMessage(obj.UserID, "Я подрядчик этого дома.", "")
			return nil
		}

		if payload == "what" {
			a.SendMessage(obj.UserID, "Я умею отвечать на ваши сообщения и управлять этим домом.", "")
			return nil
		}

		if payload == "back" {
			return &InitNode{}
		}
	}

	return nil
}
