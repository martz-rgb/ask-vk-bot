package vk

import (
	"errors"
	"maps"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type MessageParams struct {
	Id     int
	Text   string
	Params api.Params
}

type Message struct {
	ID          int
	Text        string
	Attachments []object.MessagesMessageAttachment
}

func (v *VK) MarkAsRead(user_id int) error {
	params := api.Params{
		"mark_conversation_as_read": true,
		"peer_id":                   user_id,
	}

	response, err := v.api.MessagesMarkAsRead(params)
	if err != nil {
		return zaperr.Wrap(err, "failed to mark as read vk messsage",
			zap.Any("params", params),
			zap.Int("response", response),
		)
	}

	zap.S().Debugw("successfully mark as read vk messsage",
		"params", params,
		"response", response)

	return nil
}

func (v *VK) SendMessage(user_id int, message string, keyboard string, args api.Params) (int, error) {
	params := api.Params{
		"user_id":   user_id,
		"random_id": v.r.Int(),
		"message":   message,
		"keyboard":  keyboard,
	}
	maps.Copy(params, args)

	response, err := v.api.MessagesSend(params)
	if err != nil {
		return -1, zaperr.Wrap(err, "failed to send vk messsage",
			zap.Any("params", params),
			zap.Int("response", response))
	}

	zap.S().Debugw("successfully sent vk messsage",
		"params", params,
		"response", response)

	return response, nil
}

func (v *VK) SendMessageParams(user_id int, message *MessageParams, keyboard string) (int, error) {
	params := api.Params{
		"user_id":   user_id,
		"random_id": v.r.Int(),
		"message":   message.Text,
		"keyboard":  keyboard,
	}
	maps.Copy(params, message.Params)

	response, err := v.api.MessagesSend(params)
	if err != nil {
		return -1, zaperr.Wrap(err, "failed to send vk messsage params",
			zap.Any("params", params),
			zap.Int("response", response))
	}

	zap.S().Debugw("successfully sent vk messsage params",
		"params", params,
		"response", response)

	return response, nil
}

func (v *VK) EditMessage(peer_id int, message_id int, message string, keyboard string, attachments string) error {
	params := api.Params{
		"peer_id":               peer_id,
		"message_id":            message_id,
		"message":               message,
		"keyboard":              keyboard,
		"attachment":            attachments,
		"keep_forward_messages": true,
	}

	response, err := v.api.MessagesEdit(params)
	if err != nil {
		return zaperr.Wrap(err, "failed to edit vk message",
			zap.Any("params", params),
			zap.Int("response", response))
	}

	zap.S().Debugw("successfully edited vk message",
		"params", params,
		"response", response)

	return nil
}

func (v *VK) DeleteMessage(peer_id int, message_id int, delete_for_all int) error {
	params := api.Params{
		"peer_id":        peer_id,
		"message_ids":    message_id,
		"delete_for_all": delete_for_all,
	}
	response, err := v.api.MessagesDelete(params)
	if err != nil {
		return zaperr.Wrap(err, "failed to delete vk message",
			zap.Any("params", params),
			zap.Any("response", response),
		)
	}

	zap.S().Debugw("successfully deleted vk message",
		"params", params,
		"response", response)

	return nil
}

func (v *VK) SendEventAnswer(event_id string, user_id int, peer_id int) error {
	params := api.Params{
		"event_id":  event_id,
		"user_id":   user_id,
		"random_id": v.r.Int(),
		"peer_id":   []int{peer_id},
	}

	response, err := v.api.MessagesSendMessageEventAnswer(params)
	if err != nil {
		return zaperr.Wrap(err, "failed to send vk event answer",
			zap.Any("params", params),
			zap.Int("reponse", response))
	}

	zap.S().Debugw("successfully sent vk event answer",
		"params", params,
		"response", response)

	return nil
}

func (v *VK) LastBotMessage(user_id int) (*object.MessagesMessage, error) {
	params := api.Params{
		"count":   20, // heuristic value
		"user_id": user_id,
	}

	response, err := v.api.MessagesGetHistory(params)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to get last bot message",
			zap.Any("params", params),
			zap.Any("response", response))
	}

	for _, message := range response.Items {
		if message.Out {
			return &message, nil
		}
	}

	err = errors.New("unable to find last bot message")

	return nil, zaperr.Wrap(err, "",
		zap.Any("params", params),
		zap.Any("response", response))
}

func (v *VK) ChangeKeyboard(user_id int, keyboard string) error {
	err := v.ChangeKeyboardWithoutDelete(user_id, keyboard)
	if err != nil {
		zap.S().Infow("unable to change keyboard without delete", "error", err)
		return v.ChangeKeyboardWithDelete(user_id, keyboard)
	}

	return nil
}

func (v *VK) ChangeKeyboardWithDelete(user_id int, keyboard string) error {
	id, err := v.SendMessage(user_id, "Меняю клавиатуру", keyboard, nil)
	if err != nil {
		return err
	}

	return v.DeleteMessage(user_id, id, 1)
}

func (v *VK) ChangeKeyboardWithoutDelete(user_id int, keyboard string) error {
	message, err := v.LastBotMessage(user_id)
	if err != nil {
		return err
	}

	// vk allows edit any group's message somehow
	attachments := message.Attachments

	return v.EditMessage(user_id, message.ID, message.Text, keyboard, ToAttachments(attachments))
}
