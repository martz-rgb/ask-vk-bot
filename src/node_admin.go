package main

import (
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

	buttons := [][]Button{
		{
			{
				Label:   "Брони",
				Color:   SecondaryColor,
				Command: (&AdminReservationNode{}).ID(),
			},
			{
				Label:   "Список ролей",
				Color:   "secondary",
				Command: (&RolesNode{}).ID(),
			},
		},
		{
			{
				Label:   "Назад",
				Color:   NegativeColor,
				Command: "back",
			},
		},
	}

	return c.Vk.ChangeKeyboard(user.id, CreateKeyboard(node, buttons))
}

func (node *AdminNode) NewMessage(user *User, c *Controls, message *Message) (StateNode, bool, error) {
	return nil, false, nil
}
func (node *AdminNode) KeyboardEvent(user *User, c *Controls, payload *CallbackPayload) (StateNode, bool, error) {
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
