package zram

import "fyne.io/fyne/v2/data/binding"

type Option struct {
	Name   string
	State  string
	Values []string
}

func NewOption(name string, state string, values []string) Option {
	return Option{name, state, values}
}

func NewOptionFromDataItem(item binding.DataItem) Option {
	v, _ := item.(binding.Untyped).Get()
	return v.(Option)
}
