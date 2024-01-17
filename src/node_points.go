package main

import (
	"fmt"

	"github.com/SevereCloud/vksdk/v2/events"
	"go.uber.org/zap"
)

type PointsNode struct{}

func (node *PointsNode) ID() string {
	return "points"
}

func (node *PointsNode) Entry(user_id int, ask *Ask, vk *VK, silent bool) {
	points, err := ask.Points(user_id)
	if err != nil {
		zap.S().Warnw("failed to get points",
			"error", err,
			"user_id", user_id)
		return
	}

	buttons := [][]Button{{{
		Label: "Назад",
		Color: "negative",

		Command: "back",
	}}}

	message := fmt.Sprintf("Ваше текущее количество баллов: %d", points)

	vk.SendMessage(user_id, message, CreateKeyboard(node, buttons))
}

func (node *PointsNode) Do(user_id int, ask *Ask, vk *VK, input interface{}) StateNode {
	switch obj := input.(type) {

	case events.MessageEventObject:
		payload, err := UnmarshalPayload(node, obj.Payload)
		if err != nil {
			zap.S().Errorw("failed to unmarshal payload",
				"payload", payload)
			return nil
		}

		return node.KeyboardEvent(ask, vk, payload)

	default:
		zap.S().Warnw("failed to parse vk response to message event object",
			"object", obj)
	}

	return nil
}

func (node *PointsNode) KeyboardEvent(ask *Ask, vk *VK, payload *CallbackPayload) StateNode {
	switch payload.Command {
	case "back":
		return &InitNode{}
	}

	return nil
}
