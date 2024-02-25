package main

import "ask-bot/src/vk"

type ConfirmationNode struct {
	message string

	payload string
}

func NewConfirmationNode(payload string, message string) *ConfirmationNode {
	return &ConfirmationNode{
		message: message,
		payload: payload,
	}
}

func (node *ConfirmationNode) ID() string {
	return "confirmation"
}

func (node *ConfirmationNode) Entry(user *User, c *Controls) error {
	buttons := [][]vk.Button{
		{
			{
				Label:   "Да",
				Color:   vk.PrimaryColor,
				Command: "yes",
			},
			{
				Label:   "Нет",
				Color:   vk.NegativeColor,
				Command: "no",
			},
		},
	}

	_, err := c.Vk.SendMessage(user.id, node.message, vk.CreateKeyboard(node.ID(), buttons), nil)
	return err
}

func (node *ConfirmationNode) NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error) {
	return nil, nil
}

func (node *ConfirmationNode) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error) {
	switch payload.Command {
	case "yes":
		return node.exit_action(true), nil

	case "no":
		return node.exit_action(false), nil

	}

	return nil, nil
}

func (node *ConfirmationNode) Back(user *User, c *Controls, info *ExitInfo) (*Action, error) {
	return nil, node.Entry(user, c)
}

func (node *ConfirmationNode) exit_action(answer bool) *Action {
	return NewActionExit(&ExitInfo{
		Values: map[string]interface{}{
			"confirmation": answer,
		},
		Payload: node.payload,
	})
}
