package ast

import (
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxClass struct {
	name         string
	superClass   *LoxClass
	methods      map[string]*LoxFunction
	classMethods map[string]*LoxFunction
}

func (c LoxClass) arity() int {
	initializer, ok := c.findMethod("init")
	if !ok {
		return 0
	}
	return initializer.arity()
}

func (c LoxClass) call(interpreter *Interpreter, arguments list.List[any]) (any, error) {
	instance := NewLoxInstance(c)
	initializer, ok := c.findMethod("init")
	if ok {
		initializer.bind(instance).call(interpreter, arguments)
	}
	return instance, nil
}

func (c LoxClass) Get(name token.Token) (any, error) {
	method, ok := c.classMethods[name.Lexeme]
	if ok {
		return method.bind(c), nil
	}
	return nil, loxerror.RuntimeError(name, "Undefined property '"+name.Lexeme+"'.")
}

func (c LoxClass) findMethod(name string) (*LoxFunction, bool) {
	value, ok := c.methods[name]
	if ok {
		return value, ok
	}
	if c.superClass != nil {
		return c.superClass.findMethod(name)
	}
	return value, ok
}

func (c LoxClass) String() string {
	return c.name
}
