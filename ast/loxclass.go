package ast

import "github.com/AlanLuu/lox/list"

type LoxClass struct {
	name    string
	methods map[string]LoxFunction
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

func (c LoxClass) findMethod(name string) (LoxFunction, bool) {
	value, ok := c.methods[name]
	return value, ok
}

func (c LoxClass) String() string {
	return c.name
}
