package templates

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"os"
	"reflect"
	"text/template"
	"time"

	"github.com/hori-ryota/zaperr"
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
			temp, err := template.New(string(id)).Parse(text)
			if err != nil {
				return zaperr.Wrap(err, "failed to parse template",
					zap.String("template", text))
			}
			templates[i] = temp
		}

		t.Templates = templates
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
	if reflect.TypeOf(data) != ts.Type {
		err := errors.New("invalid type of data for template")
		return "", zaperr.Wrap(err, "",
			zap.Any("expected type", ts.Type),
			zap.Any("data", data))
	}

	// choose template
	t := ts.Templates[r.Intn(len(ts.Templates))]

	wr := bytes.NewBufferString("")
	err := t.Execute(wr, data)
	if err != nil {
		return "", zaperr.Wrap(err, "failed to execute template",
			zap.Any("template", t),
			zap.Any("data", data))
	}

	return wr.String(), nil
}
