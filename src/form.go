package main

import (
	"errors"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type Form struct {
	fields []FormField
	index  int

	paginator *Paginator[Option]
}

func NewForm(fields ...FormField) (*Form, error) {
	if len(fields) == 0 {
		err := errors.New("form must not be empty")
		return nil, zaperr.Wrap(err, "")
	}

	form := &Form{
		fields: fields,
		index:  0,
	}
	form.updatePaginator()

	return form, nil
}

func (f *Form) Request() *RequestMessage {
	return f.fields[f.index].Request()
}

func (f *Form) SetFromMessageAndValidate(m *Message) (bool, string, error) {
	f.fields[f.index].SetFromMessage(m)
	return f.fields[f.index].Validate()
}

func (f *Form) SetOptionAndValidate(id string) (bool, string, error) {
	f.fields[f.index].SetOption(id)
	return f.fields[f.index].Validate()
}

func (f *Form) Next() (end bool) {
	f.index++
	if f.index >= len(f.fields) {
		f.index = len(f.fields) - 1
		return true
	}

	f.updatePaginator()
	return false
}

func (f *Form) Up() {
	f.index--
	if f.index < 0 {
		f.index = 0
	}
}

func (f *Form) Control(command string) bool {
	return f.paginator.Control(command)
}

func (f *Form) updatePaginator() {
	options := f.fields[f.index].Options()

	f.paginator = NewPaginator[Option](options, "form", RowsCount, ColsCount, OptionToLabel, OptionToValue)
}

func (f *Form) Buttons() [][]Button {
	special := []Button{}

	if f.index > 0 {
		special = append(special, Button{
			Label:   "^",
			Color:   PrimaryColor,
			Command: "up",
		})
	}

	return f.paginator.Buttons(special...)
}

func (f *Form) Value(index int) (interface{}, error) {
	if index < 0 || index >= len(f.fields) {
		err := errors.New("out of range")
		return nil, zaperr.Wrap(err, "",
			zap.Int("index", index),
			zap.Any("form", f.fields))
	}
	return f.fields[index].Value(), nil
}

func ConvertValue[T any](value interface{}) (T, error) {
	typed, ok := value.(T)
	if !ok {
		err := errors.New("failed to convert to string")
		return *new(T), zaperr.Wrap(err, "",
			zap.Any("type", *new(T)),
			zap.Any("value", value))
	}

	return typed, nil
}
