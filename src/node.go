package main

import (
	"encoding/json"
	"errors"

	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type Params map[string]interface{}

func (p Params) Bool(key string) (value bool, ok bool) {
	v, ok := p[key]
	if !ok {
		return false, false
	}

	value, ok = v.(bool)
	return
}

type StateNode interface {
	ID() string

	Entry(user_id int, ask *Ask, vk *VK, params Params) error
	Do(user_id int, ask *Ask, vk *VK, input interface{}) (StateNode, error)
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

	if payload.Id != node.ID() {
		err := errors.New("payload does not belong to node")
		return nil, zaperr.Wrap(err, "",
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
