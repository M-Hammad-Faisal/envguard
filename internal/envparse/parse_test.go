package envparse_test

import (
	"testing"

	"github.com/m-hammad-faisal/envguard/internal/envparse"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
		wantKey string
		wantVal string
	}{
		{
			name:    "basic key=value",
			input:   "DB_URL=postgres://localhost/db",
			wantLen: 1,
			wantKey: "DB_URL",
			wantVal: "postgres://localhost/db",
		},
		{
			name:    "double quoted value",
			input:   `SECRET="my secret value"`,
			wantLen: 1,
			wantKey: "SECRET",
			wantVal: "my secret value",
		},
		{
			name:    "single quoted value",
			input:   `SECRET='my secret value'`,
			wantLen: 1,
			wantKey: "SECRET",
			wantVal: "my secret value",
		},
		{
			name:    "comment line skipped",
			input:   "# this is a comment\nKEY=value12345",
			wantLen: 1,
			wantKey: "KEY",
			wantVal: "value12345",
		},
		{
			name:    "empty value skipped",
			input:   "KEY=",
			wantLen: 0,
		},
		{
			name:    "inline comment stripped",
			input:   "KEY=secretvalue123 # this is a comment",
			wantLen: 1,
			wantKey: "KEY",
			wantVal: "secretvalue123",
		},
		{
			name:    "empty line skipped",
			input:   "\n\nKEY=value12345\n\n",
			wantLen: 1,
			wantKey: "KEY",
			wantVal: "value12345",
		},
		{
			name:    "value with equals sign (base64-like)",
			input:   "KEY=base64+value==",
			wantLen: 1,
			wantKey: "KEY",
			wantVal: "base64+value==",
		},
		{
			name:    "whitespace around key and value",
			input:   "  KEY  =  somevalue  ",
			wantLen: 1,
			wantKey: "KEY",
			wantVal: "somevalue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := envparse.Parse(tt.input)
			if len(entries) != tt.wantLen {
				t.Fatalf("got %d entries, want %d (input: %q)", len(entries), tt.wantLen, tt.input)
			}
			if tt.wantLen > 0 {
				if entries[0].Key != tt.wantKey {
					t.Errorf("key: got %q, want %q", entries[0].Key, tt.wantKey)
				}
				if entries[0].Value != tt.wantVal {
					t.Errorf("value: got %q, want %q", entries[0].Value, tt.wantVal)
				}
			}
		})
	}
}

func TestBuildReverseMap(t *testing.T) {
	entries := []envparse.Entry{
		{Key: "SHORT", Value: "abc"},           // len 3 — below minLength
		{Key: "STRIPE", Value: "sk_live_abc"},  // len 11 — included
		{Key: "PORT", Value: "8080"},            // len 4 — below minLength
		{Key: "LONG_KEY", Value: "exactly8c"},  // len 9 — included
	}

	m := envparse.BuildReverseMap(entries, 8)

	if _, ok := m["abc"]; ok {
		t.Error("short value 'abc' should not appear in reverse map")
	}
	if _, ok := m["8080"]; ok {
		t.Error("short value '8080' should not appear in reverse map")
	}
	if key, ok := m["sk_live_abc"]; !ok {
		t.Error("expected 'sk_live_abc' in reverse map")
	} else if key != "STRIPE" {
		t.Errorf("expected key STRIPE, got %q", key)
	}
	if _, ok := m["exactly8c"]; !ok {
		t.Error("expected 'exactly8c' (len 9) in reverse map")
	}
}

func TestBuildReverseMapEmpty(t *testing.T) {
	m := envparse.BuildReverseMap(nil, 8)
	if len(m) != 0 {
		t.Errorf("expected empty map from nil entries, got len %d", len(m))
	}
}

func TestParseMultipleEntries(t *testing.T) {
	input := `
# Database config
DB_URL=postgres://user:password@host/db
API_KEY=sk_live_abc123xyz
PORT=3000
SECRET="my quoted secret value"
`
	entries := envparse.Parse(input)
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}
}
