package main

import (
	"encoding/json"
	"errors"

	"github.com/SevereCloud/vksdk/v2/object"
)

type StateNode interface {
	ID() string

	Entry(user_id int, ask *Ask, vk *VK, silent bool)
	Do(user_id int, ask *Ask, vk *VK, input interface{}) StateNode
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
		return nil, err
	}

	if payload.Id != node.ID() {
		return nil, errors.New("payload does not belong to node")
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
