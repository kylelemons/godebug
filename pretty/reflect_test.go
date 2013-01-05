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
			"map of [2]int",
			map[[2]int]string{
				[2]int{-1, 2}: "school",
				[2]int{0, 0}:  "origin",
				[2]int{1, 3}:  "home",
			},
			keyvals{
				{"[-1,2]", stringVal("school")},
				{"[0,0]", stringVal("origin")},
				{"[1,3]", stringVal("home")},
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
