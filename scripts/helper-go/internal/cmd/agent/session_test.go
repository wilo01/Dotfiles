package agent

import "testing"

func TestEnvironmentUpNeedsBothHalves(t *testing.T) {
	cases := map[string]bool{
		"hexer=running db=running": true,
		"hexer=running db=stopped": false,
		"hexer=running db=absent":  false,
		"hexer=stopped db=running": false,
		"hexer=stopped db=stopped": false,
		"":                         false,
	}

	for status, want := range cases {
		if got := environmentUp(status); got != want {
			t.Errorf("environmentUp(%q) = %v, want %v", status, got, want)
		}
	}
}
