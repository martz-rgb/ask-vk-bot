package main

import (
	"errors"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

// TO-DO: implement undo action
type Form struct {
	layers *Stack[*Layer]
	form   map[string]interface{}

	paginator *Paginator[Option]
}

func NewForm(fields ...*Field) *Form {
	f := &Form{
		layers: &Stack[*Layer]{NewLayer("", fields)},
	}

	f.update()

	return f
}

func (f *Form) Request() *MessageParams {
	return f.layers.Peek().Current().Request()
}

func (f *Form) SetFromMessage(m *Message) (*MessageParams, error) {
	return f.layers.Peek().SetFromMessage(m)
}

func (f *Form) SetFromOption(id string) (*MessageParams, error) {
	return f.layers.Peek().SetFromOption(id)
}

func (f *Form) Next() (end bool) {
	name, fields := f.layers.Peek().Next()
	if fields != nil {
		f.layers.Push(NewLayer(name, fields))
		f.update()
		return
	}

	for f.layers.Len() > 0 && f.layers.Peek().IsEnd() {
		layer := f.layers.Pop()

		if f.layers.Len() > 0 {
			f.layers.Peek().AddValue(layer.Name(), layer.Form())
		} else {
			f.form = layer.Form()
			return true
		}
	}

	f.update()
	return false
}

func (f *Form) Buttons() [][]Button {
	return f.paginator.Buttons()
}

func (f *Form) Control(command string) (back bool) {
	return f.paginator.Control(command)
}

func (f *Form) Values() map[string]interface{} {
	return f.form
}

func (f *Form) update() {
	if f.layers.Len() == 0 {
		return
	}

	if f.paginator == nil {
		f.paginator = NewPaginator[Option](
			f.layers.Peek().Current().Options(),
			"form",
			RowsCount,
			ColsCount,
			false,
			OptionToLabel,
			OptionToValue,
		)
		return
	}

	f.paginator.ChangeObjects(f.layers.Peek().Current().Options())
}

func ExtractValue[T any](form map[string]interface{}, name string) (T, error) {
	if form == nil {
		err := errors.New("form is nil")
		return *new(T), zaperr.Wrap(err, "")
	}
	value, ok := form[name]
	if !ok {
		err := errors.New("no such key is presented in form")
		return *new(T), zaperr.Wrap(err, "",
			zap.Any("key", name),
			zap.Any("form", form))
	}

	typed, ok := value.(T)
	if !ok {
		err := errors.New("failed to convert to required type")
		return *new(T), zaperr.Wrap(err, "",
			zap.Any("type", *new(T)),
			zap.Any("value", value))
	}

	return typed, nil
}
