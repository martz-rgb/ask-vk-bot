package paginator

import (
	"errors"

	"github.com/hori-ryota/zaperr"
)

var DefaultRows int = 2
var DefaultCols int = 3

// command string, label, value func(T) string,
type Config[T interface{}] struct {
	Command string // required

	Rows int
	Cols int

	WithoutBack bool

	ToLabel func(T) string // required
	ToColor func(T) string
	ToValue func(T) string // required

	is_build bool
}

func (c *Config[T]) Build() (*Config[T], error) {
	if c == nil {
		err := errors.New("no config is presented")
		return nil, zaperr.Wrap(err, "")
	}

	if c.Command == "" {
		err := errors.New("command is not presented")
		return nil, zaperr.Wrap(err, "")
	}
	if c.ToLabel == nil {
		err := errors.New("to label func is not presented")
		return nil, zaperr.Wrap(err, "")
	}
	if c.ToValue == nil {
		err := errors.New("to value func is not presented")
		return nil, zaperr.Wrap(err, "")
	}

	if c.Rows == 0 {
		c.Rows = DefaultRows
	}
	if c.Cols == 0 {
		c.Cols = DefaultCols
	}

	// if WithoutBack is not presented -- it is false, a default value
	// if ToColor is not presented -- it is nil, a default value

	c.is_build = true

	return c, nil
}

// build config and panics if non-nil error
func (c *Config[T]) MustBuild() *Config[T] {
	c, err := c.Build()
	if err != nil {
		panic(err)
	}

	return c
}
