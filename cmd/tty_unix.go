//go:build !windows

package cmd

import (
	"bufio"
	"fmt"
	"os"

	"golang.org/x/term"
)

// openTTY opens /dev/tty directly so interactive prompts work even when
// stdin has been redirected — which is always the case inside git hooks.
func openTTY() (*os.File, error) {
	return os.OpenFile("/dev/tty", os.O_RDWR, 0)
}

func promptPassphraseViaTTY(prompt string) (string, error) {
	tty, err := openTTY()
	if err != nil {
		return "", fmt.Errorf("cannot open terminal for secure input: %w", err)
	}
	defer tty.Close()

	fmt.Fprint(tty, prompt)
	passBytes, err := term.ReadPassword(int(tty.Fd()))
	fmt.Fprintln(tty)
	if err != nil {
		return "", err
	}
	if len(passBytes) == 0 {
		return "", fmt.Errorf("passphrase cannot be empty")
	}
	return string(passBytes), nil
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
