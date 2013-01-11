package pretty

import (
	"testing"
)

func TestDiff(t *testing.T) {
	type example struct {
		Name    string
		Age     int
		Friends []string
	}

	tests := []struct {
		desc      string
		got, want interface{}
		diff      string
	}{
		{
			desc: "basic struct",
			got: example{
				Name: "Zaphd",
				Age:  42,
				Friends: []string{
					"Ford Prefect",
					"Trillian",
					"Marvin",
				},
			},
			want: example{
				Name: "Zaphod",
				Age:  42,
				Friends: []string{
					"Ford Prefect",
					"Trillian",
				},
			},
			diff: ` {
- Name:    "Zaphd",
+ Name:    "Zaphod",
  Age:     42,
  Friends: [
            "Ford Prefect",
            "Trillian",
-           "Marvin",
           ],
 }`,
		},
	}

	for _, test := range tests {
		if got, want := Compare(test.got, test.want), test.diff; got != want {
			t.Errorf("%s:", test.desc)
			t.Errorf("  got:  %q", got)
			t.Errorf("  want: %q", want)
		}
	}
}
