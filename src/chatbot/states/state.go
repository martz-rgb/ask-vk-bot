package states

import (
	"ask-bot/src/ask"
	"ask-bot/src/dict"
	"ask-bot/src/vk"
	"ask-bot/src/watcher/postponed"
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
	NoAction ActionKind = iota
	Next
	Exit
)

type Action struct {
	value interface{}
}

func NewActionNext(node State) *Action {
	return &Action{node}
}

func NewActionExit(exit *ExitInfo) *Action {
	return &Action{exit}
}

func (a *Action) Kind() ActionKind {
	switch a.value.(type) {
	case State:
		return Next
	case *ExitInfo:
		return Exit
	default:
		return NoAction
	}
}

func (a *Action) Next() State {
	node, ok := a.value.(State)
	if !ok {
		return nil
	}

	return node
}

func (a *Action) Exit() *ExitInfo {
	exit, ok := a.value.(*ExitInfo)
	if !ok {
		return nil
	}

	return exit
}
