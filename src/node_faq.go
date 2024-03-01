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

func (node *FAQNode) NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error) {
	return nil, nil
}

func (node *FAQNode) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error) {
	switch payload.Command {
	case "who":
		_, err := c.Vk.SendMessage(user.id, "Я подрядчик этого дома.", "", nil)
		return nil, err
	case "what":
		_, err := c.Vk.SendMessage(user.id, "Я умею отвечать на ваши сообщения и управлять этим домом.", "", nil)
		return nil, err
	case "back":
		return NewActionExit(nil), nil
	}

	return nil, nil
}

func (node *FAQNode) Back(user *User, c *Controls, info *ExitInfo) (*Action, error) {
	return nil, node.Entry(user, c)
}
