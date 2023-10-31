package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"
)

type VkApi struct {
	sync.Mutex

	r *rand.Rand

	// api.VK is based on http.Client and http.Client is claimed to be concurrency safe
	group *api.VK
	admin *api.VK
}

func NewVkApi(group_token string, admin_token string) (*VkApi, error) {
	a := &VkApi{}
	source := rand.NewSource(time.Now().UnixNano())
	a.r = rand.New(source)

	a.group = api.NewVK(group_token)
	a.admin = api.NewVK(admin_token)

	return a, nil
}

func (a *VkApi) SendMessage(user_id int, message string, keyboard string) int {
	a.Lock()
	defer a.Unlock()

	params := api.Params{
		"user_id":   user_id,
		"random_id": a.r.Int(),
		"message":   message,
		"keyboard":  keyboard,
	}

	response, err := a.group.MessagesSend(params)
	if err != nil {
		fmt.Println("err", err)
		return -1
	}

	fmt.Println("send response", response)
	return response
}

func (a *VkApi) DeleteMessage(peer_id int, message_id int, delete_for_all int) {
	a.Lock()
	defer a.Unlock()

	delete_params := api.Params{
		"peer_id":        peer_id,
		"message_ids":    message_id,
		"delete_for_all": delete_for_all,
	}
	response, err := a.group.MessagesDelete(delete_params)
	if err != nil {
		fmt.Println("err", err)
		return
	}
	fmt.Println("delete response", response)
}

func (a *VkApi) SendEventAnswer(event_id string, user_id int, peer_id int) {
	a.Lock()
	defer a.Unlock()

	params := api.Params{
		"event_id":  event_id,
		"user_id":   user_id,
		"random_id": a.r.Int(),
		"peer_id":   []int{peer_id},
	}

	response, err := a.group.MessagesSendMessageEventAnswer(params)
	if err != nil {
		fmt.Println("err", err)
		return
	}
	fmt.Println("event response", response)
}
