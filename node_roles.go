package main

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/object"
)

type RolesNode struct {
	page        int
	total_pages int
	roles       []Role
}

func (node *RolesNode) String() string {
	return "roles"
}

func (node *RolesNode) Init(a *VkApi, d *Db, user_id int, silent bool) {
	err := d.sql.Select(&node.roles, "select name, tag, shown_name, caption_name from roles")
	if err != nil {
		fmt.Println("unable to query roles table: ", err)
		return
	}

	node.page = 0
	keyboard, err := node.CreateRolePage(2, 3)
	if err != nil {
		return
	}

	a.SendMessage(user_id, "Выберите нужную роль с помощи клавиатуры: ", keyboard.ToJSON())
}

func (node *RolesNode) Do(a *VkApi, db *Db, event EventType, i interface{}) StateNode {
	if event == NewMessageEvent {
		// to-do search
		return nil
	}

	if event == ChangeKeyboardEvent {
		obj, ok := i.(events.MessageEventObject)
		if !ok {
			return nil
		}

		var payload CallbackPayload

		err := json.Unmarshal([]byte(obj.Payload), &payload)
		if err != nil {
			fmt.Println("unable to decode json: ", err)
			return nil
		}

		if len(payload.Command) != 0 {
			switch payload.Command {
			case "previous":
				node.page -= 1

				keyboard, err := node.CreateRolePage(2, 3)
				if err != nil {
					return nil
				}

				a.ChangeKeyboard(obj.UserID, keyboard.ToJSON())
			case "next":
				node.page += 1

				keyboard, err := node.CreateRolePage(2, 3)
				if err != nil {
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

			role_string := fmt.Sprintf(`
					Идентификатор: %s,
					Тег: %s
					Имя: %s,
					Заголовок: %s,
			`, info.Name, info.Tag, info.Shown_name, info.Caption_name)

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
		keyboard.AddRow()

		for j := 0; j < cols; j++ {
			index := i*rows + j + node.page*cells

			if index >= len(node.roles) {
				break
			}

			payload := &CallbackPayload{
				Id:    node.String(),
				Value: node.roles[index].Name,
			}

			keyboard.AddCallbackButton(node.roles[index].Shown_name, payload, "secondary")
		}
	}

	// + доп ряд с функциональными кнопками
	keyboard.AddRow()

	// кнопка <-
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
