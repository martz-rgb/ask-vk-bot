package main

import (
	"ask-bot/src/dict"
	"ask-bot/src/vk"
)

type StateNode interface {
	ID() string

	Entry(user *User, c *Controls) error
	NewMessage(user *User, c *Controls, message *vk.Message) (*Action, error)
	KeyboardEvent(user *User, c *Controls, payload *vk.CallbackPayload) (*Action, error)
	Back(user *User, c *Controls, info *ExitInfo) (*Action, error)
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

	next StateNode
	exit *ExitInfo
}

func NewActionNext(node StateNode) *Action {
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

func (a *Action) Next() StateNode {
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
