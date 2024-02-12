package main

import (
	"errors"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type FormField interface {
	Request() string

	SetValue(*Message)
	Value() interface{}
	// if not ok, return the cause in second variable
	// last error is for techinical problems
	Validate() (bool, string, error)

	Mandatory() bool
	IsEmpty() bool
}

// Information about user for confirm reservations
type AboutField struct {
	request string

	value interface{}
}

func NewAboutField(request string) *AboutField {
	return &AboutField{
		request: request,
	}
}

func (f *AboutField) Request() string {
	return f.request
}

func (f *AboutField) SetValue(value *Message) {
	if value == nil {
		f.value = nil
		return
	}

	f.value = value.ID
}

func (f *AboutField) Value() interface{} {
	return f.value
}
func (f *AboutField) Validate() (bool, string, error) {
	id, ok := f.value.(int)
	if !ok {
		err := errors.New("failed to convert about value to int")
		return false, "", zaperr.Wrap(err, "",
			zap.String("field", "AboutField"))
	}

	if id == 0 {
		message := "поле обязательно для заполнения"
		return false, message, nil
	}

	return true, "", nil
}

func (f *AboutField) Mandatory() bool {
	return true
}
func (f *AboutField) IsEmpty() bool {
	return false
}
