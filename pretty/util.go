package pretty

import (
	"bytes"
	"fmt"
	"strings"
)

// Print writes the default representation of the given values to standard output.
func Print(vals ...interface{}) {
	DefaultOptions.Print(vals...)
}

// Print writes the configured presentation of the given values to standard output.
func (o *Options) Print(vals ...interface{}) {
	for _, val := range vals {
		fmt.Println(Reflect(val, o))
	}
}

// Compare returns a string containing a line-by-line unified diff of the
// values in got and want.  Each line is prefixed with '+', '-', or ' ' to
// indicate if it should be added to, removed from, or is correct for the "got"
// value with respect to the "want" value.
func Compare(got, want interface{}) string {
	diffOpt := &Options{Diffable: true}
	gotLines := strings.Split(Reflect(got, diffOpt), "\n")
	wantLines := strings.Split(Reflect(want, diffOpt), "\n")

	chunks := diff(gotLines, wantLines)
	return diffString(chunks)
}

type chunk struct {
	add []string
	del []string
	eq  []string
}

// Diff uses an O(D(N+M)) shortest-edit-script algorithm
// to compute the difference between A and B.
//
// algorithm: http://www.xmailserver.org/diff2.pdf
func diff(A, B []string) []chunk {
	N, M := len(A), len(B)
	MAX := N + M
	V := make([]int, 2*MAX+1)
	Vs := make([][]int, 0, 8)

	var D int
dLoop:
	for D = 0; D <= MAX; D++ {
		for k := -D; k <= D; k += 2 {
			var x int
			if k == -D || (k != D && V[MAX+k-1] < V[MAX+k+1]) {
				x = V[MAX+k+1]
			} else {
				x = V[MAX+k-1] + 1
			}
			y := x - k
			for x < N && y < M && A[x] == B[y] {
				x++
				y++
			}
			V[MAX+k] = x
			if x >= N && y >= M {
				Vs = append(Vs, append(make([]int, 0, len(V)), V...))
				break dLoop
			}
		}
		Vs = append(Vs, append(make([]int, 0, len(V)), V...))
	}
	fmt.Println(D, N-M)
	for _, V := range Vs {
		fmt.Println(V[:MAX-1], V[MAX:MAX+1], V[MAX+1:])
	}
	if D == 0 {
		return nil
	}
	chunks := make([]chunk, D+1)

	x, y := N, M
	for d := D; d > 0; d-- {
		V := Vs[d]
		k := x - y
		insert := k == -d || (k != d && V[MAX+k-1] < V[MAX+k+1])

		x1 := V[MAX+k]
		var x0, xM, kk int
		if insert {
			kk = k + 1
			x0 = V[MAX+kk]
			xM = x0
		} else {
			kk = k - 1
			x0 = V[MAX+kk]
			xM = x0 + 1
		}
		y0 := x0 - kk

		var c chunk
		if insert {
			c.add = B[y0:][:1]
		} else {
			c.del = A[x0:][:1]
		}
		c.eq = A[xM:][:x1-xM]

		x, y = x0, y0
		chunks[d] = c
	}
	chunks[0].eq = A[:x]
	return chunks
}

func diffString(chunks []chunk) string {
	buf := new(bytes.Buffer)
	for _, c := range chunks {
		for _, line := range c.del {
			fmt.Fprintf(buf, "-%s\n", line)
		}
		for _, line := range c.add {
			fmt.Fprintf(buf, "+%s\n", line)
		}
		for _, line := range c.eq {
			fmt.Fprintf(buf, " %s\n", line)
		}
	}
	return buf.String()
}
