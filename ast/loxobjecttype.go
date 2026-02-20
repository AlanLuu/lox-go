package ast

import (
	"fmt"

	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxObjectType struct {
	name       string
	properties map[string]any
}

func NewLoxObjectType(properties map[string]any) *LoxObjectType {
	return &LoxObjectType{
		name:       "",
		properties: properties,
	}
}

func (l *LoxObjectType) setName(name string) *LoxObjectType {
	l.name = name
	return l
}

func (l *LoxObjectType) Get(name *token.Token) (any, error) {
	lexemeName := name.Lexeme
	if field, ok := l.properties[lexemeName]; ok {
		return field, nil
	}
	if l.name != "" {
		return nil, loxerror.RuntimeError(
			name,
			fmt.Sprintf(
				"This built-in %v object has no property called '%v'.",
				l.name,
				lexemeName,
			),
		)
	}
	return nil, loxerror.RuntimeError(
		name,
		"This built-in object has no property called '"+lexemeName+"'.",
	)
}

func (l *LoxObjectType) String() string {
	if l.name == "" {
		return fmt.Sprintf("<built-in object at %p>", l)
	}
	return fmt.Sprintf("<built-in %v object at %p>", l.name, l)
}

func (l *LoxObjectType) Type() string {
	return "built in object"
}
