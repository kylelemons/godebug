// Copyright 2013 Google Inc.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pretty

import (
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"
)

func TestVal2nodeDefault(t *testing.T) {
	err := fmt.Errorf("err")
	var errNil error

	tests := []struct {
		desc string
		raw  interface{}
		want node
	}{
		{
			desc: "nil",
			raw:  nil,
			want: rawVal("nil"),
		},
		{
			desc: "nil ptr",
			raw:  (*int)(nil),
			want: rawVal("nil"),
		},
		{
			desc: "nil slice",
			raw:  []string(nil),
			want: list{},
		},
		{
			desc: "nil map",
			raw:  map[string]string(nil),
			want: keyvals{},
		},
		{
			desc: "string",
			raw:  "zaphod",
			want: stringVal("zaphod"),
		},
		{
			desc: "slice",
			raw:  []string{"a", "b"},
			want: list{stringVal("a"), stringVal("b")},
		},
		{
			desc: "map",
			raw: map[string]string{
				"zaphod": "beeblebrox",
				"ford":   "prefect",
			},
			want: keyvals{
				{"ford", stringVal("prefect")},
				{"zaphod", stringVal("beeblebrox")},
			},
		},
		{
			desc: "map of [2]int",
			raw: map[[2]int]string{
				[2]int{-1, 2}: "school",
				[2]int{0, 0}:  "origin",
				[2]int{1, 3}:  "home",
			},
			want: keyvals{
				{"[-1,2]", stringVal("school")},
				{"[0,0]", stringVal("origin")},
				{"[1,3]", stringVal("home")},
			},
		},
		{
			desc: "struct",
			raw:  struct{ Zaphod, Ford string }{"beeblebrox", "prefect"},
			want: keyvals{
				{"Zaphod", stringVal("beeblebrox")},
				{"Ford", stringVal("prefect")},
			},
		},
		{
			desc: "int",
			raw:  3,
			want: rawVal("3"),
		},
		{
			desc: "time.Time",
			raw:  time.Unix(1257894000, 0).UTC(),
			want: rawVal("2009-11-10 23:00:00 +0000 UTC"),
		},
		{
			desc: "net.IP",
			raw:  net.IPv4(127, 0, 0, 1),
			want: rawVal("127.0.0.1"),
		},
		{
			desc: "error",
			raw:  &err,
			want: rawVal("err"),
		},
		{
			desc: "nil error",
			raw:  &errNil,
			want: rawVal("<nil>"),
		},
	}

	for _, test := range tests {
		ref := &reflector{
			Config: DefaultConfig,
		}
		if got, want := ref.val2node(reflect.ValueOf(test.raw)), test.want; !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %#v, want %#v", test.desc, got, want)
		}
	}
}

func TestVal2node(t *testing.T) {
	tests := []struct {
		desc  string
		raw   interface{}
		cfg   *Config
		addrs map[uintptr]*label
		want  node
	}{
		{
			desc: "struct default",
			raw:  struct{ Zaphod, Ford, foo string }{"beeblebrox", "prefect", "BAD"},
			cfg:  DefaultConfig,
			want: keyvals{
				{"Zaphod", stringVal("beeblebrox")},
				{"Ford", stringVal("prefect")},
			},
		},
		{
			desc: "struct w/ IncludeUnexported",
			raw:  struct{ Zaphod, Ford, foo string }{"beeblebrox", "prefect", "GOOD"},
			cfg: &Config{
				IncludeUnexported: true,
			},
			want: keyvals{
				{"Zaphod", stringVal("beeblebrox")},
				{"Ford", stringVal("prefect")},
				{"foo", stringVal("GOOD")},
			},
		},
		{
			desc: "time default",
			raw:  struct{ Date time.Time }{time.Unix(1234567890, 0).UTC()},
			cfg:  DefaultConfig,
			want: keyvals{
				{"Date", rawVal("2009-02-13 23:31:30 +0000 UTC")},
			},
		},
		{
			desc: "time w/ nil Formatter",
			raw:  struct{ Date time.Time }{time.Unix(1234567890, 0).UTC()},
			cfg: &Config{
				PrintStringers: true,
				Formatter: map[reflect.Type]interface{}{
					reflect.TypeOf(time.Time{}): nil,
				},
			},
			want: keyvals{
				{"Date", keyvals{}},
			},
		},
		{
			desc: "time w/ PrintTextMarshalers",
			raw:  struct{ Date time.Time }{time.Unix(1234567890, 0).UTC()},
			cfg: &Config{
				PrintTextMarshalers: true,
			},
			want: keyvals{
				{"Date", stringVal("2009-02-13T23:31:30Z")},
			},
		},
		{
			desc: "time w/ PrintStringers",
			raw:  struct{ Date time.Time }{time.Unix(1234567890, 0).UTC()},
			cfg: &Config{
				PrintStringers: true,
			},
			want: keyvals{
				{"Date", stringVal("2009-02-13 23:31:30 +0000 UTC")},
			},
		},
		{
			desc: "circular list",
			raw:  circular(3),
			cfg:  DefaultConfig,
			addrs: make(map[uintptr]*label),
			want: func() node {
				l := &label{}
				l.node = keyvals{
					{"Value", rawVal("1")},
					{"Next", keyvals{
						{"Value", rawVal("2")},
						{"Next", keyvals{
							{"Value", rawVal("3")},
							{"Next", &refto{l}},
						}},
					}},
				}
				return l
			}(),
		},
		{
			desc: "non-circular dup reference",
			raw: func() interface{} {
				type Object struct{ X int }
				type ObjPair struct{ Obj1, Obj2 interface{} }
				obj := &Object{11}
				return &ObjPair{obj, obj}
			}(),
			cfg: DefaultConfig,
			addrs: make(map[uintptr]*label),
			want: keyvals{
				{"Obj1", keyvals{{"X", rawVal("11")}}},
				{"Obj2", keyvals{{"X", rawVal("11")}}},
			},
		},
	}

	for _, test := range tests {
		ref := &reflector{
			Config: test.cfg,
			addrs:  test.addrs,
		}
		if got, want := ref.val2node(reflect.ValueOf(test.raw)), test.want; !reflect.DeepEqual(got, want) {
			t.Run(test.desc, func(t *testing.T) {
				t.Errorf(" got %#v", got)
				t.Errorf("want %#v", want)
				t.Errorf("Diff: (-got +want)\n%s", Compare(got, want))
			})
		}
	}
}

type ListNode struct {
	Value int
	Next  *ListNode
}

func circular(nodes int) *ListNode {
	final := &ListNode{
		Value: nodes,
	}
	final.Next = final

	recent := final
	for i := nodes - 1; i > 0; i-- {
		n := &ListNode{
			Value: i,
			Next:  recent,
		}
		final.Next = n
		recent = n
	}
	return recent
}

func BenchmarkVal2node(b *testing.B) {
	benchmarks := []struct {
		desc string
		cfg  *Config
		raw  interface{}
	}{
		{
			desc: "acyclic/struct",
			cfg:  Acyclic,
			raw:  struct{ Zaphod, Ford string }{"beeblebrox", "prefect"},
		},
		{
			desc: "acyclic/map",
			cfg:  Acyclic,
			raw: map[[2]int]string{
				[2]int{-1, 2}: "school",
				[2]int{0, 0}:  "origin",
				[2]int{1, 3}:  "home",
			},
		},
		{
			desc: "struct",
			cfg:  DefaultConfig,
			raw:  struct{ Zaphod, Ford string }{"beeblebrox", "prefect"},
		},
		{
			desc: "map",
			cfg:  DefaultConfig,
			raw: map[[2]int]string{
				[2]int{-1, 2}: "school",
				[2]int{0, 0}:  "origin",
				[2]int{1, 3}:  "home",
			},
		},
		{
			desc: "circlist/small",
			cfg:  DefaultConfig,
			raw:  circular(3),
		},
		{
			desc: "circlist/med",
			cfg:  DefaultConfig,
			raw:  circular(300),
		},
		{
			desc: "circlist/large",
			cfg:  DefaultConfig,
			raw:  circular(3000),
		},
	}

	for _, bench := range benchmarks {
		b.Run(bench.desc, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				ref := &reflector{
					Config: bench.cfg,
				}
				if !bench.cfg.AssumeAcyclic {
					ref.addrs = make(map[uintptr]*label)
				}
				ref.val2node(reflect.ValueOf(bench.raw))
			}
		})
	}
}
