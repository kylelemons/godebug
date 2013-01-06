package pretty

import (
	"bytes"
	"strconv"
	"strings"
)

// An Options represents optional configuration parameters for formatting.
//
// Some options, notably ShortList, dramatically increase the overhead
// of pretty-printing a value.
type Options struct {
	Compact  bool // One-line output. Overrides Diffable.
	Diffable bool // Adds extra newlines for more easily diffable output.

	ShortList int // Maximum character length for short lists if nonzero.
}

var DefaultOptions = &Options{}

func (o *Options) dup() *Options {
	o2 := *o
	return &o2
}

type node interface {
	WriteTo(w *bytes.Buffer, indent string, opts *Options)
}

func compactString(n node) string {
	switch k := n.(type) {
	case stringVal:
		return string(k)
	case rawVal:
		return string(k)
	}

	buf := new(bytes.Buffer)
	n.WriteTo(buf, "", &Options{Compact: true})
	return buf.String()
}

type stringVal string

func (str stringVal) WriteTo(w *bytes.Buffer, indent string, opts *Options) {
	w.WriteString(strconv.Quote(string(str)))
}

type rawVal string

func (r rawVal) WriteTo(w *bytes.Buffer, indent string, opts *Options) {
	w.WriteString(string(r))
}

type keyval struct {
	key string
	val node
}

type keyvals []keyval

func (l keyvals) Len() int           { return len(l) }
func (l keyvals) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l keyvals) Less(i, j int) bool { return l[i].key < l[j].key }

func (l keyvals) WriteTo(w *bytes.Buffer, indent string, opts *Options) {
	keyWidth := 0

	for _, kv := range l {
		if kw := len(kv.key); kw > keyWidth {
			keyWidth = kw
		}
	}
	padding := strings.Repeat(" ", keyWidth+1)

	inner := indent + "  " + padding
	w.WriteByte('{')
	for i, kv := range l {
		if opts.Compact {
			w.WriteString(kv.key)
			w.WriteByte(':')
		} else {
			if i > 0 || opts.Diffable {
				w.WriteString("\n ")
				w.WriteString(indent)
			}
			w.WriteString(kv.key)
			w.WriteByte(':')
			w.WriteString(padding[len(kv.key):])
		}
		kv.val.WriteTo(w, inner, opts)
		if i+1 < len(l) || opts.Diffable {
			w.WriteByte(',')
		}
	}
	if !opts.Compact && opts.Diffable && len(l) > 0 {
		w.WriteString("\n")
		w.WriteString(indent)
	}
	w.WriteByte('}')
}

type list []node

func (l list) WriteTo(w *bytes.Buffer, indent string, opts *Options) {
	if max := opts.ShortList; max > 0 {
		short := compactString(l)
		if len(short) <= max {
			w.WriteString(short)
			return
		}
	}

	inner := indent + " "
	w.WriteByte('[')
	for i, v := range l {
		if !opts.Compact && (i > 0 || opts.Diffable) {
			w.WriteByte('\n')
			w.WriteString(inner)
		}
		v.WriteTo(w, inner, opts)
		if i+1 < len(l) || opts.Diffable {
			w.WriteByte(',')
		}
	}
	if !opts.Compact && opts.Diffable && len(l) > 0 {
		w.WriteByte('\n')
		w.WriteString(indent)
	}
	w.WriteByte(']')
}
