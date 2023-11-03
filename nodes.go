package main

import (
	"encoding/json"

	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/object"
)

type EventType int

const (
	ChangeKeyboardEvent EventType = iota
	NewMessageEvent
)

type StateNode interface {
	Init(a *VkApi, db *Db, user_id int, silent bool)
	Do(a *VkApi, db *Db, event EventType, i interface{}) StateNode
	String() string
}

type InitNode struct{}

func (node *InitNode) Init(a *VkApi, d *Db, user_id int, silent bool) {
	keyboard := object.NewMessagesKeyboard(false)
	keyboard.AddRow()
	keyboard.AddCallbackButton("FAQ", (&FAQNode{}).String(), "secondary")

	if !silent {
		a.SendMessage(user_id, "Здравствуйте!", keyboard.ToJSON())
	} else {
		id := a.SendMessage(user_id, "Меняю клавиатуру", keyboard.ToJSON())
		a.DeleteMessage(user_id, id, 1)
	}
}

func (node *InitNode) Do(a *VkApi, d *Db, event EventType, i interface{}) StateNode {
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

		if payload == (&FAQNode{}).String() {
			return &FAQNode{}
		}
	}
	return nil
}

func (node *InitNode) String() string {
	return "init"
}

type FAQNode struct{}

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

func (node *FAQNode) String() string {
	return "FAQ"
}
