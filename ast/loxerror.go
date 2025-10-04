package ast

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxError struct {
	theError   error
	properties map[string]any
}

func NewLoxError(theError error) *LoxError {
	return &LoxError{
		theError:   theError,
		properties: make(map[string]any),
	}
}

func (l *LoxError) isNil() bool {
	return l.theError == nil
}

func (l *LoxError) Get(name *token.Token) (any, error) {
	propertyName := name.Lexeme
	if property, ok := l.properties[propertyName]; ok {
		return property, nil
	}
	errorProperty := func(property any) (any, error) {
		if _, ok := l.properties[propertyName]; !ok {
			l.properties[propertyName] = property
		}
		return property, nil
	}
	switch propertyName {
	case "isNil":
		return errorProperty(l.isNil())
	case "message":
		if l.isNil() {
			return errorProperty(EmptyLoxString())
		}
		return errorProperty(NewLoxString(l.theError.Error(), '\''))
	case "messageOnly":
		if l.isNil() {
			return errorProperty(EmptyLoxString())
		}
		errStr := l.theError.Error()
		errStrLower := strings.ToLower(errStr)
		index := strings.LastIndex(errStrLower, "\n[line ")
		if index > 0 {
			index--
			inLoop := false
			for index > 0 && unicode.IsSpace(rune(errStrLower[index])) {
				inLoop = true
				index--
			}
			if !inLoop {
				index++
			}
			return errorProperty(NewLoxString(errStr[:index], '\''))
		}
		index2 := strings.LastIndex(errStrLower, "[line ")
		if index2 == 0 {
			index3 := strings.LastIndexByte(errStrLower, ']')
			if index3 > 0 {
				index3++
				if index3 < len(errStrLower) {
					for index3 < len(errStrLower) && unicode.IsSpace(rune(errStrLower[index3])) {
						index3++
					}
					return errorProperty(NewLoxString(errStr[:index3], '\''))
				}
			}
		} else if index2 > 0 {
			index2--
			inLoop := false
			for index2 > 0 && unicode.IsSpace(rune(errStrLower[index2])) {
				inLoop = true
				index2--
			}
			if !inLoop {
				index2++
			}
			return errorProperty(NewLoxString(errStr[:index2], '\''))
		}
		return errorProperty(NewLoxString(errStr, '\''))
	}
	return nil, loxerror.RuntimeError(name, "Error objects have no property called '"+propertyName+"'.")
}

func (l *LoxError) String() string {
	if l.isNil() {
		return fmt.Sprintf("<nil error object at %p>", l)
	}
	return fmt.Sprintf("<error object at %p>", l)
}

func (l *LoxError) Type() string {
	return "error"
}
