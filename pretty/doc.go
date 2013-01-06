// Package pretty pretty-prints Go structures.
//
// This package uses reflection to examine a Go value and can
// print out in a nice, aligned fashion.  It supports three
// modes (normal, compact, and extended) for advanced use.
//
// See the Reflect and Print examples for what the output looks like.
//
// TODO:
//   - Catch cycles
//   - Logging helpers
package pretty
