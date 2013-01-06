package pretty

import (
	"testing"
)

func TestDiff(t *testing.T) {
	type example struct {
		name    string
		age     int
		friends []string
	}

	tests := []struct {
		desc      string
		got, want interface{}
		diff      string
	}{
		{
			desc: "basic struct",
			got: example{
				name: "Zaphd",
				age:  42,
				friends: []string{
					"Ford Prefect",
					"Trillian",
					"Marvin",
				},
			},
			want: example{
				name: "Zaphod",
				age:  42,
				friends: []string{
					"Ford Prefect",
					"Trillian",
				},
			},
			diff: ` {
- name:    "Zaphd",
+ name:    "Zaphod",
  age:     42,
  friends: [
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
