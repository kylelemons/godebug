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

import "testing"

type goStringerOnly struct{}

func (goStringerOnly) GoString() string { return "go-stringer" }

type bothStringers struct{}

func (bothStringers) String() string   { return "stringer" }
func (bothStringers) GoString() string { return "go-stringer" }

func TestGoStringer(t *testing.T) {
	cfg := &Config{PrintGoStringers: true, Compact: true}
	if got, want := cfg.Sprint(goStringerOnly{}), `"go-stringer"`; got != want {
		t.Errorf("Sprint(goStringerOnly{}) = %q, want %q", got, want)
	}
}

func TestGoStringerTakesPriority(t *testing.T) {
	cfg := &Config{PrintGoStringers: true, PrintStringers: true, Compact: true}
	if got, want := cfg.Sprint(bothStringers{}), `"go-stringer"`; got != want {
		t.Errorf("Sprint(bothStringers{}) = %q, want %q", got, want)
	}
}

func TestStringerUsedWhenGoStringerDisabled(t *testing.T) {
	cfg := &Config{PrintStringers: true, Compact: true}
	if got, want := cfg.Sprint(bothStringers{}), `"stringer"`; got != want {
		t.Errorf("Sprint(bothStringers{}) = %q, want %q", got, want)
	}
}
