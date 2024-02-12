package main

import (
	"errors"
	"fmt"

	"github.com/hori-ryota/zaperr"
)

type FormNode struct {
	Form      *Form
	FilledOut bool
}

func (node *FormNode) ID() string {
	return "form"
}

func (node *FormNode) Entry(user *User, ask *Ask, vk *VK) error {
	if node.Form == nil {
		err := errors.New("no form is provided")
		return zaperr.Wrap(err, "")
	}

	request, err := node.Form.Request()
	if err != nil {
		return err
	}

	_, err = vk.SendMessage(user.id, request, CreateKeyboard(node, node.Form.Buttons()), nil)
	return err
}

func (node *FormNode) NewMessage(user *User, ask *Ask, vk *VK, message *Message) (StateNode, bool, error) {
	ok, user_error, err := node.Form.SetAndValidate(message)
	if err != nil {
		return nil, false, err
	}

	if !ok {
		text := fmt.Sprintf("Поле не корректно: %s", user_error)
		_, err = vk.SendMessage(user.id, text, "", nil)
		return nil, false, err
	}

	end := node.Form.Next()
	if !end {
		request, err := node.Form.Request()
		if err != nil {
			return nil, false, err
		}

		_, err = vk.SendMessage(user.id,
			request,
			CreateKeyboard(node, node.Form.Buttons()),
			nil)
		return nil, false, err
	}

	node.FilledOut = true
	return nil, true, nil
}

func (node *FormNode) KeyboardEvent(user *User, ask *Ask, vk *VK, payload *CallbackPayload) (StateNode, bool, error) {
	switch payload.Command {
	case "previous":
		node.Form.Previous()
		request, err := node.Form.Request()
		if err != nil {
			return nil, false, err
		}

		_, err = vk.SendMessage(user.id,
			request,
			CreateKeyboard(node, node.Form.Buttons()),
			nil)
		return nil, false, err
	case "back":
		return nil, true, nil
	case "empty":
		ok, user_error, err := node.Form.SetAndValidate(nil)
		if err != nil {
			return nil, false, err
		}

		if !ok {
			text := fmt.Sprintf("Поле не корректно: %s", user_error)
			_, err = vk.SendMessage(user.id, text, "", nil)
			return nil, false, err
		}

		end := node.Form.Next()
		if !end {
			request, err := node.Form.Request()
			if err != nil {
				return nil, false, err
			}

			_, err = vk.SendMessage(user.id,
				request,
				CreateKeyboard(node, node.Form.Buttons()),
				nil)
			return nil, false, err
		}

		node.FilledOut = true
		return nil, true, nil
	}

	return nil, false, nil
}

func (node *FormNode) Back(user *User, ask *Ask, vk *VK, prev_state StateNode) (bool, error) {
	return false, nil
}
