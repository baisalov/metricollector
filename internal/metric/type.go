package metric

import (
	"errors"
	"strings"
)

var (
	ErrIncorrectType = errors.New("incorrect metric type")
)

type Type string

const (
	Counter Type = "counter"
	Gauge   Type = "gauge"
)

func ParseType(s string) Type {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	return Type(s)
}

//goland:noinspection GoMixedReceiverTypes
func (t Type) IsValid() bool {
	switch t {
	case Counter, Gauge:
		return true
	default:
		return false
	}
}

//goland:noinspection GoMixedReceiverTypes
func (t Type) String() string {
	return string(t)
}

//goland:noinspection GoMixedReceiverTypes
func (t *Type) UnmarshalJSON(bytes []byte) error {
	v := strings.Trim(string(bytes), `"`)

	tt := ParseType(v)

	if !tt.IsValid() {
		return ErrIncorrectValue
	}

	*t = tt

	return nil
}
