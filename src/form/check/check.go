package check

import (
	"ask-bot/src/vk"
	"errors"
	"strings"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type Result string

func NewResult(err string) *Result {
	res := Result(err)
	return &res
}

func (res *Result) Ok() bool {
	if res == nil {
		return true
	}

	return false
}

func (res *Result) Error() string {
	if res == nil {
		return ""
	}

	return string(*res)
}

func (res *Result) ErrorToMessageParams() *vk.MessageParams {
	if res == nil {
		return nil
	}

	return &vk.MessageParams{
		Text: res.Error(),
	}
}

func NotEmpty(value interface{}) (*Result, error) {
	if value == nil {
		return NewResult("Поле обязательно для заполнения."), nil
	}

	return nil, nil
}

func NotEmptyPositiveInt(value interface{}) (*Result, error) {
	if value == nil {
		return NewResult("Поле обязательно для заполнения."), nil
	}

	message, ok := value.(int)
	if !ok {
		err := errors.New("failed to convert about value to int")
		return nil, zaperr.Wrap(err, "",
			zap.Any("value", value),
			zap.String("field", "AboutField"))
	}

	if message == 0 {
		return NewResult("Поле обязательно для заполнения."), nil
	}

	return nil, nil
}

func NotEmptyBool(value interface{}) (*Result, error) {
	if value == nil {
		return NewResult("Поле обязательно для заполнения."), nil
	}

	if _, ok := value.(bool); !ok {
		err := errors.New("failed to convert about value to bool")
		return nil, zaperr.Wrap(err, "",
			zap.Any("value", value),
			zap.String("field", "ConfirmReservationField"))
	}

	return nil, nil
}

func NotEmptyPhotoAttachment(value interface{}) (*Result, error) {
	if value == nil {
		return NewResult("Поле обязательно для заполнения."), nil
	}

	attachments, ok := value.(string)
	if !ok {
		err := errors.New("failed to convert about value to bool")
		return nil, zaperr.Wrap(err, "",
			zap.Any("value", value),
			zap.String("field", "ConfirmReservationField"))
	}

	items := strings.Split(attachments, ",")

	is_photo := false
	for _, item := range items {
		if strings.HasPrefix(item, "photo") {
			is_photo = true
			break
		}
	}

	if !is_photo {
		return NewResult("Сообщение должно содержать изображения."), nil
	}

	return nil, nil
}
