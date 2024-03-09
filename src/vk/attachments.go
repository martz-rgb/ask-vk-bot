package vk

import (
	"io"
	"strings"

	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

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
