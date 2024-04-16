package env

import (
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type Environment struct {
	values map[string]any
}

func NewEnvironment() *Environment {
	return &Environment{
		values: make(map[string]any),
	}
}

func (e *Environment) Assign(name token.Token, value any) error {
	_, ok := e.values[name.Lexeme]
	if ok {
		e.values[name.Lexeme] = value
		return nil
	}
	return loxerror.RuntimeError(name, "undefined variable '"+name.Lexeme+"'.")
}

func (e *Environment) Define(name string, value any) {
	e.values[name] = value
}

func (e *Environment) Get(name token.Token) (any, error) {
	value, ok := e.values[name.Lexeme]
	if !ok {
		return nil, loxerror.RuntimeError(name, "undefined variable '"+name.Lexeme+"'.")
	}
	return value, nil
}
