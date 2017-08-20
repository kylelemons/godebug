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
	"encoding"
	"fmt"
	"reflect"
	"sort"
)

func isZeroVal(val reflect.Value) bool {
	if !val.CanInterface() {
		return false
	}
	z := reflect.Zero(val.Type()).Interface()
	return reflect.DeepEqual(val.Interface(), z)
}

type pointerTracker struct {
	addrs map[uintptr]int // addr[address] = seen count
}

// track tracks following a reference (pointer, slice, map, etc).  Every call to
// track should be paired with a call to untrack.
func (p *pointerTracker) track(ptr uintptr) {
	if p.addrs == nil {
		p.addrs = make(map[uintptr]int)
	}
	p.addrs[ptr]++
}

// seen returns whether the pointer was previously seen along this path.
func (p *pointerTracker) seen(ptr uintptr) bool {
	_, ok := p.addrs[ptr]
	return ok
}

// untrack registers that we have backtracked over the reference to the pointer.
func (p *pointerTracker) untrack(ptr uintptr) {
	p.addrs[ptr]--
	if p.addrs[ptr] == 0 {
		delete(p.addrs, ptr)
	}
}

type reflector struct {
	*Config
	*pointerTracker

	inContext        bool
	remainingContext int
}

// follow handles following a possiblly-recursive reference to the given value
// from the given ptr address.
func (ref *reflector) follow(ptr uintptr, val reflect.Value) (n node) {
	if ref.pointerTracker == nil {
		// Tracking disabled
		return ref.val2node(val)
	}

	if ref.inContext {
		// Decrement the context and restore it
		ref.remainingContext--
		defer func() {
			ref.remainingContext++
		}()

		// If we've exhausted our context, don't recurse
		if ref.remainingContext <= 0 {
			return rawVal("...")
		}
	} else if ref.seen(ptr) {
		// Make note that we're now in "context mode"
		ref.inContext = true
		ref.remainingContext = ref.RecursiveContext
		// Wrap the return value with the recursive marker
		defer func() {
			ref.inContext = false
			n = recursive{
				value: n,
			}
		}()
	}

	ref.track(ptr)
	defer ref.untrack(ptr)
	return ref.val2node(val)
}

func (ref *reflector) val2node(val reflect.Value) node {
	if !val.IsValid() {
		return rawVal("nil")
	}

	if val.CanInterface() {
		v := val.Interface()
		if formatter, ok := ref.Formatter[val.Type()]; ok {
			if formatter != nil {
				res := reflect.ValueOf(formatter).Call([]reflect.Value{val})
				return rawVal(res[0].Interface().(string))
			}
		} else {
			if s, ok := v.(fmt.Stringer); ok && ref.PrintStringers {
				return stringVal(s.String())
			}
			if t, ok := v.(encoding.TextMarshaler); ok && ref.PrintTextMarshalers {
				if raw, err := t.MarshalText(); err == nil { // if NOT an error
					return stringVal(string(raw))
				}
			}
		}
	}

	switch kind := val.Kind(); kind {
	case reflect.Ptr:
		if val.IsNil() {
			return rawVal("nil")
		}
		return ref.follow(val.Pointer(), val.Elem())
	case reflect.Interface:
		if val.IsNil() {
			return rawVal("nil")
		}
		return ref.val2node(val.Elem())
	case reflect.String:
		return stringVal(val.String())
	case reflect.Slice:
		n := list{}
		length := val.Len()
		ptr := val.Pointer()
		for i := 0; i < length; i++ {
			n = append(n, ref.follow(ptr, val.Index(i)))
		}
		return n
	case reflect.Array:
		n := list{}
		length := val.Len()
		for i := 0; i < length; i++ {
			n = append(n, ref.val2node(val.Index(i)))
		}
		return n
	case reflect.Map:
		n := keyvals{}
		keys := val.MapKeys()
		ptr := val.Pointer()
		for _, key := range keys {
			// TODO(kevlar): Support arbitrary type keys?
			n = append(n, keyval{
				key: compactString(ref.follow(ptr, key)),
				val: ref.follow(ptr, val.MapIndex(key)),
			})
		}
		sort.Sort(n)
		return n
	case reflect.Struct:
		n := keyvals{}
		typ := val.Type()
		fields := typ.NumField()
		for i := 0; i < fields; i++ {
			sf := typ.Field(i)
			if !ref.IncludeUnexported && sf.PkgPath != "" {
				continue
			}
			field := val.Field(i)
			if ref.SkipZeroFields && isZeroVal(field) {
				continue
			}
			n = append(n, keyval{sf.Name, ref.val2node(field)})
		}
		return n
	case reflect.Bool:
		if val.Bool() {
			return rawVal("true")
		}
		return rawVal("false")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rawVal(fmt.Sprintf("%d", val.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rawVal(fmt.Sprintf("%d", val.Uint()))
	case reflect.Uintptr:
		return rawVal(fmt.Sprintf("0x%X", val.Uint()))
	case reflect.Float32, reflect.Float64:
		return rawVal(fmt.Sprintf("%v", val.Float()))
	case reflect.Complex64, reflect.Complex128:
		return rawVal(fmt.Sprintf("%v", val.Complex()))
	}

	// Fall back to the default %#v if we can
	if val.CanInterface() {
		return rawVal(fmt.Sprintf("%#v", val.Interface()))
	}

	return rawVal(val.String())
}
