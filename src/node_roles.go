package main

import (
	"ask-bot/src/ask"
	"ask-bot/src/paginator"
	"ask-bot/src/vk"
	"fmt"
)

type RolesNode struct {
	paginator *paginator.Paginator[ask.Role]
}

func (node *RolesNode) ID() string {
	return "roles"
}

func (node *RolesNode) Entry(user *User, c *Controls) error {
	roles, err := c.Ask.Roles()
	if err != nil {
		return err
	}

	to_label := func(role ask.Role) string {
		return role.ShownName
	}
	to_value := func(role ask.Role) string {
		return role.Name
	}

	node.paginator = paginator.New[ask.Role](
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

func (node *RolesNode) NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error) {
	roles, err := c.Ask.RolesStartWith(message.Text)
	if err != nil {
		return nil, err
	}

	node.paginator.ChangeObjects(roles)

	return nil, c.Vk.ChangeKeyboard(user.id, vk.CreateKeyboard(node.ID(), node.paginator.Buttons()))
}

func (node *RolesNode) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error) {
	switch payload.Command {
	case "roles":
		role, err := node.paginator.Object(payload.Value)
		if err != nil {
			return nil, err
		}

		message := fmt.Sprintf("Идентификатор: %s\nТег: %s\nИмя: %s\nЗаголовок: %s\n",
			role.Name, role.Tag, role.ShownName, role.CaptionName.String)

		_, err = c.Vk.SendMessage(user.id, message, "", nil)
		return nil, err
	case "paginator":
		back := node.paginator.Control(payload.Value)

		if back {
			return NewActionExit(nil), nil
		}

		return nil, c.Vk.ChangeKeyboard(user.id,
			vk.CreateKeyboard(node.ID(), node.paginator.Buttons()))
	}

	return nil, nil
}

func (node *RolesNode) Back(user *User, c *Controls, info *ExitInfo) (*Action, error) {
	return nil, node.Entry(user, c)
}
