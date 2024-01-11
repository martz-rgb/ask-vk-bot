package main

type EventType int

const (
	ChangeKeyboardEvent EventType = iota
	NewMessageEvent
)

type StateNode interface {
	String() string

	Init(a *VK, db *DB, user_id int, silent bool)
	Do(a *VK, db *DB, event EventType, i interface{}) StateNode
}

type CallbackPayload struct {
	Id      string `json:"id"`
	Command string `json:"command"`
	Value   string `json:"value"`
}
