package pretty

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
)

// Reflect returns the string representation of the given value with the given
// options.  If cfg is nil, the DefaultConfig are used.
func Reflect(val interface{}, cfg *Config) string {
	if cfg == nil {
		cfg = DefaultConfig
	}

	node := val2node(reflect.ValueOf(val))

	buf := new(bytes.Buffer)
	node.WriteTo(buf, "", cfg)
	return buf.String()
}

func val2node(val reflect.Value) node {
	// TODO(kevlar): pointer tracking?

	switch kind := val.Kind(); kind {
	case reflect.Ptr, reflect.Interface:
		if val.IsNil() {
			return rawVal("nil")
		}
		return val2node(val.Elem())
	case reflect.String:
		return stringVal(val.String())
	case reflect.Slice, reflect.Array:
		n := list{}
		length := val.Len()
		for i := 0; i < length; i++ {
			n = append(n, val2node(val.Index(i)))
		}
		return n
	case reflect.Map:
		n := keyvals{}
		keys := val.MapKeys()
		for _, key := range keys {
			// TODO(kevlar): Support arbitrary type keys?
			n = append(n, keyval{compactString(val2node(key)), val2node(val.MapIndex(key))})
		}
		sort.Sort(n)
		return n
	case reflect.Struct:
		n := keyvals{}
		typ := val.Type()
		fields := typ.NumField()
		for i := 0; i < fields; i++ {
			n = append(n, keyval{typ.Field(i).Name, val2node(val.Field(i))})
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
