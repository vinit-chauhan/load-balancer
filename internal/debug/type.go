package debug

import "fmt"

type Printable struct {
	value interface{}
}

func NewPrintable(val interface{}) Printable {
	return Printable{
		value: val,
	}
}

func (p Printable) String() string {
	switch p.value.(type) {
	case int:
		return fmt.Sprintf("%d", p.value)
	case float64:
		return fmt.Sprintf("%f", p.value)
	case bool:
		return fmt.Sprintf("%t", p.value)
	case string:
		return p.value.(string)
	default:
		return fmt.Sprintf("%v", p.value)
	}
}
