package main

import (
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

func (a *VkApi) MarkAsRead(user_id int) {
	a.Lock()
	defer a.Unlock()

	params := api.Params{
		"mark_conversation_as_read": true,
		"peer_id":                   user_id,
	}

	response, err := a.group.MessagesMarkAsRead(params)
	if err != nil {
		logger.Errorw("failed to mark as read vk messsage",
			"error", err,
			"params", params,
			"response", response)
		return
	}

	logger.Debugw("successfully mark as read vk messsage",
		"params", params,
		"response", response)
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
		logger.Errorw("failed to send vk messsage",
			"error", err,
			"params", params,
			"response", response)
		return -1
	}

	logger.Debugw("successfully sent vk messsage",
		"params", params,
		"response", response)

	return response
}

func (a *VkApi) DeleteMessage(peer_id int, message_id int, delete_for_all int) {
	a.Lock()
	defer a.Unlock()

	params := api.Params{
		"peer_id":        peer_id,
		"message_ids":    message_id,
		"delete_for_all": delete_for_all,
	}
	response, err := a.group.MessagesDelete(params)
	if err != nil {
		logger.Errorw("failed to delete vk message",
			"error", err,
			"params", params,
			"response", response)
		return
	}

	logger.Debugw("successfully deleted vk message",
		"params", params,
		"response", response)
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
		logger.Errorw("failed to send vk event answer",
			"error", err,
			"params", params,
			"response", response)
		return
	}

	logger.Debugw("successfully sent vk event answer",
		"params", params,
		"response", response)
}

func (a *VkApi) ChangeKeyboard(user_id int, keyboard string) {
	id := a.SendMessage(user_id, "Меняю клавиатуру", keyboard)
	a.DeleteMessage(user_id, id, 1)
}
