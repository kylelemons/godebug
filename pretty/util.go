package pretty

import (
	"fmt"

	"github.com/kylelemons/godebug/diff"
)

// Print writes the default representation of the given values to standard output.
func Print(vals ...interface{}) {
	DefaultConfig.Print(vals...)
}

// Print writes the configured presentation of the given values to standard output.
func (cfg *Config) Print(vals ...interface{}) {
	for _, val := range vals {
		fmt.Println(Reflect(val, cfg))
	}
}

// Compare returns a string containing a line-by-line unified diff of the
// values in got and want.  Each line is prefixed with '+', '-', or ' ' to
// indicate if it should be added to, removed from, or is correct for the "got"
// value with respect to the "want" value.
func Compare(got, want interface{}) string {
	diffOpt := &Config{Diffable: true}

	return diff.Diff(Reflect(got, diffOpt), Reflect(want, diffOpt))
}
