package pretty

import (
	"bytes"
	"strconv"
	"strings"
)

// A Config represents optional configuration parameters for formatting.
//
// Some options, notably ShortList, dramatically increase the overhead
// of pretty-printing a value.
type Config struct {
	Compact  bool // One-line output. Overrides Diffable.
	Diffable bool // Adds extra newlines for more easily diffable output.

	ShortList int // Maximum character length for short lists if nonzero.
}

var DefaultConfig = &Config{}

type node interface {
	WriteTo(w *bytes.Buffer, indent string, cfg *Config)
}

func compactString(n node) string {
	switch k := n.(type) {
	case stringVal:
		return string(k)
	case rawVal:
		return string(k)
	}

	buf := new(bytes.Buffer)
	n.WriteTo(buf, "", &Config{Compact: true})
	return buf.String()
}

type stringVal string

func (str stringVal) WriteTo(w *bytes.Buffer, indent string, cfg *Config) {
	w.WriteString(strconv.Quote(string(str)))
}

type rawVal string

func (r rawVal) WriteTo(w *bytes.Buffer, indent string, cfg *Config) {
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

func (l keyvals) WriteTo(w *bytes.Buffer, indent string, cfg *Config) {
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
		if cfg.Compact {
			w.WriteString(kv.key)
			w.WriteByte(':')
		} else {
			if i > 0 || cfg.Diffable {
				w.WriteString("\n ")
				w.WriteString(indent)
			}
			w.WriteString(kv.key)
			w.WriteByte(':')
			w.WriteString(padding[len(kv.key):])
		}
		kv.val.WriteTo(w, inner, cfg)
		if i+1 < len(l) || cfg.Diffable {
			w.WriteByte(',')
		}
	}
	if !cfg.Compact && cfg.Diffable && len(l) > 0 {
		w.WriteString("\n")
		w.WriteString(indent)
	}
	w.WriteByte('}')
}

type list []node

func (l list) WriteTo(w *bytes.Buffer, indent string, cfg *Config) {
	if max := cfg.ShortList; max > 0 {
		short := compactString(l)
		if len(short) <= max {
			w.WriteString(short)
			return
		}
	}

	inner := indent + " "
	w.WriteByte('[')
	for i, v := range l {
		if !cfg.Compact && (i > 0 || cfg.Diffable) {
			w.WriteByte('\n')
			w.WriteString(inner)
		}
		v.WriteTo(w, inner, cfg)
		if i+1 < len(l) || cfg.Diffable {
			w.WriteByte(',')
		}
	}
	if !cfg.Compact && cfg.Diffable && len(l) > 0 {
		w.WriteByte('\n')
		w.WriteString(indent)
	}
	w.WriteByte(']')
}
