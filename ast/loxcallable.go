package ast

import "github.com/AlanLuu/lox/list"

type LoxCallable interface {
	arity() int
	call(interpreter *Interpreter, arguments list.List[any]) (any, error)
}

type ProtoLoxCallable struct {
	arityMethod  func() int
	callMethod   func(interpreter *Interpreter, arguments list.List[any]) (any, error)
	stringMethod func() string
}

func (l ProtoLoxCallable) arity() int {
	return l.arityMethod()
}

func (l ProtoLoxCallable) call(interpreter *Interpreter, arguments list.List[any]) (any, error) {
	return l.callMethod(interpreter, arguments)
}

func (l ProtoLoxCallable) String() string {
	return l.stringMethod()
}
