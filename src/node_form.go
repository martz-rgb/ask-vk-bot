package main

import (
	"errors"

	"github.com/hori-ryota/zaperr"
)

type FormNode struct {
	form *Form
}

func NewFormNode(fields ...*Field) *FormNode {
	return &FormNode{
		form: NewForm(fields...),
	}
}

func (node *FormNode) ID() string {
	return "form"
}

func (node *FormNode) Entry(user *User, c *Controls) error {
	if node.form == nil {
		err := errors.New("no form is provided")
		return zaperr.Wrap(err, "")
	}

	return node.sendRequest(user, c)
}

func (node *FormNode) NewMessage(user *User, c *Controls, message *Message) (StateNode, bool, error) {
	end, err := node.set(user, c, message)
	return nil, end, err
}

func (node *FormNode) KeyboardEvent(user *User, c *Controls, payload *CallbackPayload) (StateNode, bool, error) {
	switch payload.Command {
	case "form":
		end, err := node.set(user, c, payload.Value)
		return nil, end, err

	case "paginator":
		back := node.form.Control(payload.Value)

		if back {
			return nil, true, nil
		}

		return nil, false, c.Vk.ChangeKeyboard(user.id,
			CreateKeyboard(node, node.form.Buttons()))
	}

	return nil, false, nil
}

func (node *FormNode) Back(user *User, c *Controls, prev_state StateNode) (bool, error) {
	return false, nil
}

func (node *FormNode) sendRequest(user *User, c *Controls) error {
	request := node.form.Request()

	_, err := c.Vk.SendMessage(
		user.id,
		request.Text,
		CreateKeyboard(node, node.form.Buttons()),
		request.Params)
	return err
}

func (node *FormNode) set(user *User, c *Controls, input interface{}) (end bool, err error) {
	var info *MessageParams

	switch value := input.(type) {
	case *Message:
		node.form.SetFromMessage(value)
	case string:
		node.form.SetFromOption(value)
	}

	if err != nil {
		return false, err
	}

	if info != nil {
		_, err = c.Vk.SendMessage(user.id, info.Text, "", info.Params)
		return false, err
	}

	end = node.form.Next()
	if !end {
		return false, node.sendRequest(user, c)
	}
	return true, nil
}

func (node *FormNode) Values() map[string]interface{} {
	return node.form.Values()
}

func (node *FormNode) IsFilled() bool {
	return node.form.Values() != nil
}
