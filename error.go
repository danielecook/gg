package main

import (
	"fmt"
	"os"
)

// ThrowError - Formats an error for CLI
func ThrowError(errString string, exitCode int) {
	errorMsg(fmt.Sprintf("\n\t%s\n\n", errString))
	os.Exit(exitCode)
}
