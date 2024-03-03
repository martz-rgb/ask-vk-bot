package paginator

import (
	"ask-bot/src/vk"
	"errors"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

var DeafultRows int = 2
var DefaultCols int = 3

type Paginator[T interface{}] struct {
	objects     []T
	page        int
	total_pages int
	command     string

	rows         int
	cols         int
	without_back bool

	ToLabel func(T) string
	ToColor func(T) string
	ToValue func(T) string
}

func New[T interface{}](objects []T, command string, rows, cols int, without_back bool, label, color, value func(T) string) *Paginator[T] {
	return &Paginator[T]{
		objects: objects,
		page:    0,
		// ceil function
		total_pages:  1 + (len(objects)-1)/(rows*cols),
		command:      command,
		rows:         rows,
		cols:         cols,
		without_back: without_back,
		ToLabel:      label,
		ToColor:      color,
		ToValue:      value,
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

func (p *Paginator[T]) Buttons(special ...vk.Button) [][]vk.Button {
	buttons := [][]vk.Button{}

	for i := 0; i < p.rows; i++ {
		if p.page*(p.rows*p.cols)+i*p.cols >= len(p.objects) {
			break
		}

		buttons = append(buttons, []vk.Button{})

		for j := 0; j < p.cols; j++ {
			index := i*p.cols + j + p.page*(p.rows*p.cols)

			if index >= len(p.objects) {
				i = p.rows
				break
			}

			var color string
			if p.ToColor == nil {
				color = vk.SecondaryColor
			} else {
				color = p.ToColor(p.objects[index])
				if color == vk.NoneColor {
					color = vk.SecondaryColor
				}
			}

			buttons[i] = append(buttons[i], vk.Button{
				Label: p.ToLabel(p.objects[index]),
				Color: color,

				Command: p.command,
				Value:   p.ToValue(p.objects[index]),
			})
		}
	}

	// + доп ряд с функциональными кнопками
	controls := special

	if p.page > 0 {
		controls = append(controls, vk.Button{
			Label: "<<",
			Color: "primary",

			Command: "paginator",
			Value:   "previous",
		})
	}

	if p.page < p.total_pages-1 {
		controls = append(controls, vk.Button{
			Label: ">>",
			Color: "primary",

			Command: "paginator",
			Value:   "next",
		})
	}

	if !p.without_back {
		controls = append(controls, vk.Button{
			Label: "Назад",
			Color: "negative",

			Command: "paginator",
			Value:   "back",
		})
	}

	if len(controls) > 0 {
		buttons = append(buttons, controls)
	}

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
