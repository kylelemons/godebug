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
	"bytes"
	"strconv"
	"strings"
)

type node interface {
	WriteTo(w *bytes.Buffer, indent string, wout *writeout)
}

// writeout holds configuration and state needed for nodes write-out process.
type writeout struct {
	*Config

	// labels are initially created with unallocated id.
	// curLabelID is used at writeout to allocate ids to labels serially
	// along the way as they are being written to output.
	curLabelID int64
}

func compactString(n node) string {
	switch k := n.(type) {
	case stringVal:
		return string(k)
	case rawVal:
		return string(k)
	}

	buf := new(bytes.Buffer)
	n.WriteTo(buf, "", &writeout{Config: &Config{Compact: true}})
	return buf.String()
}

type stringVal string

func (str stringVal) WriteTo(w *bytes.Buffer, indent string, wout *writeout) {
	w.WriteString(strconv.Quote(string(str)))
}

type rawVal string

func (r rawVal) WriteTo(w *bytes.Buffer, indent string, wout *writeout) {
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

func (l keyvals) WriteTo(w *bytes.Buffer, indent string, wout *writeout) {
	w.WriteByte('{')

	switch {
	case wout.Compact:
		// All on one line:
		for i, kv := range l {
			if i > 0 {
				w.WriteByte(',')
			}
			w.WriteString(kv.key)
			w.WriteByte(':')
			kv.val.WriteTo(w, indent, wout)
		}
	case wout.Diffable:
		w.WriteByte('\n')
		inner := indent + " "
		// Each value gets its own line:
		for _, kv := range l {
			w.WriteString(inner)
			w.WriteString(kv.key)
			w.WriteString(": ")
			kv.val.WriteTo(w, inner, wout)
			w.WriteString(",\n")
		}
		w.WriteString(indent)
	default:
		keyWidth := 0
		for _, kv := range l {
			if kw := len(kv.key); kw > keyWidth {
				keyWidth = kw
			}
		}
		alignKey := indent + " "
		alignValue := strings.Repeat(" ", keyWidth)
		inner := alignKey + alignValue + "  "
		// First and last line shared with bracket:
		for i, kv := range l {
			if i > 0 {
				w.WriteString(",\n")
				w.WriteString(alignKey)
			}
			w.WriteString(kv.key)
			w.WriteString(": ")
			w.WriteString(alignValue[len(kv.key):])
			kv.val.WriteTo(w, inner, wout)
		}
	}

	w.WriteByte('}')
}

type list []node

func (l list) WriteTo(w *bytes.Buffer, indent string, wout *writeout) {
	if max := wout.ShortList; max > 0 {
		short := compactString(l)
		if len(short) <= max {
			w.WriteString(short)
			return
		}
	}

	w.WriteByte('[')

	switch {
	case wout.Compact:
		// All on one line:
		for i, v := range l {
			if i > 0 {
				w.WriteByte(',')
			}
			v.WriteTo(w, indent, wout)
		}
	case wout.Diffable:
		w.WriteByte('\n')
		inner := indent + " "
		// Each value gets its own line:
		for _, v := range l {
			w.WriteString(inner)
			v.WriteTo(w, inner, wout)
			w.WriteString(",\n")
		}
		w.WriteString(indent)
	default:
		inner := indent + " "
		// First and last line shared with bracket:
		for i, v := range l {
			if i > 0 {
				w.WriteString(",\n")
				w.WriteString(inner)
			}
			v.WriteTo(w, inner, wout)
		}
	}

	w.WriteByte(']')
}

// label names a node with a label.
// a node can be be referenced by label via ref.
type label struct {
	id   int64 // label id; 0 until labels are resolved
	node node  // node this label names
}

func (l *label) WriteTo(w *bytes.Buffer, indent string, wout *writeout) {
	// allocate label id.
	// since nodes are written out sequentially in order this natuarally
	// assigns sequential ids to labeled nodes.
	wout.curLabelID++
	l.id = wout.curLabelID

	// start labels from new line
	if w.Len() != 0 {
		w.WriteByte('\n')
	}

	w.WriteByte('<')
	w.WriteString(strconv.FormatInt(l.id, 10))
	w.WriteString(">:\n")
	w.WriteString(indent)
	l.node.WriteTo(w, indent, wout)
}

// ref references a label
type refto struct {
	label *label
}

func (r *refto) WriteTo(w *bytes.Buffer, indent string, wout *writeout) {
	w.WriteString("-> <")
	w.WriteString(strconv.FormatInt(r.label.id, 10))
	w.WriteByte('>')
}
