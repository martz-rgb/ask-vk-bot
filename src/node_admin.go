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

func (node *AdminNode) Entry(user *User, ask *Ask, vk *VK) error {
	ok, err := ask.IsAdmin(user.id)
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
				Command: "reservation",
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

	return vk.ChangeKeyboard(user.id, CreateKeyboard(node, buttons))
}

func (node *AdminNode) NewMessage(user *User, ask *Ask, vk *VK, message *Message) (StateNode, bool, error) {
	return nil, false, nil
}
func (node *AdminNode) KeyboardEvent(user *User, ask *Ask, vk *VK, payload *CallbackPayload) (StateNode, bool, error) {
	switch payload.Command {
	case "reservation":
		return &AdminReservationNode{}, false, nil
	case "back":
		return nil, true, nil
	}

	return nil, false, nil
}
func (node *AdminNode) Back(user *User, ask *Ask, vk *VK, prev_state StateNode) (bool, error) {
	return false, node.Entry(user, ask, vk)
}
