//go:build windows

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/term"
)

// openTTY opens CONIN$ — the Windows equivalent of /dev/tty.
// This gives direct console access even when stdin has been redirected,
// which is always the case inside a git hook on Windows.
func openTTY() (*os.File, error) {
	// CONIN$ is the Windows console input device.
	// Unlike os.Stdin, it is not affected by shell redirection or git hook stdin capture.
	tty, err := os.OpenFile("CONIN$", os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot open Windows console (CONIN$): %w", err)
	}
	return tty, nil
}

func readLineFromTTY(tty *os.File) (string, error) {
	scanner := bufio.NewScanner(tty)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	return "", scanner.Err()
}

// promptPassphrase reads a hidden passphrase from stdin.
// Used by push and pull where stdin is always a real TTY.
func promptPassphrase(prompt string) (string, error) {
	fmt.Print(prompt)
	passBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}
	if len(passBytes) == 0 {
		return "", fmt.Errorf("passphrase cannot be empty")
	}
	return string(passBytes), nil
}
