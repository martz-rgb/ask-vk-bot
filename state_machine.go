package main

type EventType int

const (
	ChangeKeyboardEvent EventType = iota
	NewMessageEvent
)

type StateNode string

const (
	UndefindedState StateNode = ""
	InitState       StateNode = "Init"
	FaqState        StateNode = "FAQ"
)

type EntryHandler func(a *VkApi, user_id int, silent bool)
type DoHandler func(a *VkApi, event EventType, i interface{}) (next StateNode, change bool)

type State struct {
	Entry EntryHandler
	Do    DoHandler
}

type StateMachine map[StateNode]*State

func (d StateMachine) GetNode(s StateNode) (EntryHandler, DoHandler) {
	node, ok := d[s]
	if !ok {
		return nil, nil
	}
	return node.Entry, node.Do
}
