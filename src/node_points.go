package main

import (
	"encoding/json"
	"fmt"

	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/object"
	"go.uber.org/zap"
)

type PointsNode struct{}

func (node *PointsNode) String() string {
	return "points"
}

func (node *PointsNode) Entry(ask *Ask, vk *VK, user_id int, silent bool) {
	points, err := ask.Points(user_id)
	if err != nil {
		zap.S().Warnw("failed to get points",
			"error", err,
			"user_id", user_id)
		return
	}

	keyboard := object.NewMessagesKeyboard(false)
	keyboard.AddRow()
	keyboard.AddCallbackButton("Назад", CallbackPayload{
		Id:    node.String(),
		Value: "back",
	}, "negative")

	vk.SendMessage(user_id, fmt.Sprintf("Ваше текущее количество баллов: %d", points), keyboard.ToJSON())
}

func (node *PointsNode) Do(ask *Ask, vk *VK, event EventType, i interface{}) StateNode {
	if event == ChangeKeyboardEvent {
		obj, ok := i.(events.MessageEventObject)
		if !ok {
			zap.S().Warnw("failed to parse vk response to new message object",
				"object", obj)
			return nil
		}

		var payload CallbackPayload

		err := json.Unmarshal([]byte(obj.Payload), &payload)
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

		if payload.Value == "back" {
			return &InitNode{}
		}
	}

	return nil
}
