package main

import (
	"errors"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type Form struct {
	fields []FormField
	index  int
}

func NewForm(fields []FormField) *Form {
	return &Form{
		fields: fields,
		index:  0,
	}
}

func (f *Form) Request() (string, error) {
	if f.index < len(f.fields) {
		return f.fields[f.index].Request(), nil
	}
	err := errors.New("out of fields range")
	return "", zaperr.Wrap(err, "",
		zap.Int("index", f.index),
		zap.Any("fields", f.fields))
}

func (f *Form) SetAndValidate(value *Message) (bool, string, error) {
	f.fields[f.index].SetValue(value)
	return f.fields[f.index].Validate()
}

func (f *Form) Next() (end bool) {
	f.index++
	if f.index >= len(f.fields) {
		f.index = len(f.fields) - 1
		return true
	}
	return false
}

func (f *Form) Previous() {
	f.index--
	if f.index < 0 {
		f.index = 0
	}
}

func (f *Form) Buttons() [][]Button {
	buttons := [][]Button{
		{
			{
				Label:   "Назад",
				Color:   NegativeColor,
				Command: "back",
			},
		},
	}

	if f.index > 0 {
		prev_button := []Button{
			{
				Label:   "К предыдущему",
				Color:   PrimaryColor,
				Command: "previous",
			},
		}
		buttons[0] = append(prev_button, buttons[0]...)
	}

	if !f.fields[f.index].Mandatory() {
		skip_button := [][]Button{
			{
				{
					Label:   "Оставить пустым",
					Color:   SecondaryColor,
					Command: "skip",
				},
			},
		}

		buttons = append(skip_button, buttons...)
	}

	return buttons
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
