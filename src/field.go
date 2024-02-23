package main

type Field struct {
	name string

	request *MessageParams
	options []Option

	extract  func(*Message) interface{}
	validate func(interface{}) (*MessageParams, error)
	next     func(interface{}) (string, []*Field)

	value interface{}
}

func NewField(name string,
	request *MessageParams,
	options []Option,
	extract func(*Message) interface{},
	validate func(interface{}) (*MessageParams, error),
	next func(interface{}) (string, []*Field)) *Field {

	return &Field{
		name:     name,
		request:  request,
		options:  options,
		extract:  extract,
		validate: validate,

		value: nil,
	}
}

func (f *Field) Name() string {
	return f.name
}

func (f *Field) Request() *MessageParams {
	return f.request
}

func (f *Field) Options() []Option {
	return f.options
}

func (f *Field) Next() (string, []*Field) {
	if f.next == nil {
		return "", nil
	}
	return f.next(f.value)
}

func (f *Field) Value() interface{} {
	return f.value
}

func (f *Field) SetFromMessage(message *Message) {
	if f.extract == nil {
		f.value = nil
		return
	}
	f.value = f.extract(message)
}

func (f *Field) SetFromOption(id string) {
	f.value = nil

	for _, option := range f.options {
		if option.ID == id {
			f.value = option.Value
			return
		}
	}
}

func (f *Field) Validate() (*MessageParams, error) {
	if f.validate == nil {
		// no validation is required
		return nil, nil
	}
	return f.validate(f.value)
}
