package vk

import (
	"encoding/json"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

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

func UnmarshalPayload(message json.RawMessage) (*CallbackPayload, error) {
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

func CreateKeyboard(id string, buttons [][]Button) string {
	keyboard := object.NewMessagesKeyboard(false)

	for i := 0; i < len(buttons); i++ {
		keyboard.AddRow()

		for j := 0; j < len(buttons[i]); j++ {
			keyboard.AddCallbackButton(buttons[i][j].Label, CallbackPayload{
				Id:      id,
				Command: buttons[i][j].Command,
				Value:   buttons[i][j].Value,
			}, buttons[i][j].Color)
		}
	}

	return keyboard.ToJSON()
}

type ForwardMessage struct {
	PeerId      int   `json:"peer_id"`
	MessagesIds []int `json:"message_ids"`
}

func ForwardParam(vk_id int, messages []int) (string, error) {
	forward, err := json.Marshal(ForwardMessage{
		vk_id,
		messages,
	})
	if err != nil {
		return "", zaperr.Wrap(err, "failed to marshal forward message param",
			zap.Int("vk_id", vk_id),
			zap.Any("messages", messages))
	}

	return string(forward), nil
}
