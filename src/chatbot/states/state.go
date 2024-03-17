package states

import (
	"ask-bot/src/ask"
	"ask-bot/src/dict"
	"ask-bot/src/postponed"
	"ask-bot/src/vk"
)

type State interface {
	ID() string

	Entry(user *User, c *Controls) error
	NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error)
	KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error)
	Back(user *User, c *Controls, info *ExitInfo) (*Action, error)
}

type User struct {
	Id int
}

type Controls struct {
	Ask       *ask.Ask
	Vk        *vk.VK
	Notify    chan *vk.MessageParams
	Postponed *postponed.Postponed
}

type ExitInfo struct {
	Values  dict.Dictionary
	Payload string
}

type ActionKind int

const (
	NoAction ActionKind = 0
	Next     ActionKind = 1
	Exit     ActionKind = 2
)

type Action struct {
	kind ActionKind

	next State
	exit *ExitInfo
}

func NewActionNext(node State) *Action {
	if node == nil {
		return nil
	}

	return &Action{
		kind: Next,
		next: node,
	}
}

func NewActionExit(exit *ExitInfo) *Action {
	return &Action{
		kind: Exit,
		exit: exit,
	}
}

func (a *Action) Kind() ActionKind {
	if a == nil {
		return NoAction
	}

	return a.kind
}

func (a *Action) Next() State {
	if a == nil {
		return nil
	}
	return a.next
}

func (a *Action) Exit() *ExitInfo {
	if a == nil {
		return nil
	}
	return a.exit
}
