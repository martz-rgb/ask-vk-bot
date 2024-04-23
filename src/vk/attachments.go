package vk

import (
	"fmt"
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

// i don't know why it returns array
func (v *VK) UploadPhotoToWall(file io.Reader) ([]object.PhotosPhoto, error) {
	// group_id here should be grateer than 0
	id := v.id
	if id < 0 {
		id = -id
	}

	response, err := v.api.UploadGroupWallPhoto(id, file)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to upload photo to group wall",
			zap.Int("peer_id", v.id),
			zap.Any("response", response))
	}

	return response, nil
}

// TO-DO maybe they all should have access key?
func ToAttachments(attachments []object.MessagesMessageAttachment) string {
	result := []string{}

	for _, a := range attachments {
		switch a.Type {
		case object.AttachmentTypePhoto:
			attachment := fmt.Sprintf("photo%d_%d_%s", a.Photo.OwnerID, a.Photo.ID, a.Photo.AccessKey)
			result = append(result, attachment)

			//result = append(result, a.Photo.ToAttachment())
		case object.AttachmentTypeVideo:
			result = append(result, a.Video.ToAttachment())
		case object.AttachmentTypeAudio:
			result = append(result, a.Audio.ToAttachment())
		case object.AttachmentTypeDoc:
			result = append(result, a.Doc.ToAttachment())
		case object.AttachmentTypeLink:
			//result = append(result, a.Link.ToAttachment())
		}
	}

	return strings.Join(result, ",")
}
