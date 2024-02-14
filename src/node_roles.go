package main

import (
	"fmt"
)

type RolesNode struct {
	paginator *Paginator[Role]
}

func (node *RolesNode) ID() string {
	return "roles"
}

func (node *RolesNode) Entry(user *User, ask *Ask, vk *VK) error {
	roles, err := ask.Roles()
	if err != nil {
		return err
	}

	to_label := func(role Role) string {
		return role.ShownName
	}
	to_value := func(role Role) string {
		return role.Name
	}

	node.paginator = NewPaginator[Role](roles, "roles", RowsCount, ColsCount, to_label, to_value)

	message := `Выберите нужную роль с помощи клавиатуры или начните вводить и отправьте часть, с которой начинается имя роли.
				Отправьте специальный символ '%' для того, чтобы вернуться к полному списку ролей.`

	_, err = vk.SendMessage(user.id, message, CreateKeyboard(node, node.paginator.Buttons()), nil)
	return err
}

func (node *RolesNode) NewMessage(user *User, ask *Ask, vk *VK, message *Message) (StateNode, bool, error) {
	roles, err := ask.RolesStartWith(message.Text)
	if err != nil {
		return nil, false, err
	}

	node.paginator.ChangeObjects(roles)

	return nil, false, vk.ChangeKeyboard(user.id, CreateKeyboard(node, node.paginator.Buttons()))
}

func (node *RolesNode) KeyboardEvent(user *User, ask *Ask, vk *VK, payload *CallbackPayload) (StateNode, bool, error) {
	switch payload.Command {
	case "roles":
		role, err := node.paginator.Object(payload.Value)
		if err != nil {
			return nil, false, err
		}

		message := fmt.Sprintf("Идентификатор: %s\nТег: %s\nИмя: %s\nЗаголовок: %s\n",
			role.Name, role.Tag, role.ShownName, role.CaptionName.String)

		_, err = vk.SendMessage(user.id, message, "", nil)
		return nil, false, err
	case "paginator":
		back := node.paginator.Control(payload.Value)

		if back {
			return nil, true, nil
		}

		return nil, false, vk.ChangeKeyboard(user.id,
			CreateKeyboard(node, node.paginator.Buttons()))
	}

	return nil, false, nil
}

func (node *RolesNode) Back(user *User, ask *Ask, vk *VK, prev_state StateNode) (bool, error) {
	return false, nil
}
