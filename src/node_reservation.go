package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/hori-ryota/zaperr"
)

type ReservationNode struct {
	paginator *RolesPaginator

	role *Role
}

func (node *ReservationNode) ID() string {
	return "reservation"
}

func (node *ReservationNode) Entry(user *User, ask *Ask, vk *VK) error {
	roles, err := ask.AvailableRoles()
	if err != nil {
		return err
	}

	node.paginator = NewRolesPaginator(roles, RowsCount, ColsCount)

	message := `Выберите нужную роль с помощи клавиатуры или начните вводить и отправьте часть, с которой начинается имя роли.
				Отправьте специальный символ '%' для того, чтобы вернуться к полному списку ролей.`

	_, err = vk.SendMessage(user.id, message, CreateKeyboard(node, node.paginator.Buttons()), nil)
	return err
}

func (node *ReservationNode) NewMessage(user *User, ask *Ask, vk *VK, message string) (StateNode, bool, error) {
	roles, err := ask.AvailableRolesStartWith(message)
	if err != nil {
		return nil, false, err
	}

	node.paginator.ChangeRoles(roles)

	return nil, false, vk.ChangeKeyboard(user.id, CreateKeyboard(node, node.paginator.Buttons()))
}

func (node *ReservationNode) KeyboardEvent(user *User, ask *Ask, vk *VK, payload *CallbackPayload) (StateNode, bool, error) {
	switch payload.Command {
	case "roles":
		role, err := node.paginator.Role(payload.Value)
		if err != nil {
			return nil, false, err
		}

		node.role = role
		message := fmt.Sprintf("Вы хотите забронировать роль %s?",
			role.ShownName)

		return &ConfirmationNode{
			Message: message,
		}, false, err
	case "previous":
		node.paginator.Previous()

		return nil, false, vk.ChangeKeyboard(user.id, CreateKeyboard(node, node.paginator.Buttons()))
	case "next":
		node.paginator.Next()

		return nil, false, vk.ChangeKeyboard(user.id, CreateKeyboard(node, node.paginator.Buttons()))
	case "back":
		return nil, true, nil
	}

	return nil, false, nil
}

func (node *ReservationNode) Back(user *User, ask *Ask, vk *VK, prev_state StateNode) error {
	confirmation, ok := prev_state.(*ConfirmationNode)
	if !ok {
		return nil
	}

	if confirmation.Answer {
		if node.role == nil {
			err := errors.New("no role was chosen to confirm")
			return zaperr.Wrap(err, "")
		}

		deadline, err := ask.AddReservation(node.role.Name, user.id)
		if err != nil {
			return err
		}

		message := fmt.Sprintf("Роль %s забронирована до %s.",
			node.role.ShownName,
			deadline.Format(time.DateTime),
		)

		_, err = vk.SendMessage(user.id, message, "", nil)
		return err
	}

	node.role = nil
	return node.Entry(user, ask, vk)
}
