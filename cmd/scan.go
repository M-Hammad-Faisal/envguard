package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"github.com/m-hammad-faisal/envguard/internal/envparse"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan staged files for hardcoded secrets (invoked automatically by pre-commit hook)",
	RunE:  runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)
}

type match struct {
	file        string
	lineNum     int    // 1-indexed
	original    string // exact original line content
	replacement string // proposed replacement line
	secretKey   string // the .env KEY whose value was found
}

func runScan(cmd *cobra.Command, args []string) error {
	stagedFiles, err := getStagedFiles()
	if err != nil {
		return fmt.Errorf("failed to get staged files: %w", err)
	}
	if len(stagedFiles) == 0 {
		return nil
	}

	envContent, err := os.ReadFile(".env")
	if err != nil {
		// No .env present — nothing to scan against. Not an error.
		return nil
	}

	entries := envparse.Parse(string(envContent))
	reverseMap := envparse.BuildReverseMap(entries, 8)
	if len(reverseMap) == 0 {
		return nil
	}

	template, err := loadTemplate()
	if err != nil {
		template = `ENV["{{KEY}}"]`
		fmt.Fprintf(os.Stderr, "Warning: could not load .envguard/config.json — using fallback template\n")
	}

	var matches []match
	for _, file := range stagedFiles {
		fileMatches, err := scanFile(file, reverseMap, template)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not scan %s: %v\n", file, err)
			continue
		}
		matches = append(matches, fileMatches...)
	}

	if len(matches) == 0 {
		fmt.Println("✓ No hardcoded secrets detected.")
		return nil
	}

	// Open /dev/tty directly — stdin is always redirected inside a git hook.
	tty, err := openTTY()
	if err != nil {
		fmt.Fprintln(os.Stderr, "✗ Cannot open terminal for interactive prompt. Aborting commit for safety.")
		os.Exit(1)
	}
	defer tty.Close()

	commitBlocked := false
	for i := range matches {
		action := presentAndPrompt(&matches[i], tty)
		switch action {
		case "yes":
			if err := applyFix(&matches[i]); err != nil {
				return fmt.Errorf("failed to apply fix to %s: %w", matches[i].file, err)
			}
			if err := restageFile(matches[i].file); err != nil {
				return fmt.Errorf("failed to re-stage %s: %w", matches[i].file, err)
			}
			fmt.Fprintf(tty, "✓ Fixed and re-staged %s\n", matches[i].file)
		default:
			commitBlocked = true
		}
	}

	if commitBlocked {
		fmt.Fprintln(os.Stderr, "\n✗ Commit aborted. Remove or replace hardcoded secrets and try again.")
		os.Exit(1)
	}

	return nil
}

func getStagedFiles() ([]string, error) {
	out, err := exec.Command("git", "diff", "--cached", "--name-only").Output()
	if err != nil {
		return nil, err
	}
	var files []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

// isBinaryFile returns true if data contains null bytes or is not valid UTF-8.
func isBinaryFile(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return !utf8.Valid(data)
}

func scanFile(filePath string, reverseMap map[string]string, template string) ([]match, error) {
	// Never scan the encrypted store itself
	if strings.HasPrefix(filePath, ".envguard/") {
		return nil, nil
	}

	// File may have been deleted in this commit
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	if isBinaryFile(data) {
		return nil, nil
	}

	var matches []match
	lines := strings.Split(string(data), "\n")
	for lineIdx, line := range lines {
		for secretValue, secretKey := range reverseMap {
			if !strings.Contains(line, secretValue) {
				continue
			}
			replacementExpr := strings.ReplaceAll(template, "{{KEY}}", secretKey)
			newLine := strings.ReplaceAll(line, secretValue, replacementExpr)
			matches = append(matches, match{
				file:        filePath,
				lineNum:     lineIdx + 1,
				original:    line,
				replacement: newLine,
				secretKey:   secretKey,
			})
		}
	}
	return matches, nil
}

func presentAndPrompt(m *match, tty *os.File) string {
	const (
		red   = "\033[31m"
		green = "\033[32m"
		bold  = "\033[1m"
		reset = "\033[0m"
	)

	fmt.Fprintf(tty, "\n%s⚠  Secret detected: %s — line %d%s\n", bold, m.file, m.lineNum, reset)
	fmt.Fprintln(tty, "────────────────────────────────────────────────────────")
	fmt.Fprintf(tty, "%s- %s%s\n", red, m.original, reset)
	fmt.Fprintf(tty, "%s+ %s%s\n", green, m.replacement, reset)
	fmt.Fprintln(tty, "────────────────────────────────────────────────────────")
	fmt.Fprintln(tty, "⚠  Note: applying this fix will stage the entire file (git limitation).")
	fmt.Fprint(tty, "Apply this fix? [Y/n/skip]: ")

	input, _ := readLineFromTTY(tty)
	input = strings.ToLower(strings.TrimSpace(input))

	switch input {
	case "", "y", "yes":
		return "yes"
	default:
		return "no"
	}
}

func applyFix(m *match) error {
	data, err := os.ReadFile(m.file)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	if m.lineNum-1 >= len(lines) {
		return fmt.Errorf("line %d out of range in %s (file changed during scan)", m.lineNum, m.file)
	}
	// Sanity check: if the line changed between scan and fix, abort rather than corrupt
	if lines[m.lineNum-1] != m.original {
		return fmt.Errorf("line %d in %s has changed since scan — skipping fix to avoid corruption", m.lineNum, m.file)
	}
	lines[m.lineNum-1] = m.replacement
	return os.WriteFile(m.file, []byte(strings.Join(lines, "\n")), 0644)
}

func restageFile(filePath string) error {
	return exec.Command("git", "add", filePath).Run()
}

func loadTemplate() (string, error) {
	data, err := os.ReadFile(".envguard/config.json")
	if err != nil {
		return "", err
	}
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return "", err
	}
	if config.ReplacementTemplate == "" {
		return "", fmt.Errorf("replacement_template is empty in config.json")
	}
	return config.ReplacementTemplate, nil
}
