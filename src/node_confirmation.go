package main

type ConfirmationNode struct {
	Message string
	Next    StateNode
	Answer  bool
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

	_, err := c.Vk.SendMessage(user.id, node.Message, CreateKeyboard(node, buttons), nil)
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

		if node.Next != nil {
			return node.Next, false, nil
		}
	case "no":
		node.Answer = false
	}

	return nil, true, nil
}

func (node *ConfirmationNode) Back(user *User, c *Controls, prev_state StateNode) (bool, error) {
	return true, nil
}
