package main

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
	buttons := [][]Button{
		{
			{
				Label:   "Да",
				Color:   PrimaryColor,
				Command: "yes",
			},
			{
				Label:   "Нет",
				Color:   NegativeColor,
				Command: "no",
			},
		},
	}

	_, err := c.Vk.SendMessage(user.id, node.message, CreateKeyboard(node, buttons), nil)
	return err
}

func (node *ConfirmationNode) NewMessage(user *User, c *Controls, message *Message) (StateNode, bool, error) {
	node.Answer = false

	return nil, true, nil
}

func (node *ConfirmationNode) KeyboardEvent(user *User, c *Controls, payload *CallbackPayload) (StateNode, bool, error) {
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
