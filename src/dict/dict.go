package dict

import (
	"errors"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
)

// TO-DO: extract from map to struct

type Dictionary map[string]interface{}

func ExtractValue[T any](dict Dictionary, keys ...string) (T, error) {
	var err error

	if dict == nil {
		err = errors.New("dictionary is nil")
		return *new(T), zaperr.Wrap(err, "")
	}

	for i := 0; i < len(keys)-1; i++ {
		dict, err = ExtractFlatten[Dictionary](dict, keys[i])
		if err != nil {
			return *new(T), err
		}
	}

	return ExtractFlatten[T](dict, keys[len(keys)-1])
}

func ExtractFlatten[T any](dict Dictionary, key string) (T, error) {
	if dict == nil {
		err := errors.New("dictionary is nil")
		return *new(T), zaperr.Wrap(err, "")
	}

	value, ok := dict[key]
	if !ok {
		err := errors.New("no such key is presented in form")
		return *new(T), zaperr.Wrap(err, "",
			zap.Any("key", key),
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
