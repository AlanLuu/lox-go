package loxerror

import (
	"errors"
	"fmt"
	"os"

	"github.com/AlanLuu/lox/token"
)

type SyntaxErr struct {
	Line    int
	Where   string
	Message string
	FullMsg string
}

func (l *SyntaxErr) Error() string {
	return l.FullMsg
}

type RuntimeErr struct {
	TheToken *token.Token
	Message  string
	FullMsg  string
}

func (l *RuntimeErr) Error() string {
	return l.FullMsg
}

func Error(message string) error {
	return errors.New(message)
}

func GiveError(line int, where string, message string) error {
	errorMsg := fmt.Sprintf("[line %v] Error%v: %v", line, where, message)
	return &SyntaxErr{
		Line:    line,
		Where:   where,
		Message: message,
		FullMsg: errorMsg,
	}
}

func RuntimeError(theToken *token.Token, message string) error {
	errorStr := message + "\n[line " + fmt.Sprint(theToken.Line) + "]"
	return &RuntimeErr{
		TheToken: theToken,
		Message:  message,
		FullMsg:  errorStr,
	}
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
