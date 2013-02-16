package pretty

import (
	"reflect"
	"strings"
	"testing"
)

func TestVal2nodeDefault(t *testing.T) {
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
			struct{ Zaphod, Ford string }{"beeblebrox", "prefect"},
			keyvals{
				{"Zaphod", stringVal("beeblebrox")},
				{"Ford", stringVal("prefect")},
			},
		},
		{
			"int",
			3,
			rawVal("3"),
		},
	}

	for _, test := range tests {
		if got, want := DefaultConfig.val2node(reflect.ValueOf(test.raw), emptyset), test.want; !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %#v, want %#v", test.desc, got, want)
		}
	}
}

func TestVal2node(t *testing.T) {
	tests := []struct {
		desc string
		raw  interface{}
		cfg  *Config
		want node
	}{
		{
			"struct default",
			struct{ Zaphod, Ford, foo string }{"beeblebrox", "prefect", "BAD"},
			DefaultConfig,
			keyvals{
				{"Zaphod", stringVal("beeblebrox")},
				{"Ford", stringVal("prefect")},
			},
		},
		{
			"struct w/ unexported",
			struct{ Zaphod, Ford, foo string }{"beeblebrox", "prefect", "GOOD"},
			&Config{
				IncludeUnexported: true,
			},
			keyvals{
				{"Zaphod", stringVal("beeblebrox")},
				{"Ford", stringVal("prefect")},
				{"foo", stringVal("GOOD")},
			},
		},
	}

	for _, test := range tests {
		if got, want := test.cfg.val2node(reflect.ValueOf(test.raw), emptyset), test.want; !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %#v, want %#v", test.desc, got, want)
		}
	}
}

func TestRecursion(t *testing.T) {

	type Recursive struct {
		R *Recursive
	}

	r := &Recursive{}
	r.R = r
	x := Sprint(r)

	if !strings.Contains(x, "recursion on") {
		t.Errorf("Expected formatted string to contain 'recursion on', got: %q",
			x)
	}

}
