package main

import (
	"ask-bot/src/ask"
	"ask-bot/src/dict"
	"ask-bot/src/form"
	"ask-bot/src/paginator"
	"ask-bot/src/vk"
	"errors"
	"fmt"

	"github.com/SevereCloud/vksdk/v2/api"
)

type ReservationNode struct {
	paginator *paginator.Paginator[ask.Role]

	role *ask.Role
}

func (node *ReservationNode) ID() string {
	return "reservation"
}

func (node *ReservationNode) Entry(user *User, c *Controls) error {
	roles, err := c.Ask.AvailableRoles()
	if err != nil {
		return err
	}

	to_label := func(role ask.Role) string {
		return role.ShownName
	}
	to_value := func(role ask.Role) string {
		return role.Name
	}

	node.paginator = paginator.New(
		roles,
		"roles",
		paginator.DeafultRows,
		paginator.DefaultCols,
		false,
		to_label,
		to_value)

	message := `Выберите нужную роль с помощи клавиатуры или начните вводить и отправьте часть, с которой начинается имя роли.
				Отправьте специальный символ '%' для того, чтобы вернуться к полному списку ролей.`

	_, err = c.Vk.SendMessage(user.id,
		message,
		vk.CreateKeyboard(node.ID(), node.paginator.Buttons()),
		nil)
	return err
}

func (node *ReservationNode) NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error) {
	roles, err := c.Ask.AvailableRolesStartWith(message.Text)
	if err != nil {
		return nil, err
	}

	node.paginator.ChangeObjects(roles)

	return nil, c.Vk.ChangeKeyboard(user.id, vk.CreateKeyboard(node.ID(), node.paginator.Buttons()))
}

func (node *ReservationNode) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error) {
	switch payload.Command {
	case "roles":
		role, err := node.paginator.Object(payload.Value)
		if err != nil {
			return nil, err
		}
		node.role = role

		message := fmt.Sprintf("Вы хотите забронировать роль %s?",
			role.AccusativeName)

		return NewActionNext(NewConfirmationNode("confirmation", message)), nil

	case "paginator":
		back := node.paginator.Control(payload.Value)

		if back {
			return NewActionExit(&ExitInfo{}), nil
		}

		return nil, c.Vk.ChangeKeyboard(user.id,
			vk.CreateKeyboard(node.ID(), node.paginator.Buttons()))
	}

	return nil, nil
}

func (node *ReservationNode) Back(user *User, c *Controls, info *ExitInfo) (*Action, error) {
	if info == nil {
		return nil, node.Entry(user, c)
	}

	switch info.Payload {
	case "confirmation":
		answer, err := dict.ExtractValue[bool](info.Values, "confirmation")
		if err != nil {
			return nil, err
		}

		if answer {
			request := &vk.MessageParams{
				Text: "Расскажите про себя в одном сообщении.",
			}

			field := form.NewField(
				"about",
				request,
				nil,
				ExtractID,
				InfoAboutValidate,
				nil)

			return NewActionNext(NewFormNode("about", field)), nil
		}

	case "about":
		if node.role == nil {
			return nil, errors.New("no role in node")
		}

		values, err := dict.ExtractValue[dict.Dictionary](info.Values, "form")
		if err != nil {
			return nil, err
		}
		if values == nil {
			return nil, node.Entry(user, c)
		}

		id, err := dict.ExtractValue[int](values, "about")
		if err != nil {
			return nil, err
		}

		err = c.Ask.AddReservation(node.role.Name, user.id, id)
		if err != nil {
			return nil, err
		}

		message := fmt.Sprintf("Отлично! Ваша заявка на бронь %s будет рассмотрена в ближайшее время. Вам придет сообщение.",
			node.role.AccusativeName)
		forward, err := vk.ForwardParam(user.id, []int{id})
		if err != nil {
			return nil, err
		}

		_, err = c.Vk.SendMessage(user.id, message, "", api.Params{
			"forward": forward,
		})
		return NewActionExit(&ExitInfo{}), err
	}

	return nil, node.Entry(user, c)
}
