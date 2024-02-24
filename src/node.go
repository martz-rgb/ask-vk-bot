package main

import "ask-bot/src/vk"

type StateNode interface {
	ID() string

	Entry(user *User, c *Controls) error
	NewMessage(user *User, c *Controls, message *vk.Message) (StateNode, bool, error)
	KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (StateNode, bool, error)
	Back(user *User, c *Controls, prev_state StateNode) (bool, error)
}
