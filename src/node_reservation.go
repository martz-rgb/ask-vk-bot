package main

import (
	"errors"
	"fmt"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/hori-ryota/zaperr"
)

type ReservationNode struct {
	paginator *Paginator[Role]

	role *Role
}

func (node *ReservationNode) ID() string {
	return "reservation"
}

func (node *ReservationNode) Entry(user *User, c *Controls) error {
	roles, err := c.Ask.AvailableRoles()
	if err != nil {
		return err
	}

	to_label := func(role Role) string {
		return role.ShownName
	}
	to_value := func(role Role) string {
		return role.Name
	}

	node.paginator = NewPaginator(
		roles,
		"roles",
		RowsCount,
		ColsCount,
		false,
		to_label,
		to_value)

	message := `Выберите нужную роль с помощи клавиатуры или начните вводить и отправьте часть, с которой начинается имя роли.
				Отправьте специальный символ '%' для того, чтобы вернуться к полному списку ролей.`

	_, err = c.Vk.SendMessage(user.id, message, CreateKeyboard(node, node.paginator.Buttons()), nil)
	return err
}

func (node *ReservationNode) NewMessage(user *User, c *Controls, message *Message) (StateNode, bool, error) {
	roles, err := c.Ask.AvailableRolesStartWith(message.Text)
	if err != nil {
		return nil, false, err
	}

	node.paginator.ChangeObjects(roles)

	return nil, false, c.Vk.ChangeKeyboard(user.id, CreateKeyboard(node, node.paginator.Buttons()))
}

func (node *ReservationNode) KeyboardEvent(user *User, c *Controls, payload *CallbackPayload) (StateNode, bool, error) {
	switch payload.Command {
	case "roles":
		role, err := node.paginator.Object(payload.Value)
		if err != nil {
			return nil, false, err
		}

		node.role = role

		message := fmt.Sprintf("Вы хотите забронировать роль %s?",
			role.ShownName)
		request := &MessageParams{
			Text: "Расскажите про себя в одном сообщении.",
		}

		field := NewField(
			"info",
			request,
			nil,
			ExtractID,
			InfoAboutValidate,
			nil)
		return NewConfirmationNode(message, NewFormNode(field)), false, nil

	case "paginator":
		back := node.paginator.Control(payload.Value)

		if back {
			return nil, true, nil
		}

		return nil, false, c.Vk.ChangeKeyboard(user.id,
			CreateKeyboard(node, node.paginator.Buttons()))
	}

	return nil, false, nil
}

func (node *ReservationNode) Back(user *User, c *Controls, prev_state StateNode) (bool, error) {
	confirmation, ok := prev_state.(*ConfirmationNode)
	if !ok {
		return false, node.Entry(user, c)
	}

	if confirmation.Answer {
		if node.role == nil {
			err := errors.New("no role was chosen to confirm")
			return false, zaperr.Wrap(err, "")
		}

		form, ok := confirmation.Next().(*FormNode)
		if !ok {
			err := errors.New("no form is presented")
			return false, zaperr.Wrap(err, "")
		}

		if !form.IsFilled() {
			return false, node.Entry(user, c)
		}

		values := form.Values()
		id, err := ExtractValue[int](values, "info")
		if err != nil {
			return false, err
		}

		err = c.Ask.AddReservation(node.role.Name, user.id, id)
		if err != nil {
			return false, err
		}

		message := fmt.Sprintf("Отлично! Ваша заявка на бронь %s будет рассмотрена в ближайшее время. Вам придет сообщение.",
			node.role.ShownName)
		forward, err := ForwardParam(user.id, []int{id})
		if err != nil {
			return false, err
		}

		_, err = c.Vk.SendMessage(user.id, message, "", api.Params{
			"forward": forward,
		})
		return true, err
	}

	node.role = nil
	return false, node.Entry(user, c)
}
