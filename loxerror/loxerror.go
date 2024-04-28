package loxerror

import (
	"errors"
	"fmt"
	"os"

	"github.com/AlanLuu/lox/token"
)

func Error(message string) error {
	return errors.New(message)
}

func GiveError(line int, where string, message string) error {
	errorMsg := fmt.Sprintf("[line %v] Error%v: %v", line, where, message)
	return errors.New(errorMsg)
}

func RuntimeError(theToken token.Token, message string) error {
	errorStr := message + "\n[line " + fmt.Sprint(theToken.Line) + "]"
	return errors.New(errorStr)
}

func PrintErrorObject(e error) {
	if len(e.Error()) > 0 {
		fmt.Fprintf(os.Stderr, "%v\n", e.Error())
	}
}

func PrintError(line int, where string, message string) {
	e := GiveError(line, where, message)
	PrintErrorObject(e)
}
