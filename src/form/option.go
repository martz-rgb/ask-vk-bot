package form

type Option struct {
	ID    string
	Label string
	Color string
	Value interface{}
}

func OptionToLabel(option Option) string {
	return option.Label
}

func OptionToColor(option Option) string {
	return option.Color
}

func OptionToValue(option Option) string {
	return option.ID
}
