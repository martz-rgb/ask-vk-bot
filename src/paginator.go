package main

import (
	"errors"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type Paginator[T interface{}] struct {
	objects     []T
	page        int
	total_pages int
	command     string

	rows int
	cols int

	ToLabel func(T) string
	ToValue func(T) string
}

var RowsCount int = 2
var ColsCount int = 3

func NewPaginator[T interface{}](objects []T, command string, rows, cols int, label, value func(T) string) *Paginator[T] {
	return &Paginator[T]{
		objects: objects,
		page:    0,
		// ceil function
		total_pages: 1 + (len(objects)-1)/(rows*cols),
		command:     command,
		rows:        rows,
		cols:        cols,
		ToLabel:     label,
		ToValue:     value,
	}
}

func (p *Paginator[T]) Next() {
	p.page += 1
	if p.page >= p.total_pages {
		p.page = p.total_pages - 1
	}
}

func (p *Paginator[T]) Previous() {
	p.page -= 1
	if p.page < 0 {
		p.page = 0
	}
}

func (p *Paginator[T]) ChangeObjects(objects []T) {
	p.objects = objects
	p.page = 0
	p.total_pages = 1 + (len(objects)-1)/(p.rows*p.cols)
}

func (p *Paginator[T]) Object(value string) (*T, error) {
	for _, object := range p.objects {
		if p.ToValue(object) == value {
			return &object, nil
		}
	}

	err := errors.New("failed to find value in paginator")
	return nil, zaperr.Wrap(err, "",
		zap.String("value", value),
		zap.Any("objects", p.objects))
}

func (p *Paginator[T]) Buttons(special ...Button) [][]Button {
	buttons := [][]Button{}

	for i := 0; i < p.rows; i++ {
		if i*p.cols >= len(p.objects) {
			break
		}

		buttons = append(buttons, []Button{})

		for j := 0; j < p.cols; j++ {
			index := i*p.cols + j + p.page*(p.rows*p.cols)

			if index >= len(p.objects) {
				i = p.rows
				break
			}

			buttons[i] = append(buttons[i], Button{
				Label: p.ToLabel(p.objects[index]),
				Color: "secondary",

				Command: p.command,
				Value:   p.ToValue(p.objects[index]),
			})
		}
	}

	// + доп ряд с функциональными кнопками
	controls := special

	if p.page > 0 {
		controls = append(controls, Button{
			Label: "<<",
			Color: "primary",

			Command: "paginator",
			Value:   "previous",
		})
	}

	if p.page < p.total_pages-1 {
		controls = append(controls, Button{
			Label: ">>",
			Color: "primary",

			Command: "paginator",
			Value:   "next",
		})
	}

	controls = append(controls, Button{
		Label: "Назад",
		Color: "negative",

		Command: "paginator",
		Value:   "back",
	})

	buttons = append(buttons, controls)

	return buttons
}

func (p *Paginator[T]) Control(command string) bool {
	switch command {
	case "next":
		p.Next()
	case "previous":
		p.Previous()
	case "back":
		return true
	}

	return false
}
