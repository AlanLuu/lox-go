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
	case "lineNo":
		const defLineNo = -1
		if l.isNil() {
			return int64(defLineNo), nil
		}
		switch err := l.theError.(type) {
		case *loxerror.SyntaxErr:
			return int64(err.Line), nil
		case *loxerror.RuntimeErr:
			if err.TheToken == nil {
				return int64(defLineNo), nil
			}
			return int64(err.TheToken.Line), nil
		default:
			return int64(defLineNo), nil
		}
	case "message":
		if l.isNil() {
			return errorProperty(EmptyLoxString())
		}
		return errorProperty(NewLoxStringQuote(l.theError.Error()))
	case "messageOnly":
		if l.isNil() {
			return errorProperty(EmptyLoxString())
		}
		switch err := l.theError.(type) {
		case *loxerror.SyntaxErr:
			return errorProperty(NewLoxStringQuote(err.Message))
		case *loxerror.RuntimeErr:
			return errorProperty(NewLoxStringQuote(err.Message))
		default:
			if element, ok := l.properties["message"]; ok {
				return element, nil
			}
			return errorProperty(NewLoxStringQuote(err.Error()))
		}
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
