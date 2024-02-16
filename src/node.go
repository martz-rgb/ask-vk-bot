package main

import (
	"encoding/json"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type StateNode interface {
	ID() string

	Entry(user *User, c *Controls) error
	NewMessage(user *User, c *Controls, message *Message) (StateNode, bool, error)
	KeyboardEvent(user *User, c *Controls, payload *CallbackPayload) (StateNode, bool, error)
	Back(user *User, c *Controls, prev_state StateNode) (bool, error)
}

type MessageParams struct {
	Id     int
	Text   string
	Params api.Params
}

type Message struct {
	ID          int
	Text        string
	Attachments []object.MessagesMessageAttachment
}

type CallbackPayload struct {
	Command string `json:"command"`
	Value   string `json:"value"`
	Id      string `json:"id"`
}

func UnmarshalPayload(node StateNode, message json.RawMessage) (*CallbackPayload, error) {
	payload := &CallbackPayload{}

	err := json.Unmarshal(message, &payload)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to unmarshal payload",
			zap.Any("payload", payload))
	}

	return payload, nil
}

type Button struct {
	Label string
	Color string

	Command string
	Value   string
}

func CreateKeyboard(node StateNode, buttons [][]Button) string {
	keyboard := object.NewMessagesKeyboard(false)

	for i := 0; i < len(buttons); i++ {
		keyboard.AddRow()

		for j := 0; j < len(buttons[i]); j++ {
			keyboard.AddCallbackButton(buttons[i][j].Label, CallbackPayload{
				Id:      node.ID(),
				Command: buttons[i][j].Command,
				Value:   buttons[i][j].Value,
			}, buttons[i][j].Color)
		}
	}

	return keyboard.ToJSON()
}
