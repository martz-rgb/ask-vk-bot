package main

import (
	"encoding/json"

	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/object"
)

type InitNode struct{}

func (node *InitNode) String() string {
	return "init"
}

func (node *InitNode) Init(a *VkApi, d *Db, user_id int, silent bool) {
	keyboard := object.NewMessagesKeyboard(false)
	keyboard.AddRow()
	keyboard.AddCallbackButton("Список ролей", (&RolesNode{}).String(), "secondary")
	keyboard.AddCallbackButton("FAQ", (&FAQNode{}).String(), "secondary")

	if !silent {
		a.SendMessage(user_id, "Здравствуйте!", keyboard.ToJSON())
	} else {
		a.ChangeKeyboard(user_id, keyboard.ToJSON())
	}
}

func (node *InitNode) Do(a *VkApi, d *Db, event EventType, i interface{}) StateNode {
	if event == ChangeKeyboardEvent {
		obj, ok := i.(events.MessageEventObject)
		if !ok {
			return nil
		}

		// to-do change payload to CallbackPayload
		var payload string

		err := json.Unmarshal(obj.Payload, &payload)
		if err != nil {
			return nil
		}

		if payload == (&FAQNode{}).String() {
			return &FAQNode{}
		}

		if payload == (&RolesNode{}).String() {
			return &RolesNode{}
		}
	}
	return nil
}
