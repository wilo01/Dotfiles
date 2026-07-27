package token

import "testing"

func TestResolveParam(t *testing.T) {
	cases := []struct {
		name    string
		flagVal string
		cfgVal  string
		want    string
	}{
		{"flag wins over config", "from-flag", "from-config", "from-flag"},
		{"config used when flag empty", "", "from-config", "from-config"},
		{"empty when both empty", "", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveParam(tc.flagVal, tc.cfgVal); got != tc.want {
				t.Errorf("resolveParam(%q, %q) = %q, want %q", tc.flagVal, tc.cfgVal, got, tc.want)
			}
		})
	}
}

func TestResolveClientSecret(t *testing.T) {
	cases := []struct {
		name    string
		hlpVar  string
		tdsVar  string
		want    string
		wantErr bool
	}{
		{"HLP_ var wins", "hlp-secret", "tds-secret", "hlp-secret", false},
		{"falls back to TDS_ var", "", "tds-secret", "tds-secret", false},
		{"only HLP_ set", "hlp-secret", "", "hlp-secret", false},
		{"neither set is an error", "", "", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("HLP_TDS_CLIENT_SECRET", tc.hlpVar)
			t.Setenv("TDS_CLIENT_SECRET", tc.tdsVar)

			got, err := resolveClientSecret()
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error when no env var is set")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("resolveClientSecret() = %q, want %q", got, tc.want)
			}
		})
	}
}
