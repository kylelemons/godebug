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

type reflector struct {
	*Config

	// whether an address was seen while traversing current path:
	//
	// key not present - have not seen this addr at all
	// nil - seen but label not yet created
	// !nil - label created
	addrs map[uintptr]*label
}

// follow handles following a possiblly-recursive reference to the given value
// from the given ptr address.
func (ref *reflector) follow(ptr uintptr, val reflect.Value) (n node) {
	if ref.addrs == nil {
		// pointer tracking disabled
		return ref.val2node(val)
	}

	// if ptr was already seen - just create a reference to it
	if l, seen := ref.addrs[ptr]; seen {
		if l == nil {
			l = &label{id: 0, node: nil}
			ref.addrs[ptr] = l
		}

		return &refto{label: l}
	}

	// recursively convert value to node and if value was
	// pointer-referenced inside bind label to created node.
	ref.addrs[ptr] = nil
	n = ref.val2node(val)
	l := ref.addrs[ptr]
	if l != nil {
		l.node = n
		n = l
	}
	delete(ref.addrs, ptr)
	return n
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
