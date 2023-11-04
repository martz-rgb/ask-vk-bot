package main

type EventType int

const (
	ChangeKeyboardEvent EventType = iota
	NewMessageEvent
)

type StateNode interface {
	String() string

	Init(a *VkApi, db *Db, user_id int, silent bool)
	Do(a *VkApi, db *Db, event EventType, i interface{}) StateNode
}

type CallbackPayload struct {
	Id      string `json:"id"`
	Command string `json:"command"`
	Value   string `json:"value"`
}
