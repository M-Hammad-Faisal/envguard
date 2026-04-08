//go:build windows

package cmd

import (
	"bufio"
	"os"
)

func openTTY() (*os.File, error) {
	return os.Stdin, nil
}

func readLineFromTTY(tty *os.File) (string, error) {
	scanner := bufio.NewScanner(tty)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	return "", scanner.Err()
}
