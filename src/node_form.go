package main

import (
	"ask-bot/src/form"
	"ask-bot/src/vk"
	"errors"

	"github.com/hori-ryota/zaperr"
)

type FormNode struct {
	f *form.Form
}

func NewFormNode(fields ...*form.Field) *FormNode {
	return &FormNode{
		f: form.NewForm(fields...),
	}
}

func (node *FormNode) ID() string {
	return "form"
}

func (node *FormNode) Entry(user *User, c *Controls) error {
	if node.f == nil {
		err := errors.New("no form is provided")
		return zaperr.Wrap(err, "")
	}

	return node.sendRequest(user, c)
}

func (node *FormNode) NewMessage(user *User, c *Controls, message *vk.Message) (StateNode, bool, error) {
	end, err := node.set(user, c, message)
	return nil, end, err
}

func (node *FormNode) KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (StateNode, bool, error) {
	switch payload.Command {
	case "form":
		end, err := node.set(user, c, payload.Value)
		return nil, end, err

	case "paginator":
		back := node.f.Control(payload.Value)

		if back {
			return nil, true, nil
		}

		return nil, false, c.Vk.ChangeKeyboard(user.id,
			vk.CreateKeyboard(node.ID(), node.f.Buttons()))
	}

	return nil, false, nil
}

func (node *FormNode) Back(user *User, c *Controls, prev_state StateNode) (bool, error) {
	return false, nil
}

func (node *FormNode) sendRequest(user *User, c *Controls) error {
	request := node.f.Request()

	_, err := c.Vk.SendMessageParams(
		user.id,
		request,
		vk.CreateKeyboard(node.ID(), node.f.Buttons()))
	return err
}

func (node *FormNode) set(user *User, c *Controls, input interface{}) (end bool, err error) {
	var info *vk.MessageParams

	switch value := input.(type) {
	case *vk.Message:
		node.f.SetFromMessage(value)
	case string:
		node.f.SetFromOption(value)
	}

	if err != nil {
		return false, err
	}

	if info != nil {
		_, err = c.Vk.SendMessage(user.id, info.Text, "", info.Params)
		return false, err
	}

	end = node.f.Next()
	if !end {
		return false, node.sendRequest(user, c)
	}
	return true, nil
}

func (node *FormNode) Values() map[string]interface{} {
	return node.f.Values()
}

func (node *FormNode) IsFilled() bool {
	return node.f.Values() != nil
}
