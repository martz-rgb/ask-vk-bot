package states

import (
	"ask-bot/src/dict"
	"ask-bot/src/vk"
)

type Confirmation struct {
	message *vk.MessageParams

	payload string
}

func NewConfirmation(payload string, message *vk.MessageParams) *Confirmation {
	return &Confirmation{
		message: message,
		payload: payload,
	}
}

func (state *Confirmation) ID() string {
	return "confirmation"
}

func (state *Confirmation) Entry(user *User, c *Controls) error {
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

	_, err := c.Vk.SendMessageParams(user.Id, state.message, vk.CreateKeyboard(state.ID(), buttons))
	return err
}

func (state *Confirmation) NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error) {
	return nil, nil
}

func (state *Confirmation) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error) {
	switch payload.Command {
	case "yes":
		return state.exit_action(true), nil

	case "no":
		return state.exit_action(false), nil

	}

	return nil, nil
}

func (state *Confirmation) Back(user *User, c *Controls, info *ExitInfo) (*Action, error) {
	return nil, state.Entry(user, c)
}

func (state *Confirmation) exit_action(answer bool) *Action {
	return NewActionExit(&ExitInfo{
		Values: dict.Dictionary{
			"confirmation": answer,
		},
		Payload: state.payload,
	})
}
