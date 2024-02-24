package main

import (
	"ask-bot/src/vk"
	"errors"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type AdminNode struct{}

func (node *AdminNode) ID() string {
	return "admin"
}

func (node *AdminNode) Entry(user *User, c *Controls) error {
	ok, err := c.Ask.IsAdmin(user.id)
	if err != nil {
		return err
	}

	if !ok {
		err := errors.New("unathorized access")
		return zaperr.Wrap(err, "",
			zap.Any("user", user))
	}

	buttons := [][]vk.Button{
		{
			{
				Label:   "Брони",
				Color:   vk.SecondaryColor,
				Command: (&AdminReservationNode{}).ID(),
			},
			{
				Label:   "Список ролей",
				Color:   vk.SecondaryColor,
				Command: (&RolesNode{}).ID(),
			},
		},
		{
			{
				Label:   "Назад",
				Color:   vk.NegativeColor,
				Command: "back",
			},
		},
	}

	return c.Vk.ChangeKeyboard(user.id, vk.CreateKeyboard(node.ID(), buttons))
}

func (node *AdminNode) NewMessage(user *User, c *Controls, message *vk.Message) (StateNode, bool, error) {
	return nil, false, nil
}
func (node *AdminNode) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (StateNode, bool, error) {
	switch payload.Command {
	case (&AdminReservationNode{}).ID():
		return &AdminReservationNode{}, false, nil
	case (&RolesNode{}).ID():
		return &RolesNode{}, false, nil
	case "back":
		return nil, true, nil
	}

	return nil, false, nil
}
func (node *AdminNode) Back(user *User, c *Controls, prev_state StateNode) (bool, error) {
	return false, node.Entry(user, c)
}
