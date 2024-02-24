package main

import "ask-bot/src/vk"

type ConfirmationNode struct {
	message string
	next    StateNode

	Answer bool
}

func NewConfirmationNode(message string, next StateNode) *ConfirmationNode {
	return &ConfirmationNode{
		message: message,
		next:    next,
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

func (node *ConfirmationNode) NewMessage(user *User, c *Controls, message *vk.Message) (StateNode, bool, error) {
	node.Answer = false

	return nil, true, nil
}

func (node *ConfirmationNode) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (StateNode, bool, error) {
	switch payload.Command {
	case "yes":
		node.Answer = true
		return node.next, false, nil
	case "no":
		node.Answer = false
	}

	return nil, true, nil
}

func (node *ConfirmationNode) Back(user *User, c *Controls, prev_state StateNode) (bool, error) {
	return true, nil
}

func (node *ConfirmationNode) Next() StateNode {
	return node.next
}
