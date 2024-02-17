package main

import (
	"encoding/json"
	"errors"
	"io"
	"maps"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/awnumar/memguard"
	"github.com/hori-ryota/zaperr"
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

type VKForwardMessage struct {
	PeerId      int   `json:"peer_id"`
	MessagesIds []int `json:"message_ids"`
}

func ForwardParam(vk_id int, messages []int) (string, error) {
	forward, err := json.Marshal(VKForwardMessage{
		vk_id,
		messages,
	})
	if err != nil {
		return "", zaperr.Wrap(err, "failed to marshal forward message param",
			zap.Int("vk_id", vk_id),
			zap.Any("messages", messages))
	}

	return string(forward), nil
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

func (v *VK) GetLastBotMessage(user_id int) (*object.MessagesMessage, error) {
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
		zap.S().Debugw("unable to change keyboard without delete")
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
	message, err := v.GetLastBotMessage(user_id)
	if err != nil {
		return err
	}

	// vk allows edit any group's message somehow
	attachments := message.Attachments

	return v.EditMessage(user_id, message.ID, message.Text, keyboard, ToAttachments(attachments))
}

func (v *VK) WallPostNew(group_id int, message string, attachments string, signed bool, publish_date int64) error {
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
		return zaperr.Wrap(err, "failed to post on wall",
			zap.Any("params", params),
			zap.Any("response", response))
	}

	zap.S().Debugw("successfully posted on wall",
		"params", params,
		"response", response)

	return nil
}

func (v *VK) UploadDocument(peer_id int, name string, file io.Reader) (int, error) {
	response, err := v.api.UploadMessagesDoc(peer_id, "doc", name, "", file)
	if err != nil {
		return -1, zaperr.Wrap(err, "failed to upload document",
			zap.Int("peer_id", peer_id),
			zap.Any("response", response))
	}

	zap.S().Debugw("successfully upload document",
		"peer_id", peer_id,
		"response", response)

	return response.Doc.ID, nil
}

func ToAttachments(attachments []object.MessagesMessageAttachment) string {
	result := []string{}

	for _, a := range attachments {
		switch a.Type {
		case "photo":
			result = append(result, a.Photo.ToAttachment())
		case "video":
			result = append(result, a.Video.ToAttachment())
		case "audio":
			result = append(result, a.Audio.ToAttachment())
		case "doc":
			result = append(result, a.Doc.ToAttachment())
		case "link":
			//result = append(result, a.Link.ToAttachment())
		case "market":
			result = append(result, a.Market.ToAttachment())
		case "market_album":
			result = append(result, a.MarketMarketAlbum.ToAttachment())
		case "wall":
			//result = append(resul)
		case "wall_reply":
			//result = append(result, a.WallReply.ToAttachment())
		case "sticker":
			//result =append(result, a.Sticker.ToAttachment())
		case "gift":
			//result = append(result, a.Gift.ToAttachment())
		}
	}

	return strings.Join(result, ",")
}
