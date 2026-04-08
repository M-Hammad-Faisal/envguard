package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestIsBinaryFile(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantBin bool
	}{
		{"plain text", []byte("hello world\n"), false},
		{"null byte", []byte("hello\x00world"), true},
		{"invalid utf8", []byte{0xFF, 0xFE, 0x00}, true},
		{"source code with special chars", []byte(`const url = "postgres://user:p@ssw0rd!/db"`), false},
		{"empty", []byte{}, false},
		{"high ascii no null", []byte{0xC3, 0xA9}, false}, // valid UTF-8 (é)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBinaryFile(tt.input); got != tt.wantBin {
				t.Errorf("isBinaryFile(%q) = %v, want %v", tt.input, got, tt.wantBin)
			}
		})
	}
}

func TestScanFileSkipsEnvguardDir(t *testing.T) {
	reverseMap := map[string]string{"sk_live_abc123xyz": "STRIPE_KEY"}
	matches, err := scanFile(".envguard/secrets.enc", reverseMap, `process.env.{{KEY}}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches for .envguard/ file, got %d", len(matches))
	}
}

func TestScanFileSkipsDeletedFile(t *testing.T) {
	reverseMap := map[string]string{"sk_live_abc123xyz": "STRIPE_KEY"}
	matches, err := scanFile("/tmp/envguard-nonexistent-file-xyz987.js", reverseMap, `process.env.{{KEY}}`)
	if err != nil {
		t.Fatalf("unexpected error on missing file: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches for deleted file, got %d", len(matches))
	}
}

func TestScanFileDetectsSecret(t *testing.T) {
	content := "const stripe = new Stripe(\"sk_live_abc123xyz\")\n"
	path := writeTemp(t, content)

	reverseMap := map[string]string{"sk_live_abc123xyz": "STRIPE_KEY"}
	matches, err := scanFile(path, reverseMap, `process.env.{{KEY}}`)
	if err != nil {
		t.Fatalf("scanFile error: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	m := matches[0]
	if m.secretKey != "STRIPE_KEY" {
		t.Errorf("secretKey: got %q, want STRIPE_KEY", m.secretKey)
	}
	if m.lineNum != 1 {
		t.Errorf("lineNum: got %d, want 1", m.lineNum)
	}
	if !strings.Contains(m.replacement, "process.env.STRIPE_KEY") {
		t.Errorf("replacement missing expected expression: %q", m.replacement)
	}
}

func TestScanFileDetectsSecretOnCorrectLine(t *testing.T) {
	content := "line one\nline two\nconst key = \"sk_live_abc123xyz\"\nline four\n"
	path := writeTemp(t, content)

	reverseMap := map[string]string{"sk_live_abc123xyz": "STRIPE_KEY"}
	matches, err := scanFile(path, reverseMap, `process.env.{{KEY}}`)
	if err != nil {
		t.Fatalf("scanFile error: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].lineNum != 3 {
		t.Errorf("expected lineNum 3, got %d", matches[0].lineNum)
	}
}

func TestScanFileNoFalsePositives(t *testing.T) {
	// Value is shorter than minLength=8 so it won't appear in reverseMap
	content := "const port = \"3000\"\n"
	path := writeTemp(t, content)

	reverseMap := map[string]string{} // empty — short values filtered out upstream
	matches, err := scanFile(path, reverseMap, `process.env.{{KEY}}`)
	if err != nil {
		t.Fatalf("scanFile error: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

func TestScanFileSkipsBinaryFile(t *testing.T) {
	content := []byte("real source\x00binary\xFF\xFE garbage")
	path := writeRawTemp(t, content)

	reverseMap := map[string]string{"sk_live_abc123xyz": "STRIPE_KEY"}
	matches, err := scanFile(path, reverseMap, `process.env.{{KEY}}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches for binary file, got %d", len(matches))
	}
}

func TestApplyFixMutatesCorrectLine(t *testing.T) {
	content := "line one\nconst key = \"sk_live_abc123xyz\"\nline three\n"
	path := writeTemp(t, content)

	m := match{
		file:        path,
		lineNum:     2,
		original:    `const key = "sk_live_abc123xyz"`,
		replacement: `const key = process.env.STRIPE_KEY`,
		secretKey:   "STRIPE_KEY",
	}

	if err := applyFix(&m); err != nil {
		t.Fatalf("applyFix failed: %v", err)
	}

	result := readFile(t, path)
	if !strings.Contains(result, "process.env.STRIPE_KEY") {
		t.Errorf("fix was not applied: %q", result)
	}
	if !strings.Contains(result, "line one") {
		t.Error("applyFix corrupted surrounding lines")
	}
	if !strings.Contains(result, "line three") {
		t.Error("applyFix corrupted surrounding lines")
	}
}

func TestApplyFixRejectsChangedLine(t *testing.T) {
	// Simulate the file changing between scan and fix
	content := "const key = \"sk_live_DIFFERENT\"\n"
	path := writeTemp(t, content)

	m := match{
		file:        path,
		lineNum:     1,
		original:    `const key = "sk_live_abc123xyz"`, // doesn't match file content
		replacement: `const key = process.env.STRIPE_KEY`,
		secretKey:   "STRIPE_KEY",
	}

	err := applyFix(&m)
	if err == nil {
		t.Fatal("expected error when line has changed since scan, got nil")
	}
}

func TestApplyFixOutOfRange(t *testing.T) {
	content := "only one line\n"
	path := writeTemp(t, content)

	m := match{
		file:        path,
		lineNum:     99,
		original:    "only one line",
		replacement: "replaced",
		secretKey:   "KEY",
	}

	err := applyFix(&m)
	if err == nil {
		t.Fatal("expected error on out-of-range line number")
	}
}

// helpers

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	return writeRawTemp(t, []byte(content))
}

func writeRawTemp(t *testing.T, content []byte) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "envguard-test-*.js")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}
