package main

import (
	"fmt"
	"os"

	. "github.com/logrusorgru/aurora"
)

// ThrowError - Formats an error for CLI
func ThrowError(errString string, exitCode int) {
	fmt.Fprint(os.Stderr, fmt.Sprintf("\n\t%s\n\n", Bold(Red(errString))))
	os.Exit(exitCode)
}
