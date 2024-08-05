package metric

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseType(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want Type
	}{
		{
			name: "gauge 1",
			arg:  "gauge",
			want: Gauge,
		},
		{
			name: "gauge 2",
			arg:  "Gauge ",
			want: Gauge,
		},
		{
			name: "counter 1",
			arg:  "counter",
			want: Counter,
		},
		{
			name: "counter 2",
			arg:  "Counter ",
			want: Counter,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, ParseType(tt.arg), "ParseType(%v)", tt.arg)
		})
	}
}

func TestType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		t    Type
		want bool
	}{
		{
			"counter",
			Counter,
			true,
		},
		{
			"gauge",
			Gauge,
			true,
		},
		{
			"incorrect",
			Type("incorrect"),
			false,
		},
		{
			"empty",
			Type(""),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.t.IsValid(), "IsValid()")
		})
	}
}

func TestType_String(t *testing.T) {
	tests := []struct {
		name string
		t    Type
		want string
	}{
		{"counter",
			Counter,
			"counter",
		},
		{"gauge",
			Gauge,
			"gauge",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.t.String(), "String()")
		})
	}
}

func TestType_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		t       Type
		bytes   []byte
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"counter 1",
			Counter,
			[]byte(`"counter"`),
			func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err, i)
			},
		},
		{
			"gauge 1",
			Gauge,
			[]byte(`"Gauge"`),
			func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err, i)
			},
		},

		{
			"counter 2",
			Counter,
			[]byte(`"Counter"`),
			func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err, i)
			},
		},
		{
			"gauge 2",
			Gauge,
			[]byte(`"Gauge"`),
			func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err, i)
			},
		},
		{
			"empty",
			Gauge,
			[]byte(`""`),
			func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i)
			},
		},

		{
			"incorrect",
			Gauge,
			[]byte(`"incorrect"`),
			func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, tt.t.UnmarshalJSON(tt.bytes), fmt.Sprintf("UnmarshalJSON(%v)", string(tt.bytes)))
		})
	}
}
