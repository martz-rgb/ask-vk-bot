package main

type EventType int

const (
	ChangeKeyboardEvent EventType = iota
	NewMessageEvent
)

type StateNode interface {
	String() string

	Entry(ask *Ask, vk *VK, user_id int, silent bool)
	Do(ask *Ask, vk *VK, event EventType, i interface{}) StateNode
}

type CallbackPayload struct {
	Id      string `json:"id"`
	Command string `json:"command"`
	Value   string `json:"value"`
}
