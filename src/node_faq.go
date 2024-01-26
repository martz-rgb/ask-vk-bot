package main

type FAQNode struct{}

func (node *FAQNode) ID() string {
	return "faq"
}

func (node *FAQNode) Entry(user *User, ask *Ask, vk *VK, params Params) error {
	buttons := [][]Button{{
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

	_, err := vk.SendMessage(user.id, message, CreateKeyboard(node, buttons), nil)
	return err
}

func (node *FAQNode) NewMessage(user *User, ask *Ask, vk *VK, message string) (StateNode, error) {
	return nil, nil
}

func (node *FAQNode) KeyboardEvent(user *User, ask *Ask, vk *VK, payload *CallbackPayload) (StateNode, error) {
	switch payload.Command {
	case "who":
		_, err := vk.SendMessage(user.id, "Я подрядчик этого дома.", "", nil)
		return nil, err
	case "what":
		_, err := vk.SendMessage(user.id, "Я умею отвечать на ваши сообщения и управлять этим домом.", "", nil)
		return nil, err
	case "back":
		return &InitNode{}, nil
	}

	return nil, nil
}
