package envparse

import (
	"bufio"
	"strings"
)

// Entry represents a single parsed key=value pair from a .env file.
type Entry struct {
	Key   string
	Value string
}

// Parse reads .env file content and returns all valid entries.
// Handles: comments, empty lines, quoted values, inline comments.
func Parse(content string) []Entry {
	var entries []Entry
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = stripInlineComment(value)
		value = unquote(value)

		if key == "" || value == "" {
			continue
		}
		entries = append(entries, Entry{Key: key, Value: value})
	}
	return entries
}

// BuildReverseMap creates a value→key lookup map for secret scanning.
// Values shorter than minLength are excluded to limit false positives.
// Callers should use minLength >= 8.
func BuildReverseMap(entries []Entry, minLength int) map[string]string {
	m := make(map[string]string, len(entries))
	for _, e := range entries {
		if len(e.Value) >= minLength {
			m[e.Value] = e.Key
		}
	}
	return m
}

func stripInlineComment(s string) string {
	if len(s) == 0 || s[0] == '"' || s[0] == '\'' {
		return s // quoted values: leave alone, comment stripping is handled post-unquote
	}
	if idx := strings.Index(s, " #"); idx != -1 {
		return strings.TrimSpace(s[:idx])
	}
	return s
}

func unquote(s string) string {
	if len(s) < 2 {
		return s
	}
	if (s[0] == '"' && s[len(s)-1] == '"') ||
		(s[0] == '\'' && s[len(s)-1] == '\'') {
		return s[1 : len(s)-1]
	}
	return s
}
