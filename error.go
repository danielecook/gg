package main

import (
	"fmt"
	"os"

	. "github.com/logrusorgru/aurora"
)

var errorCode = map[string]int{
	"JsonParse":           2,
	"LibVersionMalformed": 3,
	"LibWriteError":       4,
	"MissingApiKey":       11,
	"UnknownError":        -1,
	"ServerError":         500,
}

var errorMsg = map[string]string{
	"JsonParse":           "Error parsing JSON",
	"LibVersionMalformed": "Libversion malformed",
	"LibWriteError":       "Unable to write library",
	"MissingApiKey":       fmt.Sprintf("%s", Bold("\n\tMissing API Key; Run 'sq login'\n\n")),
	"UnknownError":        "Unknown Error",
	"ServerError":         "Cannot connect to server",
}

// ThrowError - Formats an error for CLI
func ThrowError(errString string) {
	for k := range errorMsg {
		if errString == k {
			fmt.Fprint(os.Stderr, fmt.Sprintf("\n\t%s\n\n", Bold(Red(errorMsg[errString]))))
			os.Exit(errorCode[errString])
		}
	}
	fmt.Fprintf(os.Stderr, fmt.Sprintf("\n\t%s\n\n", Bold(Red(errString))))
	os.Exit(-1)
}
