package templates

import (
	"reflect"
	"text/template"
)

type Template struct {
	Templates []*template.Template
	Type      reflect.Type
}

var Templates = map[TemplateID]*Template{
	MessageGreeting: {
		Type: reflect.TypeOf(MessageGreetingData{}),
	},
}
