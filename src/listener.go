package main

import "context"

type Listener struct {
	group *VK
	admin *VK

	db *DB
}

func NewListener(group *VK, admin *VK, db *DB) *Listener {
	return &Listener{
		group: group,
		admin: admin,
		db:    db,
	}
}

func (l *Listener) RunLongPoll(ctx context.Context) {

}
