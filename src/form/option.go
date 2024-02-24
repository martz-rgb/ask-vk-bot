package form

type Option struct {
	ID    string
	Label string
	Value interface{}
}

func OptionToLabel(option Option) string {
	return option.Label
}

func OptionToValue(option Option) string {
	return option.ID
}
