package main

import (
	"math/rand"
	"strings"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/awnumar/memguard"
	"go.uber.org/zap"
)

type VkApi struct {
	r *rand.Rand

	// api.VK is based on http.Client and http.Client is claimed to be concurrency safe
	group *api.VK
	admin *api.VK
}

func NewVkApi(group_token *memguard.LockedBuffer, admin_token *memguard.LockedBuffer) (*VkApi, error) {
	a := &VkApi{}
	source := rand.NewSource(time.Now().UnixNano())
	a.r = rand.New(source)

	// should copy strings because VkApi saves them inside and use,
	//  but i destroy LockedBuffers with pointers on strings
	a.group = api.NewVK(strings.Clone(group_token.String()))
	a.admin = api.NewVK(strings.Clone(admin_token.String()))

	return a, nil
}

func (a *VkApi) MarkAsRead(user_id int) {
	params := api.Params{
		"mark_conversation_as_read": true,
		"peer_id":                   user_id,
	}

	response, err := a.group.MessagesMarkAsRead(params)
	if err != nil {
		zap.S().Errorw("failed to mark as read vk messsage",
			"error", err,
			"params", params,
			"response", response)
		return
	}

	zap.S().Debugw("successfully mark as read vk messsage",
		"params", params,
		"response", response)
}

func (a *VkApi) SendMessage(user_id int, message string, keyboard string) int {
	params := api.Params{
		"user_id":   user_id,
		"random_id": a.r.Int(),
		"message":   message,
		"keyboard":  keyboard,
	}

	response, err := a.group.MessagesSend(params)
	if err != nil {
		zap.S().Errorw("failed to send vk messsage",
			"error", err,
			"params", params,
			"response", response)
		return -1
	}

	zap.S().Debugw("successfully sent vk messsage",
		"params", params,
		"response", response)

	return response
}

func (a *VkApi) EditMessage(peer_id int, message_id int, message string, keyboard string) {
	params := api.Params{
		"peer_id":    peer_id,
		"message_id": message_id,
		"message":    message,
		"keyboard":   keyboard,
	}

	response, err := a.group.MessagesEdit(params)
	if err != nil {
		zap.S().Errorw("failed to edit vk message",
			"error", err,
			"params", params,
			"response", response)
		return
	}

	zap.S().Debugw("successfully edited vk message",
		"params", params,
		"response", response)
}

func (a *VkApi) DeleteMessage(peer_id int, message_id int, delete_for_all int) {
	params := api.Params{
		"peer_id":        peer_id,
		"message_ids":    message_id,
		"delete_for_all": delete_for_all,
	}
	response, err := a.group.MessagesDelete(params)
	if err != nil {
		zap.S().Errorw("failed to delete vk message",
			"error", err,
			"params", params,
			"response", response)
		return
	}

	zap.S().Debugw("successfully deleted vk message",
		"params", params,
		"response", response)
}

func (a *VkApi) SendEventAnswer(event_id string, user_id int, peer_id int) {
	params := api.Params{
		"event_id":  event_id,
		"user_id":   user_id,
		"random_id": a.r.Int(),
		"peer_id":   []int{peer_id},
	}

	response, err := a.group.MessagesSendMessageEventAnswer(params)
	if err != nil {
		zap.S().Errorw("failed to send vk event answer",
			"error", err,
			"params", params,
			"response", response)
		return
	}

	zap.S().Debugw("successfully sent vk event answer",
		"params", params,
		"response", response)
}

func (a *VkApi) GetLastBotMessage(user_id int) *object.MessagesMessage {
	params := api.Params{
		"count":   20, // heuristic value
		"user_id": user_id,
	}

	response, err := a.group.MessagesGetHistory(params)
	if err != nil {
		zap.S().Errorw("failed to get last bot message",
			"error", err,
			"params", params,
			"response", response)
		return nil
	}

	for _, message := range response.Items {
		if message.Out {
			return &message
		}
	}

	return nil
}

func (a *VkApi) ChangeKeyboard(user_id int, keyboard string) {
	ok := a.ChangeKeyboardWithoutDelete(user_id, keyboard)

	if ok == false {
		a.ChangeKeyboardWithDelete(user_id, keyboard)

		zap.S().Warnw("unable to change keyboard without delete")
	}
}

func (a *VkApi) ChangeKeyboardWithDelete(user_id int, keyboard string) {
	id := a.SendMessage(user_id, "Меняю клавиатуру", keyboard)
	a.DeleteMessage(user_id, id, 1)
}

func (a *VkApi) ChangeKeyboardWithoutDelete(user_id int, keyboard string) bool {
	message := a.GetLastBotMessage(user_id)
	if message == nil {
		return false
	}

	// vk allows edit any group's message somehow
	a.EditMessage(user_id, message.ID, message.Text, keyboard)
	return true
}
