package ast

import (
	"fmt"

	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxInstance struct {
	class  *LoxClass
	fields map[string]any
}

func NewLoxInstance(class *LoxClass) *LoxInstance {
	return &LoxInstance{
		class:  class,
		fields: make(map[string]any),
	}
}

func (i *LoxInstance) Get(name *token.Token) (any, error) {
	value, foundValue := i.fields[name.Lexeme]
	if foundValue {
		switch value.(type) {
		case LoxBuiltInProtoCallable:
		default:
			return value, nil
		}
	}
	var method *LoxFunction
	method, ok := i.class.findMethod(name.Lexeme)
	if ok {
		return method.bind(i), nil
	}
	if foundValue {
		return value, nil
	}
	return nil, loxerror.RuntimeError(name, "Undefined property '"+name.Lexeme+"'.")
}

func (i *LoxInstance) Set(name *token.Token, value any) {
	i.fields[name.Lexeme] = value
}

func (i *LoxInstance) String() string {
	return fmt.Sprintf("<%v instance at %p>", i.class.name, i)
}

func (i *LoxInstance) Type() string {
	return i.class.name
}
