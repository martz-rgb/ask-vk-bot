package main

import (
	"github.com/SevereCloud/vksdk/v2/events"
	"go.uber.org/zap"
)

type InitNode struct{}

func (node *InitNode) ID() string {
	return "init"
}

func (node *InitNode) Entry(user_id int, ask *Ask, vk *VK, silent bool) {
	buttons := [][]Button{{
		{
			Label:   "Список ролей",
			Color:   "secondary",
			Command: (&RolesNode{}).ID(),
		},
		{
			Label:   "Баллы",
			Color:   "secondary",
			Command: (&PointsNode{}).ID(),
		},
		{
			Label:   "FAQ",
			Color:   "secondary",
			Command: (&FAQNode{}).ID(),
		},
	}}

	if !silent {
		vk.SendMessage(user_id, "Здравствуйте!", CreateKeyboard(node, buttons))
	} else {
		vk.ChangeKeyboard(user_id, CreateKeyboard(node, buttons))
	}
}

func (node *InitNode) Do(user_id int, ask *Ask, vk *VK, input interface{}) StateNode {
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

func (node *InitNode) KeyboardEvent(ask *Ask, vk *VK, payload *CallbackPayload) StateNode {
	switch payload.Command {
	case (&FAQNode{}).ID():
		return &FAQNode{}
	case (&RolesNode{}).ID():
		return &RolesNode{}
	case (&PointsNode{}).ID():
		return &PointsNode{}
	}

	return nil
}
