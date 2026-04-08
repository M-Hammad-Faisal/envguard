//go:build windows

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/term"
)

func openTTY() (*os.File, error) {
	return os.Stdin, nil
}

func promptPassphraseViaTTY(prompt string) (string, error) {
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

func readLineFromTTY(tty *os.File) (string, error) {
	scanner := bufio.NewScanner(tty)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	return "", scanner.Err()
}
