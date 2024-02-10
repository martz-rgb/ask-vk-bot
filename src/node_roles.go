package main

import (
	"fmt"
)

type RolesNode struct {
	paginator *RolesPaginator
}

func (node *RolesNode) ID() string {
	return "roles"
}

func (node *RolesNode) Entry(user *User, ask *Ask, vk *VK) error {
	roles, err := ask.Roles()
	if err != nil {
		return err
	}

	node.paginator = NewRolesPaginator(roles, RowsCount, ColsCount)

	message := `Выберите нужную роль с помощи клавиатуры или начните вводить и отправьте часть, с которой начинается имя роли.
				Отправьте специальный символ '%' для того, чтобы вернуться к полному списку ролей.`

	_, err = vk.SendMessage(user.id, message, CreateKeyboard(node, node.paginator.Buttons()), nil)
	return err
}

func (node *RolesNode) NewMessage(user *User, ask *Ask, vk *VK, message string) (StateNode, bool, error) {
	roles, err := ask.RolesStartWith(message)
	if err != nil {
		return nil, false, err
	}

	node.paginator.ChangeRoles(roles)

	return nil, false, vk.ChangeKeyboard(user.id, CreateKeyboard(node, node.paginator.Buttons()))
}

func (node *RolesNode) KeyboardEvent(user *User, ask *Ask, vk *VK, payload *CallbackPayload) (StateNode, bool, error) {
	switch payload.Command {
	case "roles":
		role, err := node.paginator.Role(payload.Value)
		if err != nil {
			return nil, false, err
		}

		message := fmt.Sprintf("Идентификатор: %s\nТег: %s\nИмя: %s\nЗаголовок: %s\n",
			role.Name, role.Tag, role.ShownName, role.CaptionName.String)

		_, err = vk.SendMessage(user.id, message, "", nil)
		return nil, false, err
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

func (node *RolesNode) Back(user *User, ask *Ask, vk *VK, prev_state StateNode) (bool,  error) {
	return false, nil
}
