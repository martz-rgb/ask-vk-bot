package main

import (
	"github.com/SevereCloud/vksdk/v2/events"
	"go.uber.org/zap"
)

type InitNode struct{}

func (node *InitNode) ID() string {
	return "init"
}

func (node *InitNode) Entry(user_id int, ask *Ask, vk *VK, params Params) error {
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

	silent, ok := params.Bool("silent")

	if ok && silent {
		return vk.ChangeKeyboard(user_id, CreateKeyboard(node, buttons))
	}

	_, err := vk.SendMessage(user_id, "Здравствуйте!", CreateKeyboard(node, buttons), nil)
	return err
}

func (node *InitNode) Do(user_id int, ask *Ask, vk *VK, input interface{}) (StateNode, error) {
	switch obj := input.(type) {

	case events.MessageEventObject:
		payload, err := UnmarshalPayload(node, obj.Payload)
		if err != nil {
			return nil, err
		}

		return node.KeyboardEvent(ask, vk, payload), nil

	default:
		zap.S().Infow("failed to parse vk response to message event object",
			"object", obj)
	}

	return nil, nil
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
