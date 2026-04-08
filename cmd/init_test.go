package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestInstallHookFreshRepo(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)
	os.MkdirAll(".git/hooks", 0755)

	if err := installHook(); err != nil {
		t.Fatalf("installHook failed: %v", err)
	}

	content := readFile(t, ".git/hooks/pre-commit")
	if !strings.HasPrefix(content, "#!/bin/sh") {
		t.Error("hook is missing shebang line")
	}
	if !strings.Contains(content, "envguard scan") {
		t.Error("hook does not contain 'envguard scan'")
	}
}

func TestInstallHookNoDuplicates(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)
	os.MkdirAll(".git/hooks", 0755)

	installHook()
	installHook() // second call must not duplicate

	content := readFile(t, ".git/hooks/pre-commit")
	count := strings.Count(content, "envguard scan")
	if count != 1 {
		t.Errorf("expected 1 occurrence of 'envguard scan', got %d", count)
	}
}

func TestInstallHookAppendsToExisting(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)
	os.MkdirAll(".git/hooks", 0755)

	existing := "#!/bin/sh\nnpm run lint\n"
	os.WriteFile(".git/hooks/pre-commit", []byte(existing), 0755)

	if err := installHook(); err != nil {
		t.Fatalf("installHook failed: %v", err)
	}

	content := readFile(t, ".git/hooks/pre-commit")
	if !strings.Contains(content, "npm run lint") {
		t.Error("existing hook content was overwritten")
	}
	if !strings.Contains(content, "envguard scan") {
		t.Error("envguard scan was not appended")
	}
}

func TestDetectFrameworkGo(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)
	os.WriteFile("go.mod", []byte("module example.com/test\ngo 1.21\n"), 0644)

	template := detectFramework()
	if template != `os.Getenv("{{KEY}}")` {
		t.Errorf(`expected os.Getenv("{{KEY}}"), got %q`, template)
	}
}

func TestDetectFrameworkNode(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)
	os.WriteFile("package.json", []byte(`{"dependencies": {}}`), 0644)

	template := detectFramework()
	if template != `process.env.{{KEY}}` {
		t.Errorf("expected process.env.{{KEY}}, got %q", template)
	}
}

func TestDetectFrameworkNextJs(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)
	os.WriteFile("package.json", []byte(`{"dependencies": {"next": "14.0.0"}}`), 0644)

	template := detectFramework()
	if template != `process.env.{{KEY}}` {
		t.Errorf("expected process.env.{{KEY}}, got %q", template)
	}
}

func TestEnsureGitignoreCreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	if err := ensureGitignore(); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, ".gitignore")
	if !strings.Contains(content, ".env") {
		t.Error(".env not added to .gitignore")
	}
}

func TestEnsureGitignoreNoDuplicates(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	ensureGitignore()
	ensureGitignore()

	content := readFile(t, ".gitignore")
	if strings.Count(content, ".env") > 1 {
		t.Error(".env added to .gitignore more than once")
	}
}

func TestEnsureGitignorePreservesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	os.WriteFile(".gitignore", []byte("node_modules/\ndist/\n"), 0644)

	if err := ensureGitignore(); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, ".gitignore")
	if !strings.Contains(content, "node_modules/") {
		t.Error("existing .gitignore content was overwritten")
	}
	if !strings.Contains(content, ".env") {
		t.Error(".env not appended to existing .gitignore")
	}
}

// helpers shared across cmd tests

func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("readFile(%q): %v", path, err)
	}
	return string(data)
}
