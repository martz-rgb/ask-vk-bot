package main

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/object"
	"go.uber.org/zap"
)

var rows_count int = 2
var cols_count int = 3

type RolesNode struct {
	page        int
	total_pages int
	roles       []Role
}

func (node *RolesNode) String() string {
	return "roles"
}

func (node *RolesNode) Init(a *VkApi, d *Db, user_id int, silent bool) {
	query := "select name, tag, shown_name, caption_name from roles"

	err := d.sql.Select(&node.roles, query)
	if err != nil {
		zap.S().Errorw("failed to select from database",
			"error", err,
			"query", query)
		return
	}

	node.page = 0
	keyboard, err := node.CreateRolePage(rows_count, cols_count)
	if err != nil {
		zap.S().Errorw("failed to create keyboard",
			"error", err,
			"roles", node.roles)
		return
	}

	a.SendMessage(user_id, `Выберите нужную роль с помощи клавиатуры или начните вводить и отправьте часть, с которой начинается имя роли.
							Отправьте специальный символ '%' для того, чтобы вернуться к полному списку ролей.`,
		keyboard.ToJSON())
}

type SearchParam struct {
	ShownName string `db:"shown_name"`
}

func (node *RolesNode) Do(a *VkApi, d *Db, event EventType, i interface{}) StateNode {
	if event == NewMessageEvent {
		obj, ok := i.(events.MessageNewObject)
		if !ok {
			zap.S().Warnw("failed to parse vk response to new message object",
				"object", obj)
			return nil
		}

		query := "select name, tag, shown_name, caption_name from roles where shown_name like ?"

		err := d.sql.Select(&node.roles, query, obj.Message.Text+"%")
		if err != nil {
			zap.S().Errorw("failed to select from database with parameter",
				"error", err,
				"query", query,
				"param", obj.Message.Text+"%")
			return nil
		}

		node.page = 0
		keyboard, err := node.CreateRolePage(rows_count, cols_count)
		if err != nil {
			zap.S().Errorw("failed to create keyboard",
				"error", err,
				"roles", node.roles)
			return nil
		}

		a.ChangeKeyboard(obj.Message.FromID, keyboard.ToJSON())
		return nil
	}

	if event == ChangeKeyboardEvent {
		obj, ok := i.(events.MessageEventObject)
		if !ok {
			zap.S().Warnw("failed to parse vk response to message event object",
				"object", obj)
			return nil
		}

		var payload CallbackPayload

		err := json.Unmarshal([]byte(obj.Payload), &payload)
		if err != nil {
			zap.S().Errorw("failed to unmarshal payload",
				"payload", payload)
			return nil
		}

		if payload.Id != node.String() {
			zap.S().Infow("payload does not belong to node",
				"node", node.String(),
				"payload", payload)
			return nil
		}

		if len(payload.Command) != 0 {
			switch payload.Command {
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

				a.ChangeKeyboard(obj.UserID, keyboard.ToJSON())
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

				a.ChangeKeyboard(obj.UserID, keyboard.ToJSON())
			case "back":
				return &InitNode{}
			}
		}

		if len(payload.Value) != 0 {
			var info Role
			for _, role := range node.roles {
				if role.Name == payload.Value {
					info = role
					break
				}
			}

			if info.Name != payload.Value {
				zap.S().Errorw("failed to find role in list",
					"role", payload.Value,
					"list", node.roles)
				return nil
			}

			role_string := fmt.Sprintf("Идентификатор: %s\nТег: %s\nИмя: %s\nЗаголовок: %s\n",
				info.Name, info.Tag, info.ShownName, info.CaptionName)

			a.SendMessage(obj.UserID, role_string, "")
			return nil
		}

	}

	return nil
}

func (node *RolesNode) CreateRolePage(rows int, cols int) (*object.MessagesKeyboard, error) {
	keyboard := object.NewMessagesKeyboard(false)

	cells := rows * cols
	node.total_pages = int(math.Ceil(float64(len(node.roles)) / float64(cells)))

	for i := 0; i < rows; i++ {
		if i*cols >= len(node.roles) {
			break
		}

		keyboard.AddRow()

		for j := 0; j < cols; j++ {
			index := i*cols + j + node.page*cells

			if index >= len(node.roles) {
				i = rows
				break
			}

			payload := &CallbackPayload{
				Id:    node.String(),
				Value: node.roles[index].Name,
			}

			keyboard.AddCallbackButton(node.roles[index].ShownName, payload, "secondary")
		}
	}

	// + доп ряд с функциональными кнопками
	keyboard.AddRow()

	if node.page > 0 {
		payload := &CallbackPayload{
			Id:      node.String(),
			Command: "previous",
		}

		keyboard.AddCallbackButton("<<", payload, "primary")
	}

	if node.page < node.total_pages-1 {
		payload := &CallbackPayload{
			Id:      node.String(),
			Command: "next",
		}

		keyboard.AddCallbackButton(">>", payload, "primary")
	}

	payload := &CallbackPayload{
		Id:      node.String(),
		Command: "back",
	}
	keyboard.AddCallbackButton("Назад", payload, "negative")

	return keyboard, nil
}
