package main

import (
	"errors"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/awnumar/memguard"
	"go.uber.org/zap"
)

const (
	PrimaryColor   = "primary"
	SecondaryColor = "secondary"
	PositiveColor  = "positive"
	NegativeColor  = "negative"
)

// api.VK is based on http.Client and http.Client is claimed to be concurrency safe
type VK struct {
	id int

	api *api.VK
	r   *rand.Rand
}

func NewVKFromFile(name string, id int) (*VK, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	token, err := memguard.NewBufferFromEntireReader(file)
	if err != nil {
		return nil, err
	}
	defer token.Destroy()

	if token.Size() == 0 {
		return nil, errors.New("group token is not provided")
	}

	api, err := NewVK(id, token)
	if err != nil {
		return nil, err
	}

	return api, nil
}

func NewVK(id int, token *memguard.LockedBuffer) (*VK, error) {
	v := &VK{}
	source := rand.NewSource(time.Now().UnixNano())
	v.r = rand.New(source)
	v.id = id

	// should copy string because VK saves it inside and use,
	// but i destroy LockedBuffers with pointer on string
	v.api = api.NewVK(strings.Clone(token.String()))

	return v, nil
}

func (v *VK) MarkAsRead(user_id int) {
	params := api.Params{
		"mark_conversation_as_read": true,
		"peer_id":                   user_id,
	}

	response, err := v.api.MessagesMarkAsRead(params)
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

func (v *VK) SendMessage(user_id int, message string, keyboard string, attachment string) int {
	params := api.Params{
		"user_id":    user_id,
		"random_id":  v.r.Int(),
		"message":    message,
		"keyboard":   keyboard,
		"attachment": attachment,
	}

	response, err := v.api.MessagesSend(params)
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

func (v *VK) EditMessage(peer_id int, message_id int, message string, keyboard string) {
	params := api.Params{
		"peer_id":    peer_id,
		"message_id": message_id,
		"message":    message,
		"keyboard":   keyboard,
	}

	response, err := v.api.MessagesEdit(params)
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

func (v *VK) DeleteMessage(peer_id int, message_id int, delete_for_all int) {
	params := api.Params{
		"peer_id":        peer_id,
		"message_ids":    message_id,
		"delete_for_all": delete_for_all,
	}
	response, err := v.api.MessagesDelete(params)
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

func (v *VK) SendEventAnswer(event_id string, user_id int, peer_id int) {
	params := api.Params{
		"event_id":  event_id,
		"user_id":   user_id,
		"random_id": v.r.Int(),
		"peer_id":   []int{peer_id},
	}

	response, err := v.api.MessagesSendMessageEventAnswer(params)
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

func (v *VK) GetLastBotMessage(user_id int) *object.MessagesMessage {
	params := api.Params{
		"count":   20, // heuristic value
		"user_id": user_id,
	}

	response, err := v.api.MessagesGetHistory(params)
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

func (v *VK) ChangeKeyboard(user_id int, keyboard string) {
	ok := v.ChangeKeyboardWithoutDelete(user_id, keyboard)

	if ok == false {
		v.ChangeKeyboardWithDelete(user_id, keyboard)

		zap.S().Warnw("unable to change keyboard without delete")
	}
}

func (v *VK) ChangeKeyboardWithDelete(user_id int, keyboard string) {
	id := v.SendMessage(user_id, "Меняю клавиатуру", keyboard, "")
	v.DeleteMessage(user_id, id, 1)
}

func (v *VK) ChangeKeyboardWithoutDelete(user_id int, keyboard string) bool {
	message := v.GetLastBotMessage(user_id)
	if message == nil {
		return false
	}

	// vk allows edit any group's message somehow
	v.EditMessage(user_id, message.ID, message.Text, keyboard)
	return true
}

func (v *VK) WallPostNew(group_id int, message string, attachments string, signed bool, publish_date int64) {
	params := api.Params{
		"owner_id":     -group_id,
		"from_group":   1,
		"message":      message,
		"attachments":  attachments,
		"signed":       signed,
		"publish_date": publish_date,
	}

	response, err := v.api.WallPost(params)
	if err != nil {
		zap.S().Errorw("failed to post on wall",
			"error", err,
			"params", params,
			"response", response)
		return
	}

	zap.S().Debugw("successfully posted on wall",
		"params", params,
		"response", response)
}

func (v *VK) UploadDocument(peer_id int, name string, file io.Reader) int {
	response, err := v.api.UploadMessagesDoc(peer_id, "doc", name, "", file)
	if err != nil {
		zap.S().Errorw("failed to upload document",
			"error", err,
			"peer_id", peer_id,
			"response", response)
		return 0
	}

	zap.S().Debugw("successfully upload document",
		"peer_id", peer_id,
		"response", response)

	return response.Doc.ID
}
