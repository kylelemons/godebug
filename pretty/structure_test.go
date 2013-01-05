package pretty

import (
	"bytes"
	"testing"
)

func TestWriteTo(t *testing.T) {
	tests := []struct {
		desc     string
		node     node
		normal   string
		extended string
	}{
		{
			desc:     "string",
			node:     stringVal("zaphod"),
			normal:   `"zaphod"`,
			extended: `"zaphod"`,
		},
		{
			desc:     "raw",
			node:     rawVal("42"),
			normal:   `42`,
			extended: `42`,
		},
		{
			desc: "keyvals",
			node: keyvals{
				{"name", stringVal("zaphod")},
				{"age", rawVal("42")},
			},
			normal: `{name: "zaphod",
 age:  42}`,
			extended: `{
 name: "zaphod",
 age:  42,
}`,
		},
		{
			desc: "list",
			node: list{
				stringVal("zaphod"),
				rawVal("42"),
			},
			normal: `["zaphod",
 42]`,
			extended: `[
 "zaphod",
 42,
]`,
		},
		{
			desc: "nested",
			node: list{
				stringVal("first"),
				list{rawVal("1"), rawVal("2"), rawVal("3")},
				keyvals{
					{"trillian", keyvals{
						{"race", stringVal("human")},
						{"age", rawVal("36")},
					}},
					{"zaphod", keyvals{
						{"occupation", stringVal("president of the galaxy")},
						{"features", stringVal("two heads")},
					}},
				},
				keyvals{},
			},
			normal: `["first",
 [1,
  2,
  3],
 {trillian: {race: "human",
             age:  36},
  zaphod:   {occupation: "president of the galaxy",
             features:   "two heads"}},
 {}]`,
			extended: `[
 "first",
 [
  1,
  2,
  3,
 ],
 {
  trillian: {
             race: "human",
             age:  36,
            },
  zaphod:   {
             occupation: "president of the galaxy",
             features:   "two heads",
            },
 },
 {},
]`,
		},
	}

	for _, test := range tests {
		buf := new(bytes.Buffer)
		test.node.WriteTo(buf, "", &Options{})
		if got, want := buf.String(), test.normal; got != want {
			t.Errorf("%s: normal rendendered incorrectly\ngot:\n%s\nwant:\n%s", test.desc, got, want)
		}
		buf.Reset()
		test.node.WriteTo(buf, "", &Options{Extended: true})
		if got, want := buf.String(), test.extended; got != want {
			t.Errorf("%s: extended rendendered incorrectly\ngot:\n%s\nwant:\n%s", test.desc, got, want)
		}
	}
}
