package states

import "ask-bot/src/vk"

type FAQ struct{}

func (state *FAQ) ID() string {
	return "faq"
}

func (state *FAQ) Entry(user *User, c *Controls) error {
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

	_, err := c.Vk.SendMessage(user.Id, message, vk.CreateKeyboard(state.ID(), buttons), nil)
	return err
}

func (state *FAQ) NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error) {
	return nil, nil
}

func (state *FAQ) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error) {
	switch payload.Command {
	case "who":
		_, err := c.Vk.SendMessage(user.Id, "Я подрядчик этого дома.", "", nil)
		return nil, err
	case "what":
		_, err := c.Vk.SendMessage(user.Id, "Я умею отвечать на ваши сообщения и управлять этим домом.", "", nil)
		return nil, err
	case "back":
		return NewActionExit(nil), nil
	}

	return nil, nil
}

func (state *FAQ) Back(user *User, c *Controls, info *ExitInfo) (*Action, error) {
	return nil, state.Entry(user, c)
}
