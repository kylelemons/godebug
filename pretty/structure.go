package pretty

import (
	"fmt"
	"io"
	"strings"
)

type Options struct {
	Extended bool

	// TODO(kevlar): implement:
	//QuoteKeys bool
}

type node interface {
	WriteTo(w io.Writer, indent string, opts *Options)
}

type stringVal string

func (str stringVal) WriteTo(w io.Writer, indent string, opts *Options) {
	fmt.Fprintf(w, "%q", string(str))
}

type rawVal string

func (r rawVal) WriteTo(w io.Writer, indent string, opts *Options) {
	fmt.Fprintf(w, "%s", string(r))
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

	fmt.Fprintf(w, "{")
	for i, kv := range l {
		if i > 0 || opts.Extended {
			fmt.Fprintf(w, "\n%s ", indent)
		}
		fmt.Fprintf(w, "%s:%s", kv.key, padding[len(kv.key):])
		kv.val.WriteTo(w, indent+"  "+padding, opts)
		if i+1 < len(l) || opts.Extended {
			fmt.Fprintf(w, ",")
		}
	}
	if opts.Extended && len(l) > 0 {
		fmt.Fprintf(w, "\n%s", indent)
	}
	fmt.Fprintf(w, "}")
}

type list []node

func (l list) WriteTo(w io.Writer, indent string, opts *Options) {
	inner := indent + " "
	fmt.Fprintf(w, "[")
	for i, v := range l {
		if i > 0 || opts.Extended {
			fmt.Fprintf(w, "\n%s", inner)
		}
		v.WriteTo(w, inner, opts)
		if i+1 < len(l) || opts.Extended {
			fmt.Fprintf(w, ",")
		}
	}
	if opts.Extended && len(l) > 0 {
		fmt.Fprintf(w, "\n%s", indent)
	}
	fmt.Fprintf(w, "]")
}
