package dict

import (
	"errors"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

type Dictionary map[string]interface{}

func ExtractValue[T any](dict Dictionary, name string) (T, error) {
	if dict == nil {
		err := errors.New("dictionary is nil")
		return *new(T), zaperr.Wrap(err, "")
	}
	value, ok := dict[name]
	if !ok {
		err := errors.New("no such key is presented in form")
		return *new(T), zaperr.Wrap(err, "",
			zap.Any("key", name),
			zap.Any("form", dict))
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
