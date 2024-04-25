package ast

import "github.com/AlanLuu/lox/list"

type LoxClass struct {
	name    string
	methods map[string]LoxFunction
}

func (c LoxClass) arity() int {
	return 0
}

func (c LoxClass) call(interpreter *Interpreter, arguments list.List[any]) (any, error) {
	return NewLoxInstance(c), nil
}

func (c LoxClass) findMethod(name string) (LoxFunction, bool) {
	value, ok := c.methods[name]
	return value, ok
}

func (c LoxClass) String() string {
	return c.name
}
