package main

import (
	"encoding/json"

	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/object"
)

var dict StateMachine = StateMachine{
	InitState: &State{
		Entry: func(a *VkApi, user_id int, silent bool) {
			keyboard := object.NewMessagesKeyboard(false)
			keyboard.AddRow()
			keyboard.AddCallbackButton("FAQ", FaqState, "secondary")

			if !silent {
				a.SendMessage(user_id, "Здравствуйте!", keyboard.ToJSON())
			} else {
				id := a.SendMessage(user_id, "Меняю клавиатуру", keyboard.ToJSON())
				a.DeleteMessage(user_id, id, 1)
			}
		},
		Do: func(a *VkApi, event EventType, i interface{}) (next StateNode, change bool) {
			if event == ChangeKeyboardEvent {
				obj, ok := i.(events.MessageEventObject)
				if !ok {
					return
				}

				var payload StateNode

				err := json.Unmarshal(obj.Payload, &payload)
				if err != nil {
					return UndefindedState, false
				}

				if payload == FaqState {
					return FaqState, true
				}
			}

			return UndefindedState, false
		},
	},
	FaqState: &State{
		Entry: func(a *VkApi, user_id int, silent bool) {
			keyboard := object.NewMessagesKeyboard(false)

			keyboard.AddRow()

			keyboard.AddCallbackButton("Кто я?", "who", "secondary")
			keyboard.AddCallbackButton("Что я могу делать?", "what", "secondary")
			keyboard.AddCallbackButton("Назад", "back", "primary")

			a.SendMessage(user_id, "Выберите вопрос, который вас интересует на клавиатуре ниже.", keyboard.ToJSON())
		},
		Do: func(a *VkApi, event EventType, i interface{}) (next StateNode, change bool) {
			if event == ChangeKeyboardEvent {
				obj, ok := i.(events.MessageEventObject)
				if !ok {
					return
				}

				var payload string

				err := json.Unmarshal(obj.Payload, &payload)
				if err != nil {
					return UndefindedState, false
				}

				if payload == "who" {
					a.SendMessage(obj.UserID, "Я подрядчик этого дома.", "")
					return UndefindedState, false
				}

				if payload == "what" {
					a.SendMessage(obj.UserID, "Я умею отвечать на ваши сообщения и управлять этим домом.", "")
					return UndefindedState, false
				}

				if payload == "back" {
					return InitState, true
				}
			}

			return UndefindedState, false
		},
	},
}

//m.SendMessage(obj.Message.FromID, "Я получил ваше сообщение.", Keyboard)

// fmt.Printf("message event %+v\n", obj)

// // answer on callback to clear loading
// m.SendEventAnswer(obj.EventID, obj.UserID, obj.PeerID)

// message_id := m.SendMessage(obj.UserID, "Меняю клавиатуру...", EmptyKeyboard)

// // delete
// m.DeleteMessage(obj.PeerID, message_id, 1)
