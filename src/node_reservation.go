package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
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

func (node *ReservationNode) NewMessage(user *User, ask *Ask, vk *VK, message *Message) (StateNode, bool, error) {
	roles, err := ask.AvailableRolesStartWith(message.Text)
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

		request := "Расскажите про себя в одном сообщении."
		form := NewForm([]FormField{NewAboutField(request)})

		return &ConfirmationNode{
			Message: message,
			Next: &FormNode{
				Form: form,
			},
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

func (node *ReservationNode) Back(user *User, ask *Ask, vk *VK, prev_state StateNode) (bool, error) {
	confirmation, ok := prev_state.(*ConfirmationNode)
	if !ok {
		return false, nil
	}

	if confirmation.Answer {
		if node.role == nil {
			err := errors.New("no role was chosen to confirm")
			return false, zaperr.Wrap(err, "")
		}

		form, ok := confirmation.Next.(*FormNode)
		if !ok {
			err := errors.New("no form is presented")
			return false, zaperr.Wrap(err, "")
		}

		if !form.FilledOut {
			node.role = nil
			return false, node.Entry(user, ask, vk)
		}

		value, err := form.Form.Value(0)
		if err != nil {
			err := errors.New("form is not fullfilled")
			return false, zaperr.Wrap(err, "",
				zap.Any("form", form.Form))
		}

		id, err := ConvertValue[int](value)
		if err != nil {
			return false, err
		}

		deadline, err := ask.AddReservation(node.role.Name, user.id, id)
		if err != nil {
			return false, err
		}

		message := fmt.Sprintf("Роль %s забронирована до %s.",
			node.role.ShownName,
			deadline.Format(time.DateTime),
		)

		// forward := fmt.Sprintf(`
		// 	{
		// 		"peer_id": %d,
		// 		"message_ids": [%d]
		// 	}
		// `, user.id, id)

		_, err = vk.SendMessage(user.id, message, "", nil)
		return true, err
	}

	node.role = nil
	return false, node.Entry(user, ask, vk)
}
