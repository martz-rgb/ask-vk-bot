package form

import (
	"ask-bot/src/vk"

	"go.uber.org/zap"
)

type Layer struct {
	name string

	fields []*Field
	index  int

	values map[string]interface{}
}

func NewLayer(name string, fields []*Field) *Layer {
	return &Layer{
		name:   name,
		fields: fields,
		values: make(map[string]interface{}),
	}
}

func (l *Layer) Name() string {
	return l.name
}

func (l *Layer) Current() *Field {
	return l.fields[l.index]
}

func (l *Layer) SetFromMessage(message *vk.Message) (*vk.MessageParams, error) {
	field := l.Current()

	field.SetFromMessage(message)
	info, err := field.Validate()
	if info != nil || err != nil {
		return info, err
	}

	l.AddValue(field.Name(), field.Value())

	return nil, nil
}

func (l *Layer) SetFromOption(id string) (*vk.MessageParams, error) {
	field := l.Current()

	field.SetFromOption(id)
	info, err := field.Validate()
	if info != nil || err != nil {
		return info, err
	}

	l.AddValue(field.Name(), field.Value())

	return nil, nil
}

func (l *Layer) Next() (string, []*Field) {
	name, fields := l.Current().Next()

	l.index++

	if fields != nil {
		return name, fields
	}
	return "", nil
}

func (l *Layer) IsEnd() bool {
	return l.index >= len(l.fields)
}

func (l *Layer) AddValue(name string, value interface{}) {
	// replace if already exist
	_, ok := l.values[name]
	if ok {
		zap.S().Warnw("already exist field",
			"name", name,
			"layer form", l.values)
	}
	l.values[name] = value
}

func (l *Layer) Values() map[string]interface{} {
	return l.values
}
