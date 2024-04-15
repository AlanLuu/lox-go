package loxerror

import (
	"errors"
	"fmt"
	"os"
)

func GiveError(line int, where string, message string) error {
	errorMsg := fmt.Sprintf("[line %v] Error%v: %v", line, where, message)
	return errors.New(errorMsg)
}

func PrintErrorObject(e error) {
	fmt.Fprintf(os.Stderr, "%v\n", e.Error())
}

func PrintError(line int, where string, message string) {
	e := GiveError(line, where, message)
	PrintErrorObject(e)
}
