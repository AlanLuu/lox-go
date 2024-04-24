package ast

import (
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxInstance struct {
	class  LoxClass
	fields map[string]any
}

func NewLoxInstance(class LoxClass) *LoxInstance {
	return &LoxInstance{
		class:  class,
		fields: make(map[string]any),
	}
}

func (i *LoxInstance) Get(name token.Token) (any, error) {
	value, ok := i.fields[name.Lexeme]
	if ok {
		return value, nil
	}
	return nil, loxerror.RuntimeError(name, "Undefined property '"+name.Lexeme+"'.")
}

func (i *LoxInstance) Set(name token.Token, value any) {
	i.fields[name.Lexeme] = value
}

func (i *LoxInstance) String() string {
	return i.class.name + " instance"
}
