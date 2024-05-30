package form

import (
	"ask-bot/src/dict"
	"ask-bot/src/form/check"
	"ask-bot/src/paginator"
	"ask-bot/src/vk"
	"errors"
)

// TO-DO: implement undo action
type Form struct {
	fields []Field
	index  int

	paginator *paginator.Paginator[Option]
}

func (form *Form) current() *Field {
	return &form.fields[form.index]
}

func NewForm(fields ...Field) (*Form, error) {

	form := &Form{
		fields: fields,
		index:  -1,
	}

	end, err := form.Next()
	if err != nil {
		return nil, err
	}

	if end {
		err = errors.New("form is empty")
		return nil, err
	}

	return form, nil
}

func (form *Form) Request() *Request {
	return form.current().Request()
}

func (form *Form) SetFromMessage(m *vk.Message) (*check.Result, error) {
	form.current().SetFromMessage(m)
	return form.current().Validate()
}

func (form *Form) SetFromOption(id string) (*check.Result, error) {
	form.current().SetFromOption(id)
	return form.current().Validate()
}

func (form *Form) Next() (end bool, err error) {
	form.index++
	if form.index >= len(form.fields) {
		return true, nil
	}

	skip, err := form.current().Entry(form.Values())
	if err != nil {
		return false, err
	}

	for skip && form.index < len(form.fields)-1 {
		form.index++
		skip, err = form.current().Entry(form.Values())
		if err != nil {
			return false, err
		}
	}

	form.update()

	return skip, nil
}

func (form *Form) Buttons() [][]vk.Button {
	return form.paginator.Buttons()
}

func (form *Form) Control(command string) (back bool) {
	return form.paginator.Control(command)
}

func (form *Form) Values() dict.Dictionary {
	d := dict.Dictionary{}

	for i := 0; i < form.index; i++ {
		field := form.fields[i]
		d[field.Name] = field.Value
	}

	return d
}

func (form *Form) update() {
	if form.paginator == nil {
		config := &paginator.Config[Option]{
			Command: "form",
			ToLabel: OptionToLabel,
			ToColor: OptionToColor,
			ToValue: OptionToValue,
		}

		form.paginator = paginator.New[Option](
			form.Request().Options,
			config.MustBuild())
		return
	}

	form.paginator.ChangeObjects(form.Request().Options)
}
