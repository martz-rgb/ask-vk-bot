package main

import (
	"fmt"

	"github.com/SevereCloud/vksdk/v2/events"
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

func (node *RolesNode) Entry(user_id int, ask *Ask, vk *VK, silent bool) {
	roles, err := ask.Roles()
	if err != nil {
		zap.S().Errorw("failed to get roles from ask",
			"error", err)
		return
	}

	node.roles = roles
	node.page = 0

	keyboard, err := node.CreateRolePage(rows_count, cols_count)
	if err != nil {
		zap.S().Errorw("failed to create keyboard",
			"error", err,
			"roles", node.roles)
		return
	}

	message := `Выберите нужную роль с помощи клавиатуры или начните вводить и отправьте часть, с которой начинается имя роли.
				Отправьте специальный символ '%' для того, чтобы вернуться к полному списку ролей.`

	vk.SendMessage(user_id, message, keyboard)
}

func (node *RolesNode) Do(user_id int, ask *Ask, vk *VK, input interface{}) StateNode {
	switch obj := input.(type) {

	case events.MessageNewObject:
		return node.NewMessage(ask, vk, obj.Message.FromID, obj.Message.Text)

	case events.MessageEventObject:
		payload, err := UnmarshalPayload(node, obj.Payload)
		if err != nil {
			zap.S().Errorw("failed to unmarshal payload",
				"payload", payload)
			return nil
		}

		return node.KeyboardEvent(ask, vk, obj.UserID, payload)

	default:
		zap.S().Warnw("failed to parse vk response to message event object",
			"object", obj)
	}

	return nil
}

func (node *RolesNode) NewMessage(ask *Ask, vk *VK, user_id int, message string) StateNode {
	roles, err := ask.RolesStartWith(message)
	if err != nil {
		zap.S().Errorw("failed to get roles start with from ask",
			"error", err,
			"start with", message)
		return nil
	}

	node.roles = roles
	node.page = 0

	keyboard, err := node.CreateRolePage(rows_count, cols_count)
	if err != nil {
		zap.S().Errorw("failed to create keyboard",
			"error", err,
			"roles", node.roles)
		return nil
	}

	vk.ChangeKeyboard(user_id, keyboard)
	return nil
}

func (node *RolesNode) KeyboardEvent(ask *Ask, vk *VK, user_id int, payload *CallbackPayload) StateNode {
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
			zap.S().Errorw("failed to find role in list",
				"role", payload.Value,
				"list", node.roles)
			return nil
		}

		message := fmt.Sprintf("Идентификатор: %s\nТег: %s\nИмя: %s\nЗаголовок: %s\n",
			info.Name, info.Tag, info.ShownName, info.CaptionName)

		vk.SendMessage(user_id, message, "")

		return nil
	case "previous":
		node.page -= 1
		if node.page < 0 {
			node.page = 0
		}

		keyboard, err := node.CreateRolePage(rows_count, cols_count)
		if err != nil {
			zap.S().Errorw("failed to update keyboard",
				"error", err,
				"page", node.page,
				"list", node.roles)
			return nil
		}

		vk.ChangeKeyboard(user_id, keyboard)

		return nil
	case "next":
		node.page += 1
		if node.page >= node.total_pages {
			node.page = node.total_pages - 1
		}

		keyboard, err := node.CreateRolePage(rows_count, cols_count)
		if err != nil {
			zap.S().Errorw("failed to update keyboard",
				"error", err,
				"page", node.page,
				"list", node.roles)
			return nil
		}

		vk.ChangeKeyboard(user_id, keyboard)

		return nil
	case "back":
		return &InitNode{}
	}

	return nil
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
