package cli

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// readPassword читает пароль с терминала без эха.
func readPassword(errOut io.Writer, prompt string) (string, error) {
	fmt.Fprint(errOut, prompt)

	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		b, err := term.ReadPassword(fd)
		fmt.Fprintln(errOut)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}

	var s string
	if _, err := fmt.Fscanln(os.Stdin, &s); err != nil {
		return "", err
	}
	return s, nil
}
