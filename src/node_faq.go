package main

import "ask-bot/src/vk"

type FAQNode struct{}

func (node *FAQNode) ID() string {
	return "faq"
}

func (node *FAQNode) Entry(user *User, c *Controls) error {
	buttons := [][]vk.Button{{
		{
			Label: "Кто ты?",
			Color: "secondary",

			Command: "who",
		},
		{
			Label: "Что ты можешь делать?",
			Color: "secondary",

			Command: "what",
		},
		{
			Label: "Назад",
			Color: "secondary",

			Command: "back",
		},
	}}

	message := "Выберите вопрос, который вас интересует на клавиатуре ниже."

	_, err := c.Vk.SendMessage(user.id, message, vk.CreateKeyboard(node.ID(), buttons), nil)
	return err
}

func (node *FAQNode) NewMessage(user *User, c *Controls, message *vk.Message) (StateNode, bool, error) {
	return nil, false, nil
}

func (node *FAQNode) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (StateNode, bool, error) {
	switch payload.Command {
	case "who":
		_, err := c.Vk.SendMessage(user.id, "Я подрядчик этого дома.", "", nil)
		return nil, false, err
	case "what":
		_, err := c.Vk.SendMessage(user.id, "Я умею отвечать на ваши сообщения и управлять этим домом.", "", nil)
		return nil, false, err
	case "back":
		return nil, true, nil
	}

	return nil, false, nil
}

func (node *FAQNode) Back(user *User, c *Controls, prev_state StateNode) (bool, error) {
	return false, nil
}
