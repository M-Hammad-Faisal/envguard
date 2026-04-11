//go:build !windows

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/term"
)

// openTTY opens /dev/tty directly so interactive prompts work even when
// stdin has been redirected — which is always the case inside git hooks.
func openTTY() (*os.File, error) {
	return os.OpenFile("/dev/tty", os.O_RDWR, 0)
}

func readLineFromTTY(tty *os.File) (string, error) {
	scanner := bufio.NewScanner(tty)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", nil
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
