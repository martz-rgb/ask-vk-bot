package templates

import (
	"ask-bot/src/templates/russian"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"reflect"
	"text/template"
	"time"

	"github.com/hori-ryota/zaperr"
	"github.com/rb-go/plural-ru"
	"go.uber.org/zap"
)

/*
Works as singletone incapsulated in package.
Should be initialized with NewFromFile before any use.
*/
var r *rand.Rand

func NewFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return zaperr.Wrap(err, "failed to open file",
			zap.String("filename", filename))
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return zaperr.Wrap(err, "failed to read content of file",
			zap.String("filename", filename))
	}

	dict := map[string][]string{}

	err = json.Unmarshal(content, &dict)
	if err != nil {
		return zaperr.Wrap(err, "failed to unmarshal json from file",
			zap.String("filename", filename))
	}

	for id, t := range Templates {
		texts, ok := dict[string(id)]
		if !ok {
			err := errors.New("no template in file")
			return zaperr.Wrap(err, "",
				zap.String("id", string(id)))
		}
		if len(texts) == 0 {
			err := errors.New("no template texts")
			return zaperr.Wrap(err, "",
				zap.String("id", string(id)))
		}

		templates := make([]*template.Template, len(texts))

		for i, text := range texts {
			temp, err := template.New(string(id)).
				Funcs(template.FuncMap{
					"add": func(num1 int, num2 int) int {
						return num1 + num2
					},
					"plural": plural.Noun[int],
					"abs": func(num int) int {
						if num < 0 {
							return -num
						}
						return num
					},
					"rudate": func(t time.Time) string {
						return fmt.Sprintf("%d %s %d", t.Day(), russian.MonthGenitive(t.Month()), t.Year())
					},
					"vkid": func(id int) string {
						return fmt.Sprintf("@id%d", id)
					},
				}).
				Parse(text)
			if err != nil {
				return zaperr.Wrap(err, "failed to parse template",
					zap.String("template", text))
			}
			templates[i] = temp
		}

		t.Templates = templates
		Templates[id] = t
	}

	r = rand.New(rand.NewSource(time.Now().Unix()))

	return nil
}

func ParseTemplate(id TemplateID, data interface{}) (string, error) {
	ts, ok := Templates[id]
	if !ok {
		err := errors.New("no template with such id")
		return "", zaperr.Wrap(err, "",
			zap.String("id", string(id)))
	}

	// check data
	target_type := reflect.TypeOf(ts.Type)
	actual_type := reflect.TypeOf(data)

	if actual_type.Kind() != reflect.Pointer {
		actual_type = reflect.PointerTo(actual_type)
	}

	if target_type != actual_type {
		err := errors.New("invalid type of data for template")
		return "", zaperr.Wrap(err, "",
			zap.String("id", string(id)),
			zap.Any("expected type", ts.Type),
			zap.Any("data", data))
	}

	// choose template
	t := ts.Templates[r.Intn(len(ts.Templates))]

	wr := bytes.NewBufferString("")
	err := t.Execute(wr, data)
	if err != nil {
		return "", zaperr.Wrap(err, "failed to execute template",
			zap.String("id", string(id)),
			zap.Any("template", t),
			zap.Any("data", data))
	}

	return wr.String(), nil
}
