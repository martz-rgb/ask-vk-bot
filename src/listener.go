package main

import "context"

type Listener struct {
	ask *Ask

	group *VK
	admin *VK
}

func NewListener(ask *Ask, group *VK, admin *VK) *Listener {
	return &Listener{
		ask:   ask,
		group: group,
		admin: admin,
	}
}

func (l *Listener) RunLongPoll(ctx context.Context) {

}
