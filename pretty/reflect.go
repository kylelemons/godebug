package pretty

import (
	"fmt"
	"reflect"
	"sort"
)

func (c *Config) val2node(val reflect.Value, seen map[uintptr]bool) node {

	switch kind := val.Kind(); kind {
	case reflect.Ptr, reflect.Interface:
		if val.IsNil() {
			return rawVal("nil")
		}
		if val.Kind() == reflect.Ptr {
			p := val.Pointer()
			if present, _ := seen[p]; present {
				return rawVal(fmt.Sprintf("[recursion on 0x%x]", p))
			}
			seen[p] = true
		}
		return c.val2node(val.Elem(), seen)
	case reflect.String:
		return stringVal(val.String())
	case reflect.Slice, reflect.Array:
		n := list{}
		length := val.Len()
		for i := 0; i < length; i++ {
			n = append(n, c.val2node(val.Index(i), seen))
		}
		return n
	case reflect.Map:
		n := keyvals{}
		keys := val.MapKeys()
		for _, key := range keys {
			// TODO(kevlar): Support arbitrary type keys?
			n = append(n, keyval{compactString(c.val2node(key, seen)), c.val2node(val.MapIndex(key), seen)})
		}
		sort.Sort(n)
		return n
	case reflect.Struct:
		n := keyvals{}
		typ := val.Type()
		fields := typ.NumField()
		for i := 0; i < fields; i++ {
			sf := typ.Field(i)
			if !c.IncludeUnexported && sf.PkgPath != "" {
				continue
			}
			n = append(n, keyval{sf.Name, c.val2node(val.Field(i), seen)})
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
