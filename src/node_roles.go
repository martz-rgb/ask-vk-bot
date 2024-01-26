package main

import (
	"errors"
	"fmt"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

var rows_count int = 2
var cols_count int = 3

type RolesNode struct {
	page        int
	total_pages int
	roles       []Role
}

func (node *RolesNode) ID() string {
	return "roles"
}

func (node *RolesNode) Entry(user *User, ask *Ask, vk *VK, params Params) error {
	roles, err := ask.Roles()
	if err != nil {
		return err
	}

	node.roles = roles
	node.page = 0

	keyboard, err := node.CreateRolePage(rows_count, cols_count)
	if err != nil {
		return zaperr.Wrap(err, "failed to create keyboard",
			zap.Any("roles", roles))
	}

	message := `Выберите нужную роль с помощи клавиатуры или начните вводить и отправьте часть, с которой начинается имя роли.
				Отправьте специальный символ '%' для того, чтобы вернуться к полному списку ролей.`

	_, err = vk.SendMessage(user.id, message, keyboard, nil)
	return err
}

func (node *RolesNode) NewMessage(user *User, ask *Ask, vk *VK, message string) (StateNode, error) {
	roles, err := ask.RolesStartWith(message)
	if err != nil {
		return nil, err
	}

	node.roles = roles
	node.page = 0

	keyboard, err := node.CreateRolePage(rows_count, cols_count)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to create keyboard",
			zap.Any("roles", roles))
	}

	return nil, vk.ChangeKeyboard(user.id, keyboard)
}

func (node *RolesNode) KeyboardEvent(user *User, ask *Ask, vk *VK, payload *CallbackPayload) (StateNode, error) {
	switch payload.Command {
	case "roles":
		var info *Role
		for _, role := range node.roles {
			if role.Name == payload.Value {
				info = &role
				break
			}
		}

		if info == nil {
			err := errors.New("failed to find role in list")
			return nil, zaperr.Wrap(err, "",
				zap.String("role", payload.Value),
				zap.Any("roles", node.roles))
		}

		message := fmt.Sprintf("Идентификатор: %s\nТег: %s\nИмя: %s\nЗаголовок: %s\n",
			info.Name, info.Tag, info.ShownName, info.CaptionName.String)

		_, err := vk.SendMessage(user.id, message, "", nil)
		return nil, err
	case "previous":
		node.page -= 1
		if node.page < 0 {
			node.page = 0
		}

		keyboard, err := node.CreateRolePage(rows_count, cols_count)
		if err != nil {
			return nil, zaperr.Wrap(err, "failed to update keyboard",
				zap.Int("page", node.page),
				zap.Any("roles", node.roles))
		}

		return nil, vk.ChangeKeyboard(user.id, keyboard)
	case "next":
		node.page += 1
		if node.page >= node.total_pages {
			node.page = node.total_pages - 1
		}

		keyboard, err := node.CreateRolePage(rows_count, cols_count)
		if err != nil {
			return nil, zaperr.Wrap(err, "failed to update keyboard",
				zap.Int("page", node.page),
				zap.Any("roles", node.roles))
		}

		return nil, vk.ChangeKeyboard(user.id, keyboard)
	case "back":
		return &InitNode{}, nil
	}

	return nil, nil
}

func (node *RolesNode) CreateRolePage(rows int, cols int) (string, error) {
	buttons := [][]Button{}

	cells := rows * cols
	// ceil function
	node.total_pages = 1 + (len(node.roles)-1)/cells

	for i := 0; i < rows; i++ {
		if i*cols >= len(node.roles) {
			break
		}

		buttons = append(buttons, []Button{})

		for j := 0; j < cols; j++ {
			index := i*cols + j + node.page*cells

			if index >= len(node.roles) {
				i = rows
				break
			}

			// maybe send index?..
			buttons[i] = append(buttons[i], Button{
				Label: node.roles[index].ShownName,
				Color: "secondary",

				Command: "roles",
				Value:   node.roles[index].Name,
			})
		}
	}

	// + доп ряд с функциональными кнопками
	controls := []Button{}

	if node.page > 0 {
		controls = append(controls, Button{
			Label: "<<",
			Color: "primary",

			Command: "previous",
		})
	}

	if node.page < node.total_pages-1 {
		controls = append(controls, Button{
			Label: ">>",
			Color: "primary",

			Command: "next",
		})
	}

	controls = append(controls, Button{
		Label: "Назад",
		Color: "negative",

		Command: "back",
	})

	buttons = append(buttons, controls)

	return CreateKeyboard(node, buttons), nil
}
