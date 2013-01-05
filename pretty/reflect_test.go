package pretty

import (
	"reflect"
	"testing"
)

func TestReflect(t *testing.T) {
	tests := []struct {
		desc string
		raw  interface{}
		want node
	}{
		{
			"nil",
			(*int)(nil),
			rawVal("nil"),
		},
		{
			"string",
			"zaphod",
			stringVal("zaphod"),
		},
		{
			"slice",
			[]string{"a", "b"},
			list{stringVal("a"), stringVal("b")},
		},
		{
			"map",
			map[string]string{
				"zaphod": "beeblebrox",
				"ford":   "prefect",
			},
			keyvals{
				{"ford", stringVal("prefect")},
				{"zaphod", stringVal("beeblebrox")},
			},
		},
		{
			"struct",
			struct{ zaphod, ford string }{"beeblebrox", "prefect"},
			keyvals{
				{"zaphod", stringVal("beeblebrox")},
				{"ford", stringVal("prefect")},
			},
		},
		{
			"int",
			3,
			rawVal("3"),
		},
	}

	for _, test := range tests {
		if got, want := val2node(reflect.ValueOf(test.raw)), test.want; !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %#v, want %#v", test.desc, got, want)
		}
	}
}
