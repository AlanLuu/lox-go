package ast

import (
	"fmt"

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

func (l *LoxError) Get(name token.Token) (any, error) {
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
	case "message":
		return errorProperty(NewLoxString(l.theError.Error(), '\''))
	}
	return nil, loxerror.RuntimeError(name, "Error objects have no property called '"+propertyName+"'.")
}

func (l *LoxError) String() string {
	return fmt.Sprintf("<error object at %p>", l)
}

func (l *LoxError) Type() string {
	return "error"
}
