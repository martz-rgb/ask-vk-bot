package main

import (
	"errors"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type Option struct {
	ID    string
	Label string
	Value interface{}
}

func OptionToLabel(option Option) string {
	return option.Label
}

func OptionToValue(option Option) string {
	return option.ID
}

type FormField interface {
	Request() *MessageParams

	Options() []Option

	SetFromMessage(*Message)
	SetOption(id string) // option id
	Value() interface{}
	// if not ok, return the cause in second variable
	// last error is for techinical problems
	Validate() (bool, string, error)
}

// about field
// Information about user for confirm reservations
type AboutField struct {
	request *MessageParams

	value interface{}
}

func NewAboutField(request *MessageParams) *AboutField {
	return &AboutField{
		request: request,
		value:   nil,
	}
}

func (f *AboutField) Request() *MessageParams {
	return f.request
}

func (f *AboutField) Options() []Option {
	return nil
}

func (f *AboutField) SetFromMessage(value *Message) {
	if value == nil {
		f.value = nil
		return
	}

	f.value = value.ID
}

func (f *AboutField) SetOption(id string) {}

func (f *AboutField) Value() interface{} {
	return f.value
}
func (f *AboutField) Validate() (bool, string, error) {
	id, ok := f.value.(int)
	if !ok {
		err := errors.New("failed to convert about value to int")
		return false, "", zaperr.Wrap(err, "",
			zap.Any("value", f.value),
			zap.String("field", "AboutField"))
	}

	if id == 0 {
		message := "поле обязательно для заполнения"
		return false, message, nil
	}

	return true, "", nil
}

// confirm reservation field
type ConfirmReservationField struct {
	request *MessageParams
	value   interface{}
}

func NewConfirmReservation(request *MessageParams) *ConfirmReservationField {
	return &ConfirmReservationField{
		request: request,
		value:   nil,
	}
}

func (f *ConfirmReservationField) Request() *MessageParams {
	return f.request
}

func (f *ConfirmReservationField) Options() []Option {
	return []Option{
		{
			ID:    "confirm",
			Label: "Потвердить",
		},
		{
			ID:    "delete",
			Label: "Удалить",
		},
	}
}

func (f *ConfirmReservationField) SetFromMessage(*Message) {}

func (f *ConfirmReservationField) SetOption(id string) {
	switch id {
	case "confirm":
		f.value = true
	case "delete":
		f.value = false
	default:
		f.value = nil
	}
}

func (f *ConfirmReservationField) Value() interface{} {
	return f.value
}

// if not ok, return the cause in second variable
// last error is for techinical problems
func (f *ConfirmReservationField) Validate() (bool, string, error) {
	if f.value == nil {
		message := "поле обязательно для заполнения"
		return false, message, nil
	}

	if _, ok := f.value.(bool); !ok {
		err := errors.New("failed to convert about value to bool")
		return false, "", zaperr.Wrap(err, "",
			zap.Any("value", f.value),
			zap.String("field", "ConfirmReservationField"))
	}

	return true, "", nil
}
