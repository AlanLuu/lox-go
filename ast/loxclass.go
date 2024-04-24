package ast

import "github.com/AlanLuu/lox/list"

type LoxClass struct {
	name string
}

func (c LoxClass) arity() int {
	return 0
}

func (c LoxClass) call(interpreter *Interpreter, arguments list.List[any]) (any, error) {
	return NewLoxInstance(c), nil
}

func (c LoxClass) String() string {
	return c.name
}
