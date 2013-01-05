package pretty

import (
	"bytes"
	"io"
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
	WriteTo(w io.Writer, indent string, opts *Options)
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

func (str stringVal) WriteTo(w io.Writer, indent string, opts *Options) {
	io.WriteString(w, strconv.Quote(string(str)))
}

type rawVal string

func (r rawVal) WriteTo(w io.Writer, indent string, opts *Options) {
	io.WriteString(w, string(r))
}

type keyval struct {
	key string
	val node
}

type keyvals []keyval

func (l keyvals) Len() int           { return len(l) }
func (l keyvals) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l keyvals) Less(i, j int) bool { return l[i].key < l[j].key }

func (l keyvals) WriteTo(w io.Writer, indent string, opts *Options) {
	keyWidth := 0

	for _, kv := range l {
		if kw := len(kv.key); kw > keyWidth {
			keyWidth = kw
		}
	}
	padding := strings.Repeat(" ", keyWidth+1)

	inner := indent + "  " + padding
	io.WriteString(w, "{")
	for i, kv := range l {
		if opts.Compact {
			io.WriteString(w, kv.key)
			io.WriteString(w, ":")
		} else {
			if i > 0 || opts.Diffable {
				io.WriteString(w, "\n ")
				io.WriteString(w, indent)
			}
			io.WriteString(w, kv.key)
			io.WriteString(w, ":")
			io.WriteString(w, padding[len(kv.key):])
		}
		kv.val.WriteTo(w, inner, opts)
		if i+1 < len(l) || opts.Diffable {
			io.WriteString(w, ",")
		}
	}
	if !opts.Compact && opts.Diffable && len(l) > 0 {
		io.WriteString(w, "\n")
		io.WriteString(w, indent)
	}
	io.WriteString(w, "}")
}

type list []node

func (l list) WriteTo(w io.Writer, indent string, opts *Options) {
	if max := opts.ShortList; max > 0 {
		short := compactString(l)
		if len(short) <= max {
			io.WriteString(w, short)
			return
		}
	}

	inner := indent + " "
	io.WriteString(w, "[")
	for i, v := range l {
		if !opts.Compact && (i > 0 || opts.Diffable) {
			io.WriteString(w, "\n")
			io.WriteString(w, inner)
		}
		v.WriteTo(w, inner, opts)
		if i+1 < len(l) || opts.Diffable {
			io.WriteString(w, ",")
		}
	}
	if !opts.Compact && opts.Diffable && len(l) > 0 {
		io.WriteString(w, "\n")
		io.WriteString(w, indent)
	}
	io.WriteString(w, "]")
}
