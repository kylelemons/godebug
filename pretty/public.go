package pretty

import (
	"bytes"
	"fmt"
	"io"
	"reflect"

	"github.com/kylelemons/godebug/diff"
)

func (cfg *Config) fprint(buf *bytes.Buffer, vals ...interface{}) {
	for i, val := range vals {
		if i > 0 {
			buf.WriteByte('\n')
		}
		val2node(reflect.ValueOf(val)).WriteTo(buf, "", cfg)
	}
}

// Print writes the default representation of the given values to standard output.
func Print(vals ...interface{}) {
	DefaultConfig.Print(vals...)
}

// Print writes the configured presentation of the given values to standard output.
func (cfg *Config) Print(vals ...interface{}) {
	fmt.Println(cfg.Sprint(vals...))
}

// Sprint returns a string representation of the given value according to the default config.
func Sprint(vals ...interface{}) string {
	return DefaultConfig.Sprint(vals...)
}

// Sprint returns a string representation of the given value according to cfg.
func (cfg *Config) Sprint(vals ...interface{}) string {
	buf := new(bytes.Buffer)
	cfg.fprint(buf, vals...)
	return buf.String()
}

// Fprint writes the representation of the given value to the writer according to the default config.
func Fprint(w io.Writer, vals ...interface{}) (n int64, err error) {
	return DefaultConfig.Fprint(w, vals...)
}

// Fprint writes the representation of the given value to the writer according to the cfg.
func (cfg *Config) Fprint(w io.Writer, vals ...interface{}) (n int64, err error) {
	buf := new(bytes.Buffer)
	cfg.fprint(buf, vals...)
	return buf.WriteTo(w)
}

// Compare returns a string containing a line-by-line unified diff of the
// values in got and want.  Each line is prefixed with '+', '-', or ' ' to
// indicate if it should be added to, removed from, or is correct for the "got"
// value with respect to the "want" value.
func Compare(got, want interface{}) string {
	diffOpt := &Config{Diffable: true}

	return diff.Diff(diffOpt.Sprint(got), diffOpt.Sprint(want))
}
