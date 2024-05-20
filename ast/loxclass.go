package ast

import (
	"fmt"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxClass struct {
	name            string
	superClass      *LoxClass
	methods         map[string]*LoxFunction
	classProperties map[string]any
	instanceFields  map[string]any
	canInstantiate  bool
}

func (c *LoxClass) arity() int {
	if !c.canInstantiate {
		return -1
	}
	initializer, ok := c.findMethod("init")
	if !ok {
		return 0
	}
	return initializer.arity()
}

func (c *LoxClass) call(interpreter *Interpreter, arguments list.List[any]) (any, error) {
	if !c.canInstantiate {
		return nil, loxerror.Error(fmt.Sprintf("Cannot instantiate class '%v'.", c.name))
	}
	instance := NewLoxInstance(c)
	for name, field := range c.instanceFields {
		instance.fields[name] = field
	}
	initializer, ok := c.findMethod("init")
	if ok {
		call, callErr := initializer.bind(instance).call(interpreter, arguments)
		if callErr != nil && call == nil {
			return nil, callErr
		}
	}
	return instance, nil
}

func (c *LoxClass) Get(name token.Token) (any, error) {
	item, ok := c.classProperties[name.Lexeme]
	if ok {
		switch method := item.(type) {
		case *LoxFunction:
			return method.bind(c), nil
		}
		return item, nil
	}
	return nil, loxerror.RuntimeError(name, "Undefined property '"+name.Lexeme+"'.")
}

func (c *LoxClass) findMethod(name string) (*LoxFunction, bool) {
	value, ok := c.methods[name]
	if ok {
		return value, ok
	}
	if c.superClass != nil {
		return c.superClass.findMethod(name)
	}
	return value, ok
}

func (c *LoxClass) String() string {
	return fmt.Sprintf("<class %v at %p>", c.name, c)
}
