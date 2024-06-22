package form

import (
	"ask-bot/src/datatypes/dict"
	"ask-bot/src/datatypes/form/check"
	"ask-bot/src/vk"
)

type Field struct {
	Name string

	BuildRequest   func(dict.Dictionary) (*Request, bool, error)
	ExtrudeMessage func(*vk.Message) interface{}
	Check          func(interface{}) (*check.Result, error)

	Value interface{}

	request *Request
}

func (f *Field) Entry(d dict.Dictionary) (bool, error) {
	request, skip, err := f.BuildRequest(d)
	if err != nil {
		return false, err
	}

	f.request = request
	return skip, nil
}

func (f *Field) Request() *Request {
	return f.request
}

func (f *Field) SetFromMessage(message *vk.Message) {
	if f.ExtrudeMessage == nil {
		f.Value = nil
		return
	}
	f.Value = f.ExtrudeMessage(message)
}

func (f *Field) SetFromOption(id string) {
	f.Value = nil

	for _, option := range f.request.Options {
		if option.ID == id {
			f.Value = option.Value
			return
		}
	}
}

func (f *Field) Validate() (*check.Result, error) {
	if f.Check == nil {
		return nil, nil
	}

	return f.Check(f.Value)
}
