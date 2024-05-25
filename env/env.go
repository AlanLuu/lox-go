package env

import (
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type Environment struct {
	values    map[string]any
	enclosing *Environment
}

func NewEnvironment() *Environment {
	return &Environment{
		values:    make(map[string]any),
		enclosing: nil,
	}
}

func NewEnvironmentEnclosing(enclosing *Environment) *Environment {
	return &Environment{
		values:    make(map[string]any),
		enclosing: enclosing,
	}
}

func (e *Environment) ancestor(distance int) *Environment {
	environment := e
	for i := 0; i < distance; i++ {
		environment = environment.enclosing
	}
	return environment
}

func (e *Environment) Assign(name token.Token, value any) error {
	for tempE := e; tempE != nil; tempE = tempE.enclosing {
		_, ok := tempE.values[name.Lexeme]
		if ok {
			tempE.values[name.Lexeme] = value
			return nil
		}
	}
	return loxerror.RuntimeError(name, "undefined variable '"+name.Lexeme+"'.")
}

func (e *Environment) AssignAt(distance int, name token.Token, value any) {
	e.ancestor(distance).Assign(name, value)
}

func (e *Environment) Define(name string, value any) {
	e.values[name] = value
}

func (e *Environment) Get(name token.Token) (any, error) {
	for tempE := e; tempE != nil; tempE = tempE.enclosing {
		value, ok := tempE.values[name.Lexeme]
		if ok {
			return value, nil
		}
	}
	return nil, loxerror.RuntimeError(name, "undefined variable '"+name.Lexeme+"'.")
}

func (e *Environment) GetAt(distance int, name token.Token) (any, error) {
	value, ok := e.ancestor(distance).values[name.Lexeme]
	if ok {
		return value, nil
	}
	return nil, loxerror.RuntimeError(name, "undefined variable '"+name.Lexeme+"'.")
}

func (e *Environment) GetAtStr(distance int, name string) any {
	return e.ancestor(distance).values[name]
}

func (e *Environment) Values() map[string]any {
	return e.values
}
