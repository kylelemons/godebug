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

package diff

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestDiff(t *testing.T) {
	tests := []struct {
		desc   string
		A, B   []string
		chunks []Chunk
	}{
		{
			desc: "nil",
		},
		{
			desc: "empty",
			A:    []string{},
			B:    []string{},
		},
		{
			desc: "same",
			A:    []string{"foo"},
			B:    []string{"foo"},
		},
		{
			desc: "a empty",
			A:    []string{},
		},
		{
			desc: "b empty",
			B:    []string{},
		},
		{
			desc: "b nil",
			A:    []string{"foo"},
			chunks: []Chunk{
				0: {Deleted: []string{"foo"}},
			},
		},
		{
			desc: "a nil",
			B:    []string{"foo"},
			chunks: []Chunk{
				0: {Added: []string{"foo"}},
			},
		},
		{
			desc: "b nil multiple",
			A:    []string{"foo", "bar", "baz"},
			chunks: []Chunk{
				0: {Deleted: []string{"foo"}},
				1: {Deleted: []string{"bar"}},
				2: {Deleted: []string{"baz"}},
			},
		},
		{
			desc: "a nil multiple",
			B:    []string{"foo", "bar", "baz"},
			chunks: []Chunk{
				0: {Added: []string{"foo"}},
				1: {Added: []string{"bar"}},
				2: {Added: []string{"baz"}},
			},
		},
		{
			desc: "b empty multiple",
			A:    []string{"alpha", "beta"},
			B:    []string{},
			chunks: []Chunk{
				0: {Deleted: []string{"alpha"}},
				1: {Deleted: []string{"beta"}},
			},
		},
		{
			desc: "a empty multiple",
			A:    []string{},
			B:    []string{"alpha", "beta"},
			chunks: []Chunk{
				0: {Added: []string{"alpha"}},
				1: {Added: []string{"beta"}},
			},
		},
		{
			desc: "start with change",
			A:    []string{"a", "b", "c"},
			B:    []string{"A", "b", "c"},
			chunks: []Chunk{
				0: {Deleted: []string{"a"}},
				1: {Added: []string{"A"}, Equal: []string{"b", "c"}},
			},
		},
		{
			desc: "constitution",
			A: []string{
				"We the People of the United States, in Order to form a more perfect Union,",
				"establish Justice, insure domestic Tranquility, provide for the common defence,",
				"and secure the Blessings of Liberty to ourselves",
				"and our Posterity, do ordain and establish this Constitution for the United",
				"States of America.",
			},
			B: []string{
				"We the People of the United States, in Order to form a more perfect Union,",
				"establish Justice, insure domestic Tranquility, provide for the common defence,",
				"promote the general Welfare, and secure the Blessings of Liberty to ourselves",
				"and our Posterity, do ordain and establish this Constitution for the United",
				"States of America.",
			},
			chunks: []Chunk{
				0: {
					Equal: []string{
						"We the People of the United States, in Order to form a more perfect Union,",
						"establish Justice, insure domestic Tranquility, provide for the common defence,",
					},
				},
				1: {
					Deleted: []string{
						"and secure the Blessings of Liberty to ourselves",
					},
				},
				2: {
					Added: []string{
						"promote the general Welfare, and secure the Blessings of Liberty to ourselves",
					},
					Equal: []string{
						"and our Posterity, do ordain and establish this Constitution for the United",
						"States of America.",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got := DiffChunks(test.A, test.B)
			if got, want := len(got), len(test.chunks); got != want {
				t.Errorf("edit distance = %v, want %v", got-1, want-1)
				return
			}
			for i := range got {
				got, want := got[i], test.chunks[i]
				if got, want := got.Added, want.Added; !reflect.DeepEqual(got, want) {
					t.Errorf("chunks[%d]: Added = %v, want %v", i, got, want)
				}
				if got, want := got.Deleted, want.Deleted; !reflect.DeepEqual(got, want) {
					t.Errorf("chunks[%d]: Deleted = %v, want %v", i, got, want)
				}
				if got, want := got.Equal, want.Equal; !reflect.DeepEqual(got, want) {
					t.Errorf("chunks[%d]: Equal = %v, want %v", i, got, want)
				}
			}
		})
	}
}

func TestRender(t *testing.T) {
	tests := []struct {
		desc   string
		chunks []Chunk
		out    string
	}{
		{
			desc: "ordering",
			chunks: []Chunk{
				{
					Added:   []string{"1"},
					Deleted: []string{"2"},
					Equal:   []string{"3"},
				},
				{
					Added:   []string{"4"},
					Deleted: []string{"5"},
				},
			},
			out: strings.TrimSpace(`
+1
-2
 3
+4
-5
			`),
		},
		{
			desc: "only_added",
			chunks: []Chunk{
				{
					Added: []string{"1"},
				},
			},
			out: strings.TrimSpace(`
+1
			`),
		},
		{
			desc: "only_deleted",
			chunks: []Chunk{
				{
					Deleted: []string{"1"},
				},
			},
			out: strings.TrimSpace(`
-1
			`),
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			if got, want := Render(test.chunks), test.out; got != want {
				t.Errorf("Render(%q):", test.chunks)
				t.Errorf("GOT\n%s", got)
				t.Errorf("WANT\n%s", want)
			}
		})
	}
}

func ExampleDiff() {
	constitution := strings.TrimSpace(`
We the People of the United States, in Order to form a more perfect Union,
establish Justice, insure domestic Tranquility, provide for the common defence,
promote the general Welfare, and secure the Blessings of Liberty to ourselves
and our Posterity, do ordain and establish this Constitution for the United
States of America.
`)

	got := strings.TrimSpace(`
:wq
We the People of the United States, in Order to form a more perfect Union,
establish Justice, insure domestic Tranquility, provide for the common defence,
and secure the Blessings of Liberty to ourselves
and our Posterity, do ordain and establish this Constitution for the United
States of America.
`)

	fmt.Println(Diff(got, constitution))

	// Output:
	// -:wq
	//  We the People of the United States, in Order to form a more perfect Union,
	//  establish Justice, insure domestic Tranquility, provide for the common defence,
	// -and secure the Blessings of Liberty to ourselves
	// +promote the general Welfare, and secure the Blessings of Liberty to ourselves
	//  and our Posterity, do ordain and establish this Constitution for the United
	//  States of America.
}

func TestDiffEmptyInputs(t *testing.T) {
	tests := []struct {
		desc string
		a, b string
		want string
	}{
		{
			desc: "both empty",
			a:    "",
			b:    "",
			want: "",
		},
		{
			desc: "identical non-empty",
			a:    "alpha\nbeta",
			b:    "alpha\nbeta",
			want: "",
		},
		{
			desc: "a empty single line",
			a:    "",
			b:    "hello",
			want: "+hello",
		},
		{
			desc: "b empty single line",
			a:    "hello",
			b:    "",
			want: "-hello",
		},
		{
			desc: "a empty multiline",
			a:    "",
			b:    "alpha\nbeta\ngamma",
			want: "+alpha\n+beta\n+gamma",
		},
		{
			desc: "b empty multiline",
			a:    "alpha\nbeta\ngamma",
			b:    "",
			want: "-alpha\n-beta\n-gamma",
		},
		{
			desc: "a empty with blank lines",
			a:    "",
			b:    "alpha\n\nbeta",
			want: "+alpha\n+\n+beta",
		},
		{
			desc: "b empty with blank lines",
			a:    "alpha\n\nbeta",
			b:    "",
			want: "-alpha\n-\n-beta",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := Diff(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("Diff(%q, %q) = %q, want %q", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestDiffIdenticalFormatting(t *testing.T) {
	cases := []string{
		"",
		"single line",
		"line 1\nline 2\nline 3",
		"start\n\nmiddle\n\nend",
	}

	for _, text := range cases {
		// When A is empty, Diff should match Render(DiffChunks(nil, lines))
		lines := strings.Split(text, "\n")
		var expectedAdditions string
		if text != "" {
			expectedAdditions = Render(DiffChunks(nil, lines))
		}
		if got := Diff("", text); got != expectedAdditions {
			t.Errorf("Diff(\"\", %q) = %q, want %q", text, got, expectedAdditions)
		}

		// When B is empty, Diff should match Render(DiffChunks(lines, nil))
		var expectedDeletions string
		if text != "" {
			expectedDeletions = Render(DiffChunks(lines, nil))
		}
		if got := Diff(text, ""); got != expectedDeletions {
			t.Errorf("Diff(%q, \"\") = %q, want %q", text, got, expectedDeletions)
		}
	}
}

func BenchmarkDiffEmptyInputs(b *testing.B) {
	lines := make([]string, 1000)
	for i := range lines {
		lines[i] = fmt.Sprintf("line payload item %d", i)
	}
	text := strings.Join(lines, "\n")

	b.Run("Diff A empty", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Diff("", text)
		}
	})

	b.Run("Diff B empty", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Diff(text, "")
		}
	})

	b.Run("DiffChunks A empty", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = DiffChunks(nil, lines)
		}
	})

	b.Run("DiffChunks B empty", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = DiffChunks(lines, nil)
		}
	})
}
